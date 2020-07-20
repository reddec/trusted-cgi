package client

import (
	"context"
	client "github.com/reddec/jsonrpc2/client"
	api "github.com/reddec/trusted-cgi/api"
	application "github.com/reddec/trusted-cgi/application"
	"sync/atomic"
)

func DefaultPoliciesAPI() *PoliciesAPIClient {
	return &PoliciesAPIClient{BaseURL: "https://127.0.0.1:3434/u/"}
}

type PoliciesAPIClient struct {
	BaseURL  string
	sequence uint64
}

// List all policies
func (impl *PoliciesAPIClient) List(ctx context.Context, token *api.Token) (reply []application.Policy, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "PoliciesAPI.List", atomic.AddUint64(&impl.sequence, 1), &reply, token)
	return
}

// Create new policy
func (impl *PoliciesAPIClient) Create(ctx context.Context, token *api.Token, policy string, definition application.PolicyDefinition) (reply *application.Policy, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "PoliciesAPI.Create", atomic.AddUint64(&impl.sequence, 1), &reply, token, policy, definition)
	return
}

// Remove policy
func (impl *PoliciesAPIClient) Remove(ctx context.Context, token *api.Token, policy string) (reply bool, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "PoliciesAPI.Remove", atomic.AddUint64(&impl.sequence, 1), &reply, token, policy)
	return
}

// Update policy definition
func (impl *PoliciesAPIClient) Update(ctx context.Context, token *api.Token, policy string, definition application.PolicyDefinition) (reply bool, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "PoliciesAPI.Update", atomic.AddUint64(&impl.sequence, 1), &reply, token, policy, definition)
	return
}

// Apply policy for the resource
func (impl *PoliciesAPIClient) Apply(ctx context.Context, token *api.Token, lambda string, policy string) (reply bool, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "PoliciesAPI.Apply", atomic.AddUint64(&impl.sequence, 1), &reply, token, lambda, policy)
	return
}

// Clear applied policy for the lambda
func (impl *PoliciesAPIClient) Clear(ctx context.Context, token *api.Token, lambda string) (reply bool, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "PoliciesAPI.Clear", atomic.AddUint64(&impl.sequence, 1), &reply, token, lambda)
	return
}
