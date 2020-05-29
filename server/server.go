package server

import (
	"context"
	"github.com/reddec/jsonrpc2"
	"github.com/reddec/trusted-cgi/api"
	"github.com/reddec/trusted-cgi/api/handlers"
	"github.com/reddec/trusted-cgi/application"
	"github.com/reddec/trusted-cgi/assets"
	"github.com/reddec/trusted-cgi/stats"
	"net/http"
)

func Handler(ctx context.Context,
	dev bool,
	project *application.Project,
	tracker stats.Stats,
	tokenHandler interface {
	ValidateToken(ctx context.Context, value *api.Token) error
},
	projectAPI api.ProjectAPI,
	lambdaAPI api.LambdaAPI,
	userAPI api.UserAPI) (http.Handler, error) {
	apps, err := project.Handler(ctx, tracker)
	if err != nil {
		return nil, err
	}

	links, err := project.HandlerAlias(ctx, tracker)
	if err != nil {
		return nil, err
	}

	var router jsonrpc2.Router

	handlers.RegisterUserAPI(&router, userAPI, tokenHandler)
	handlers.RegisterLambdaAPI(&router, lambdaAPI, tokenHandler)
	handlers.RegisterProjectAPI(&router, projectAPI, tokenHandler)

	var mux http.ServeMux
	mux.Handle("/a/", openedHandler(http.StripPrefix("/a/", apps)))
	mux.Handle("/u/", secureHttpHandler(dev, jsonrpc2.HandlerRestContext(ctx, &router)))
	mux.Handle("/l/", openedHandler(http.StripPrefix("/l/", links)))
	mux.Handle("/", http.FileServer(assets.AssetFile()))
	return &mux, nil
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

func secureHttpHandler(dev bool, handler http.Handler) http.Handler {
	if dev {
		return openedHandler(handler)
	} else {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("X-XSS-Protection", "1; mode=block")
			writer.Header().Set("X-Frame-Options", "SAMEORIGIN")
			writer.Header().Set("X-Content-Type-Options", "nosniff")
			handler.ServeHTTP(writer, request)
		})
	}
}
