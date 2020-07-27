package server

import (
	"context"
	"net/http"

	"github.com/reddec/jsonrpc2"

	"github.com/reddec/trusted-cgi/api"
	"github.com/reddec/trusted-cgi/api/handlers"
	"github.com/reddec/trusted-cgi/application"
	"github.com/reddec/trusted-cgi/assets"
	"github.com/reddec/trusted-cgi/stats"
)

func Handler(ctx context.Context,
	dev bool,
	policies application.Validator,
	platform application.Platform,
	queues application.Queues,
	tracker stats.Stats,
	tokenHandler interface {
		ValidateToken(ctx context.Context, value *api.Token) error
	},
	projectAPI api.ProjectAPI,
	lambdaAPI api.LambdaAPI,
	userAPI api.UserAPI,
	queuesAPI api.QueuesAPI,
	policiesAPI api.PoliciesAPI,
) (http.Handler, error) {
	var mux http.ServeMux
	// main API
	apps := application.HandlerByUID(ctx, policies, tracker, platform)
	links := application.HandlerByLinks(ctx, policies, tracker, platform)
	queuesHandler := application.HandlerByQueues(queues)

	mux.Handle("/a/", openedHandler(http.StripPrefix("/a/", apps)))
	mux.Handle("/l/", openedHandler(http.StripPrefix("/l/", links)))
	mux.Handle("/q/", openedHandler(http.StripPrefix("/q/", queuesHandler)))

	// admin API
	var router jsonrpc2.Router

	handlers.RegisterUserAPI(&router, userAPI, tokenHandler)
	handlers.RegisterLambdaAPI(&router, lambdaAPI, tokenHandler)
	handlers.RegisterProjectAPI(&router, projectAPI, tokenHandler)
	handlers.RegisterQueuesAPI(&router, queuesAPI, tokenHandler)
	handlers.RegisterPoliciesAPI(&router, policiesAPI, tokenHandler)
	mux.Handle("/u/", chooseHandler(dev, jsonrpc2.HandlerRestContext(ctx, &router)))

	// UI
	mux.Handle("/", http.FileServer(assets.AssetFile()))
	return &mux, nil
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
