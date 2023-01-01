package server

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/reddec/jsonrpc2"

	"github.com/reddec/trusted-cgi/api"
	"github.com/reddec/trusted-cgi/api/handlers"
	"github.com/reddec/trusted-cgi/application"
	"github.com/reddec/trusted-cgi/assets"
	"github.com/reddec/trusted-cgi/stats"
	"github.com/reddec/trusted-cgi/types"
)

type TokenHandler interface {
	ValidateToken(ctx context.Context, value *api.Token) error
}

type Server struct {
	Policies     application.Policies
	Platform     application.Platform
	Cases        application.Cases
	Queues       application.Queues
	Dev          bool
	BehindProxy  bool
	Tracker      stats.Recorder
	TokenHandler TokenHandler
	ProjectAPI   api.ProjectAPI
	LambdaAPI    api.LambdaAPI
	UserAPI      api.UserAPI
	QueuesAPI    api.QueuesAPI
	PoliciesAPI  api.PoliciesAPI
}

func (srv *Server) Handler(ctx context.Context) http.Handler {
	mux := http.NewServeMux()
	srv.installAPI(ctx, mux)
	srv.installPublicRoutes(ctx, mux)
	srv.installUI(mux)
	return mux
}

func (srv *Server) installAPI(ctx context.Context, mux *http.ServeMux) {
	var router jsonrpc2.Router
	handlers.RegisterUserAPI(&router, srv.UserAPI, srv.TokenHandler)
	handlers.RegisterLambdaAPI(&router, srv.LambdaAPI, srv.TokenHandler)
	handlers.RegisterProjectAPI(&router, srv.ProjectAPI, srv.TokenHandler)
	handlers.RegisterQueuesAPI(&router, srv.QueuesAPI, srv.TokenHandler)
	handlers.RegisterPoliciesAPI(&router, srv.PoliciesAPI, srv.TokenHandler)

	mux.Handle("/u/", chooseHandler(srv.Dev, jsonrpc2.HandlerRestContext(ctx, &router)))
}

func (srv *Server) installUI(mux *http.ServeMux) {
	mux.Handle("/", http.FileServer(assets.AssetFile()))
}

func (srv *Server) installPublicRoutes(ctx context.Context, mux *http.ServeMux) {
	mux.Handle("/a/", openedHandler(http.StripPrefix("/a/", srv.withRequest(ctx, srv.handleLambda))))
	mux.Handle("/l/", openedHandler(http.StripPrefix("/l/", srv.withRequest(ctx, srv.handleLink))))
	mux.Handle("/q/", openedHandler(http.StripPrefix("/q/", srv.withRequest(ctx, srv.handleQueue))))
}
func (srv *Server) handleQueue(ctx context.Context, req *types.Request, writer http.ResponseWriter, record *stats.Record, uid string) {
	q, err := srv.Queues.Get(uid)
	if err != nil {
		record.Err = err.Error()
		http.Error(writer, err.Error(), http.StatusNotFound)
		return
	}
	err = srv.Policies.Inspect(q.Target, req)

	if err != nil {
		record.Err = err.Error()
		http.Error(writer, err.Error(), http.StatusForbidden)
		return
	}

	err = srv.Queues.Put(uid, req)
	if err != nil {
		record.Err = err.Error()
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}
func (srv *Server) handleLambda(ctx context.Context, req *types.Request, writer http.ResponseWriter, record *stats.Record, uid string) {
	lambda, err := srv.Platform.FindByUID(uid)

	if err != nil {
		record.Err = err.Error()
		http.Error(writer, err.Error(), http.StatusNotFound)
		return
	}

	srv.runLambda(ctx, req, writer, lambda, record)
}

func (srv *Server) handleLink(ctx context.Context, req *types.Request, writer http.ResponseWriter, record *stats.Record, uid string) {
	lambda, err := srv.Platform.FindByLink(uid)

	if err != nil {
		record.Err = err.Error()
		http.Error(writer, err.Error(), http.StatusNotFound)
		return
	}

	srv.runLambda(ctx, req, writer, lambda, record)
}

func (srv *Server) runLambda(ctx context.Context, req *types.Request, writer http.ResponseWriter, lambda *application.Definition, record *stats.Record) {
	err := srv.Policies.Inspect(lambda.UID, req)
	if err != nil {
		record.End = time.Now()
		record.Err = err.Error()
		http.Error(writer, err.Error(), http.StatusForbidden)
		return
	}
	for k, v := range lambda.Lambda.Manifest().OutputHeaders {
		writer.Header().Set(k, v)
	}

	writer.WriteHeader(http.StatusOK)

	err = srv.Platform.Invoke(ctx, lambda.Lambda, *req, writer)
	record.End = time.Now()
	if err != nil {
		record.Err = err.Error()
	}
}

type resourceHandler func(ctx context.Context, req *types.Request, writer http.ResponseWriter, rec *stats.Record, uid string)

func (srv *Server) withRequest(ctx context.Context, next resourceHandler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		sections := strings.SplitN(strings.Trim(request.URL.Path, "/"), "/", 2)
		uid := sections[0]
		req := types.FromHTTP(request, srv.BehindProxy)
		var record = stats.Record{
			UID:     uid,
			Request: *req,
			Begin:   time.Now(),
		}
		next(ctx, req, writer, &record, uid)
		record.End = time.Now()
		srv.Tracker.Track(record)
	})
}

func chooseHandler(dev bool, handler http.Handler) http.Handler {
	if dev {
		return openedHandler(handler)
	} else {
		return securedHttpHandler(handler)
	}
}

func openedHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Access-Control-Allow-Origin", "*")
		writer.Header().Set("Access-Control-Allow-Credentials", "true")
		writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if request.Method == "OPTIONS" {
			writer.WriteHeader(http.StatusNoContent)
			return
		}

		handler.ServeHTTP(writer, request)
	})
}

func securedHttpHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("X-XSS-Protection", "1; mode=block")
		writer.Header().Set("X-Frame-Options", "SAMEORIGIN")
		writer.Header().Set("X-Content-Type-Options", "nosniff")
		handler.ServeHTTP(writer, request)
	})
}
