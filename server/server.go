package server

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/reddec/jsonrpc2"
	"github.com/reddec/trusted-cgi/application"
	"github.com/reddec/trusted-cgi/assets"
	"github.com/reddec/trusted-cgi/stats"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func Open(configFile string, initialPassword string) (*Server, error) {
	var srv = &Server{
		configFile: configFile,
	}
	err := os.MkdirAll(filepath.Dir(configFile), 0755)
	if err != nil {
		return nil, err
	}
	err = srv.read()
	if err == nil {
		return srv, nil
	}
	if !os.IsNotExist(err) {
		return nil, err
	}
	srv.SetPassword(initialPassword)
	srv.Admin = "admin"
	srv.LifeTime = 30 * 24 * time.Hour
	srv.Secret = uuid.New().String()
	return srv, srv.Save()
}

type Server struct {
	Admin    string        `json:"admin"`     // login for admin authorization
	Salt     string        `json:"salt"`      // password salt
	Hash     []byte        `json:"hash"`      // password hash
	Secret   string        `json:"secret"`    // secret for JWT
	LifeTime time.Duration `json:"life_time"` // life time for JWT

	configFile string
	lock       sync.Mutex
}

func (srv *Server) validateUser(login, password string) error {

	data := sha512.Sum512([]byte(password + srv.Salt))
	if bytes.Compare(data[:], srv.Hash) != 0 || srv.Admin != login {
		return fmt.Errorf("password or login is invalid")
	}
	return nil
}

func (srv *Server) SetPassword(password string) {
	srv.lock.Lock()
	defer srv.lock.Unlock()

	salt := uuid.New().String()
	data := sha512.Sum512([]byte(password + salt))
	srv.Salt = salt
	srv.Hash = data[:]
}

func (srv *Server) Save() error {
	srv.lock.Lock()
	defer srv.lock.Unlock()

	f, err := os.Create(srv.configFile)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(srv)
}

func (srv *Server) read() error {
	f, err := os.Open(srv.configFile)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(srv)
}

func (srv *Server) Handler(ctx context.Context, project *application.Project, templatesDir string, dev bool, tracker stats.Stats) (http.Handler, error) {
	apps, err := project.Handler(ctx, tracker)
	if err != nil {
		return nil, err
	}

	links, err := project.HandlerAlias(ctx, tracker)
	if err != nil {
		return nil, err
	}

	var userApi jsonrpc2.Router
	registerAPI(&userApi, &apiImpl{
		server:       srv,
		project:      project,
		templatesDir: templatesDir,
		tracker:      tracker,
	}, srv)

	var mux http.ServeMux
	mux.Handle("/a/", openedHandler(http.StripPrefix("/a/", apps)))
	mux.Handle("/u/", secureHttpHandler(dev, jsonrpc2.HandlerRestContext(ctx, &userApi)))
	mux.Handle("/l/", openedHandler(http.StripPrefix("/l/", links)))
	mux.Handle("/", http.FileServer(assets.AssetFile()))
	return &mux, nil
}

// Login (as admin) by user and password and get JWT
func (srv *Server) Login(login, password string) (*Token, error) {
	srv.lock.Lock()
	defer srv.lock.Unlock()
	err := srv.validateUser(login, password)
	if err != nil {
		return nil, err
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iat":  time.Now().Add(srv.LifeTime),
		"user": login,
	})
	v, err := tok.SignedString([]byte(srv.Secret))
	return &Token{Data: v}, err
}

// Validate user token
func (srv *Server) ValidateToken(ctx context.Context, token *Token) error {
	if token == nil {
		return fmt.Errorf("token not provided")
	}
	srv.lock.Lock()
	secret := srv.Secret
	srv.lock.Unlock()
	claims, err := jwt.Parse(token.Data, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return &jsonrpc2.Error{
			Code:    403,
			Message: fmt.Sprintf("token validation failed: %s", err),
		}
	}

	if payload, ok := claims.Claims.(jwt.MapClaims); ok {
		if usr, ok := payload["user"]; ok {
			if v, ok := usr.(string); ok {
				token.Login = v
			}
		}
	}
	if token.Login == "" {
		return &jsonrpc2.Error{
			Code:    1403,
			Message: fmt.Sprintf("token validation failed: no login in payload"),
		}
	}

	return nil
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
