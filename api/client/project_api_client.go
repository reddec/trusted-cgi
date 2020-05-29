package client

import (
	"context"
	client "github.com/reddec/jsonrpc2/client"
	api "github.com/reddec/trusted-cgi/api"
	application "github.com/reddec/trusted-cgi/application"
	stats "github.com/reddec/trusted-cgi/stats"
	"sync/atomic"
)

func DefaultProjectAPI() *ProjectAPIClient {
	return &ProjectAPIClient{BaseURL: "https://127.0.0.1:3434/u/"}
}

type ProjectAPIClient struct {
	BaseURL  string
	sequence uint64
}

// Get global configuration
func (impl *ProjectAPIClient) Config(ctx context.Context, token *api.Token) (reply *api.Settings, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "ProjectAPI.Config", atomic.AddUint64(&impl.sequence, 1), &reply, token)
	return
}

// Change effective user
func (impl *ProjectAPIClient) SetUser(ctx context.Context, token *api.Token, user string) (reply *api.Settings, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "ProjectAPI.SetUser", atomic.AddUint64(&impl.sequence, 1), &reply, token, user)
	return
}

// Get all templates without filtering
func (impl *ProjectAPIClient) AllTemplates(ctx context.Context, token *api.Token) (reply []*api.TemplateStatus, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "ProjectAPI.AllTemplates", atomic.AddUint64(&impl.sequence, 1), &reply, token)
	return
}

// List available apps (lambdas) in a project
func (impl *ProjectAPIClient) List(ctx context.Context, token *api.Token) (reply []*application.App, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "ProjectAPI.List", atomic.AddUint64(&impl.sequence, 1), &reply, token)
	return
}

// Templates with filter by availability including embedded
func (impl *ProjectAPIClient) Templates(ctx context.Context, token *api.Token) (reply []*api.Template, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "ProjectAPI.Templates", atomic.AddUint64(&impl.sequence, 1), &reply, token)
	return
}

// Global last records
func (impl *ProjectAPIClient) Stats(ctx context.Context, token *api.Token, limit int) (reply []stats.Record, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "ProjectAPI.Stats", atomic.AddUint64(&impl.sequence, 1), &reply, token, limit)
	return
}

// Create new app (lambda)
func (impl *ProjectAPIClient) Create(ctx context.Context, token *api.Token) (reply *application.App, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "ProjectAPI.Create", atomic.AddUint64(&impl.sequence, 1), &reply, token)
	return
}

// Create new app/lambda/function using pre-defined template
func (impl *ProjectAPIClient) CreateFromTemplate(ctx context.Context, token *api.Token, templateName string) (reply *application.App, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "ProjectAPI.CreateFromTemplate", atomic.AddUint64(&impl.sequence, 1), &reply, token, templateName)
	return
}
