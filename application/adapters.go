package application

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/reddec/trusted-cgi/stats"
	"github.com/reddec/trusted-cgi/types"
)

// Expose lambda over HTTP handler. First part of path will be used as lambda UID
func HandlerByUID(globalCtx context.Context, policies Validator, tracker stats.Recorder, platform Platform) http.HandlerFunc {
	return handler(globalCtx, policies, platform, tracker, platform.FindByUID)
}

// Expose lambda over HTTP handler. First part of path will be used as lambda alias
func HandlerByLinks(globalCtx context.Context, policies Validator, tracker stats.Recorder, platform Platform) http.HandlerFunc {
	return handler(globalCtx, policies, platform, tracker, platform.FindByLink)
}

// Expose queues over HTTP handler. First part of path will be used as queue name
func HandlerByQueues(queues Queues) http.HandlerFunc {
	return handlerQueue(queues)
}

// Expose lambda handlers by UID (/a/) and by links (/l/) and for queues (/q/)
func Handler(globalCtx context.Context, policies Validator, tracker stats.Recorder, platform Platform, queues Queues) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/a/", http.StripPrefix("/a/", HandlerByUID(globalCtx, policies, tracker, platform)))
	mux.Handle("/l/", http.StripPrefix("/l/", HandlerByLinks(globalCtx, policies, tracker, platform)))
	mux.Handle("/q/", http.StripPrefix("/q/", HandlerByQueues(queues)))
	//TODO: queue balancer
	return mux
}

func handler(
	globalCtx context.Context,
	policies Validator,
	platform Platform,
	tracker stats.Recorder,
	lookup func(string) (*Definition, error)) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		defer request.Body.Close()

		sections := strings.SplitN(strings.Trim(request.URL.Path, "/"), "/", 2)
		uid := sections[0]
		req := types.FromHTTP(request)
		var record = stats.Record{
			UID:     uid,
			Request: *req,
			Begin:   time.Now(),
		}

		lambda, err := lookup(uid)

		if err != nil {
			record.End = time.Now()
			record.Err = err.Error()
			tracker.Track(record)
			http.Error(writer, err.Error(), http.StatusNotFound)
			return
		}
		err = policies.Inspect(uid, req)
		if err != nil {
			record.End = time.Now()
			record.Err = err.Error()
			tracker.Track(record)
			http.Error(writer, err.Error(), http.StatusForbidden)
			return
		}
		for k, v := range lambda.Lambda.Manifest().OutputHeaders {
			writer.Header().Set(k, v)
		}

		writer.WriteHeader(http.StatusOK)

		err = platform.Invoke(globalCtx, lambda.Lambda, *req, writer)
		record.End = time.Now()
		if err != nil {
			record.Err = err.Error()
		}
		tracker.Track(record)
	}
}

func handlerQueue(queues Queues) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		defer request.Body.Close()
		sections := strings.SplitN(strings.Trim(request.URL.Path, "/"), "/", 2)
		uid := sections[0]
		req := types.FromHTTP(request)

		err := queues.Put(uid, req)

		if err != nil {
			http.Error(writer, err.Error(), http.StatusNotFound)
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	}
}
