package services

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/reddec/jsonrpc2"
	"github.com/reddec/trusted-cgi/api"
)

const (
	defaultLifeTime = 30 * 24 * time.Hour
	defaultLogin    = "admin"
)

func CreateUserSrv(configFile string, initialPassword string) (*userSrv, error) {
	if srv, err := LoadUserSrv(configFile); err == nil {
		return srv, nil
	}

	srv := &userSrv{
		configFile: configFile,
		config: userConfig{
			LifeTime: defaultLifeTime,
			Admin:    defaultLogin,
		},
		secret: uuid.New().String(),
	}
	err := os.MkdirAll(filepath.Dir(configFile), 0755)
	if err != nil {
		return nil, err
	}
	_, err = srv.ChangePassword(context.Background(), &api.Token{
		Login: defaultLogin,
	}, initialPassword)
	return srv, err
}

func LoadUserSrv(configFile string) (*userSrv, error) {
	var cfg userConfig
	err := cfg.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	return &userSrv{
		configFile: configFile,
		config:     cfg,
		secret:     uuid.New().String(),
	}, nil
}

type userSrv struct {
	configFile string
	config     userConfig
	secret     string
	lock       sync.RWMutex
}

func (srv *userSrv) Login(ctx context.Context, login, password string) (*api.Token, error) {
	srv.lock.RLock()
	defer srv.lock.RUnlock()
	err := srv.config.ValidateUser(login, password)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iat":  now.Unix(),
		"exp":  now.Add(srv.config.LifeTime).Unix(),
		"user": login,
	})
	v, err := tok.SignedString([]byte(srv.secret))
	return &api.Token{Data: v}, err
}

func (srv *userSrv) ChangePassword(ctx context.Context, token *api.Token, password string) (bool, error) {
	srv.lock.Lock()
	defer srv.lock.Unlock()

	salt := uuid.New().String()
	data := sha512.Sum512([]byte(password + salt))
	srv.config.Salt = salt
	srv.config.Hash = data[:]

	err := srv.config.WriteFile(srv.configFile)
	return err == nil, err
}

func (srv *userSrv) ValidateToken(ctx context.Context, token *api.Token) error {
	if token == nil {
		return fmt.Errorf("token not provided")
	}
	claims, err := jwt.Parse(token.Data, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(srv.secret), nil
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

type userConfig struct {
	Admin    string        `json:"admin"`     // login for admin authorization
	Salt     string        `json:"salt"`      // password salt
	Hash     []byte        `json:"hash"`      // password hash
	LifeTime time.Duration `json:"life_time"` // life time for JWT
}

func (uc *userConfig) ReadFile(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(uc)
}

func (uc *userConfig) WriteFile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(uc)
}

func (uc *userConfig) ValidateUser(login, password string) error {
	data := sha512.Sum512([]byte(password + uc.Salt))
	if bytes.Compare(data[:], uc.Hash) != 0 || uc.Admin != login {
		return fmt.Errorf("password or login is invalid")
	}
	return nil
}
