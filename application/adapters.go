package application

import (
	"context"
	"github.com/reddec/trusted-cgi/stats"
	"github.com/reddec/trusted-cgi/types"
	"net/http"
	"strings"
	"time"
)

// Expose lambda over HTTP handler. First part of path will be used as lambda UID
func HandlerByUID(globalCtx context.Context, tracker stats.Recorder, platform Platform) http.HandlerFunc {
	return handler(globalCtx, platform, tracker, platform.FindByUID)
}

// Expose lambda over HTTP handler. First part of path will be used as lambda alias
func HandlerByLinks(globalCtx context.Context, tracker stats.Recorder, platform Platform) http.HandlerFunc {
	return handler(globalCtx, platform, tracker, platform.FindByLink)
}

// Expose lambda handlers by UID (/a/) and by links (/l/)
func Handler(globalCtx context.Context, tracker stats.Recorder, platform Platform) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/a/", http.StripPrefix("/a/", HandlerByUID(globalCtx, tracker, platform)))
	mux.Handle("/l/", http.StripPrefix("/l/", HandlerByLinks(globalCtx, tracker, platform)))
	return mux
}

func handler(globalCtx context.Context, platform Platform, tracker stats.Recorder, lookup func(string) (*Definition, error)) http.HandlerFunc {
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
