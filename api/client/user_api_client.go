package client

import (
	"context"
	client "github.com/reddec/jsonrpc2/client"
	api "github.com/reddec/trusted-cgi/api"
	"sync/atomic"
)

func DefaultUserAPI() *UserAPIClient {
	return &UserAPIClient{BaseURL: "https://127.0.0.1:3434/u/"}
}

type UserAPIClient struct {
	BaseURL  string
	sequence uint64
}

// Login user by username and password. Returns signed JWT
func (impl *UserAPIClient) Login(ctx context.Context, login string, password string) (reply *api.Token, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "UserAPI.Login", atomic.AddUint64(&impl.sequence, 1), &reply, login, password)
	return
}

// Change password for the user
func (impl *UserAPIClient) ChangePassword(ctx context.Context, token *api.Token, password string) (reply bool, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "UserAPI.ChangePassword", atomic.AddUint64(&impl.sequence, 1), &reply, token, password)
	return
}
