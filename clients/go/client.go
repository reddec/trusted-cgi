package api

import (
	"context"
	client "github.com/reddec/jsonrpc2/client"
	application "github.com/reddec/trusted-cgi/application"
	server "github.com/reddec/trusted-cgi/server"
	stats "github.com/reddec/trusted-cgi/stats"
	types "github.com/reddec/trusted-cgi/types"
	"sync/atomic"
)

func Default() *APIClient {
	return &APIClient{BaseURL: "https://127.0.0.1:3434/u/"}
}

type APIClient struct {
	BaseURL  string
	sequence uint64
}

// Login user by username and password. Returns signed JWT
func (impl *APIClient) Login(ctx context.Context, login string, password string) (reply *server.Token, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "API.Login", atomic.AddUint64(&impl.sequence, 1), &reply, login, password)
	return
}

// Change password for the user
func (impl *APIClient) ChangePassword(ctx context.Context, token *server.Token, password string) (reply bool, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "API.ChangePassword", atomic.AddUint64(&impl.sequence, 1), &reply, token, password)
	return
}

// Create new app (lambda)
func (impl *APIClient) Create(ctx context.Context, token *server.Token) (reply *application.App, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "API.Create", atomic.AddUint64(&impl.sequence, 1), &reply, token)
	return
}

// Project configuration
func (impl *APIClient) Config(ctx context.Context, token *server.Token) (reply *application.ProjectConfig, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "API.Config", atomic.AddUint64(&impl.sequence, 1), &reply, token)
	return
}

// Apply new configuration and save it
func (impl *APIClient) Apply(ctx context.Context, token *server.Token, config application.ProjectConfig) (reply bool, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "API.Apply", atomic.AddUint64(&impl.sequence, 1), &reply, token, config)
	return
}

// Get all templates without filtering
func (impl *APIClient) AllTemplates(ctx context.Context, token *server.Token) (reply []*server.TemplateStatus, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "API.AllTemplates", atomic.AddUint64(&impl.sequence, 1), &reply, token)
	return
}

// Create new app/lambda/function using pre-defined template
func (impl *APIClient) CreateFromTemplate(ctx context.Context, token *server.Token, templateName string) (reply *application.App, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "API.CreateFromTemplate", atomic.AddUint64(&impl.sequence, 1), &reply, token, templateName)
	return
}

// Upload content from .tar.gz archive to app and call Install handler (if defined)
func (impl *APIClient) Upload(ctx context.Context, token *server.Token, uid string, tarGz []byte) (reply bool, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "API.Upload", atomic.AddUint64(&impl.sequence, 1), &reply, token, uid, tarGz)
	return
}

// Download content as .tar.gz archive from app
func (impl *APIClient) Download(ctx context.Context, token *server.Token, uid string) (reply []byte, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "API.Download", atomic.AddUint64(&impl.sequence, 1), &reply, token, uid)
	return
}

// Push single file to app
func (impl *APIClient) Push(ctx context.Context, token *server.Token, uid string, file string, content []byte) (reply bool, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "API.Push", atomic.AddUint64(&impl.sequence, 1), &reply, token, uid, file, content)
	return
}

// Pull single file from app
func (impl *APIClient) Pull(ctx context.Context, token *server.Token, uid string, file string) (reply []byte, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "API.Pull", atomic.AddUint64(&impl.sequence, 1), &reply, token, uid, file)
	return
}

// List available apps (lambdas) in a project
func (impl *APIClient) List(ctx context.Context, token *server.Token) (reply []*application.App, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "API.List", atomic.AddUint64(&impl.sequence, 1), &reply, token)
	return
}

// Remove app and call Uninstall handler (if defined)
func (impl *APIClient) Remove(ctx context.Context, token *server.Token, uid string) (reply bool, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "API.Remove", atomic.AddUint64(&impl.sequence, 1), &reply, token, uid)
	return
}

// Templates with filter by availability including embedded
func (impl *APIClient) Templates(ctx context.Context, token *server.Token) (reply []*server.Template, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "API.Templates", atomic.AddUint64(&impl.sequence, 1), &reply, token)
	return
}

// Files in func dir
func (impl *APIClient) Files(ctx context.Context, token *server.Token, name string, dir string) (reply []*server.File, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "API.Files", atomic.AddUint64(&impl.sequence, 1), &reply, token, name, dir)
	return
}

// Info about application
func (impl *APIClient) Info(ctx context.Context, token *server.Token, uid string) (reply *application.App, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "API.Info", atomic.AddUint64(&impl.sequence, 1), &reply, token, uid)
	return
}

// Update application manifest
func (impl *APIClient) Update(ctx context.Context, token *server.Token, uid string, manifest types.Manifest) (reply *application.App, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "API.Update", atomic.AddUint64(&impl.sequence, 1), &reply, token, uid, manifest)
	return
}

// Create file or directory inside app
func (impl *APIClient) CreateFile(ctx context.Context, token *server.Token, uid string, path string, dir bool) (reply bool, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "API.CreateFile", atomic.AddUint64(&impl.sequence, 1), &reply, token, uid, path, dir)
	return
}

// Remove file or directory
func (impl *APIClient) RemoveFile(ctx context.Context, token *server.Token, uid string, path string) (reply bool, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "API.RemoveFile", atomic.AddUint64(&impl.sequence, 1), &reply, token, uid, path)
	return
}

// Rename file or directory
func (impl *APIClient) RenameFile(ctx context.Context, token *server.Token, uid string, oldPath string, newPath string) (reply bool, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "API.RenameFile", atomic.AddUint64(&impl.sequence, 1), &reply, token, uid, oldPath, newPath)
	return
}

// Global last records
func (impl *APIClient) GlobalStats(ctx context.Context, token *server.Token, limit int) (reply []stats.Record, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "API.GlobalStats", atomic.AddUint64(&impl.sequence, 1), &reply, token, limit)
	return
}

// Stats for the app
func (impl *APIClient) Stats(ctx context.Context, token *server.Token, uid string, limit int) (reply []stats.Record, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "API.Stats", atomic.AddUint64(&impl.sequence, 1), &reply, token, uid, limit)
	return
}

// Actions available for the app
func (impl *APIClient) Actions(ctx context.Context, token *server.Token, uid string) (reply []string, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "API.Actions", atomic.AddUint64(&impl.sequence, 1), &reply, token, uid)
	return
}

// Invoke action in the app (if make installed)
func (impl *APIClient) Invoke(ctx context.Context, token *server.Token, uid string, action string) (reply bool, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "API.Invoke", atomic.AddUint64(&impl.sequence, 1), &reply, token, uid, action)
	return
}

// Make link/alias for app
func (impl *APIClient) Link(ctx context.Context, token *server.Token, uid string, alias string) (reply *application.App, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "API.Link", atomic.AddUint64(&impl.sequence, 1), &reply, token, uid, alias)
	return
}

// Remove link
func (impl *APIClient) Unlink(ctx context.Context, token *server.Token, alias string) (reply *application.App, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "API.Unlink", atomic.AddUint64(&impl.sequence, 1), &reply, token, alias)
	return
}
