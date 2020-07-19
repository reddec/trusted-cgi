package client

import (
	"context"
	client "github.com/reddec/jsonrpc2/client"
	api "github.com/reddec/trusted-cgi/api"
	application "github.com/reddec/trusted-cgi/application"
	"sync/atomic"
)

func DefaultQueuesAPI() *QueuesAPIClient {
	return &QueuesAPIClient{BaseURL: "https://127.0.0.1:3434/u/"}
}

type QueuesAPIClient struct {
	BaseURL  string
	sequence uint64
}

// Create queue and link it to lambda and start worker
func (impl *QueuesAPIClient) Create(ctx context.Context, token *api.Token, name string, lambda string) (reply *application.Queue, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "QueuesAPI.Create", atomic.AddUint64(&impl.sequence, 1), &reply, token, name, lambda)
	return
}

// Remove queue and stop worker
func (impl *QueuesAPIClient) Remove(ctx context.Context, token *api.Token, name string) (reply bool, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "QueuesAPI.Remove", atomic.AddUint64(&impl.sequence, 1), &reply, token, name)
	return
}

// Linked queues for lambda
func (impl *QueuesAPIClient) Linked(ctx context.Context, token *api.Token, lambda string) (reply []application.Queue, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "QueuesAPI.Linked", atomic.AddUint64(&impl.sequence, 1), &reply, token, lambda)
	return
}

// List of all queues
func (impl *QueuesAPIClient) List(ctx context.Context, token *api.Token) (reply []application.Queue, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "QueuesAPI.List", atomic.AddUint64(&impl.sequence, 1), &reply, token)
	return
}

// Assign lambda to queue (re-link)
func (impl *QueuesAPIClient) Assign(ctx context.Context, token *api.Token, name string, lambda string) (reply bool, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "QueuesAPI.Assign", atomic.AddUint64(&impl.sequence, 1), &reply, token, name, lambda)
	return
}
