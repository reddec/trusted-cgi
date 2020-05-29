package client

import (
	"context"
	client "github.com/reddec/jsonrpc2/client"
	api "github.com/reddec/trusted-cgi/api"
	application "github.com/reddec/trusted-cgi/application"
	stats "github.com/reddec/trusted-cgi/stats"
	types "github.com/reddec/trusted-cgi/types"
	"sync/atomic"
)

func DefaultLambdaAPI() *LambdaAPIClient {
	return &LambdaAPIClient{BaseURL: "https://127.0.0.1:3434/u/"}
}

type LambdaAPIClient struct {
	BaseURL  string
	sequence uint64
}

// Upload content from .tar.gz archive to app and call Install handler (if defined)
func (impl *LambdaAPIClient) Upload(ctx context.Context, token *api.Token, uid string, tarGz []byte) (reply bool, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "LambdaAPI.Upload", atomic.AddUint64(&impl.sequence, 1), &reply, token, uid, tarGz)
	return
}

// Download content as .tar.gz archive from app
func (impl *LambdaAPIClient) Download(ctx context.Context, token *api.Token, uid string) (reply []byte, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "LambdaAPI.Download", atomic.AddUint64(&impl.sequence, 1), &reply, token, uid)
	return
}

// Push single file to app
func (impl *LambdaAPIClient) Push(ctx context.Context, token *api.Token, uid string, file string, content []byte) (reply bool, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "LambdaAPI.Push", atomic.AddUint64(&impl.sequence, 1), &reply, token, uid, file, content)
	return
}

// Pull single file from app
func (impl *LambdaAPIClient) Pull(ctx context.Context, token *api.Token, uid string, file string) (reply []byte, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "LambdaAPI.Pull", atomic.AddUint64(&impl.sequence, 1), &reply, token, uid, file)
	return
}

// Remove app and call Uninstall handler (if defined)
func (impl *LambdaAPIClient) Remove(ctx context.Context, token *api.Token, uid string) (reply bool, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "LambdaAPI.Remove", atomic.AddUint64(&impl.sequence, 1), &reply, token, uid)
	return
}

// Files in func dir
func (impl *LambdaAPIClient) Files(ctx context.Context, token *api.Token, uid string, dir string) (reply []*api.File, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "LambdaAPI.Files", atomic.AddUint64(&impl.sequence, 1), &reply, token, uid, dir)
	return
}

// Info about application
func (impl *LambdaAPIClient) Info(ctx context.Context, token *api.Token, uid string) (reply *application.App, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "LambdaAPI.Info", atomic.AddUint64(&impl.sequence, 1), &reply, token, uid)
	return
}

// Update application manifest
func (impl *LambdaAPIClient) Update(ctx context.Context, token *api.Token, uid string, manifest types.Manifest) (reply *application.App, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "LambdaAPI.Update", atomic.AddUint64(&impl.sequence, 1), &reply, token, uid, manifest)
	return
}

// Create file or directory inside app
func (impl *LambdaAPIClient) CreateFile(ctx context.Context, token *api.Token, uid string, path string, dir bool) (reply bool, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "LambdaAPI.CreateFile", atomic.AddUint64(&impl.sequence, 1), &reply, token, uid, path, dir)
	return
}

// Remove file or directory
func (impl *LambdaAPIClient) RemoveFile(ctx context.Context, token *api.Token, uid string, path string) (reply bool, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "LambdaAPI.RemoveFile", atomic.AddUint64(&impl.sequence, 1), &reply, token, uid, path)
	return
}

// Rename file or directory
func (impl *LambdaAPIClient) RenameFile(ctx context.Context, token *api.Token, uid string, oldPath string, newPath string) (reply bool, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "LambdaAPI.RenameFile", atomic.AddUint64(&impl.sequence, 1), &reply, token, uid, oldPath, newPath)
	return
}

// Stats for the app
func (impl *LambdaAPIClient) Stats(ctx context.Context, token *api.Token, uid string, limit int) (reply []stats.Record, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "LambdaAPI.Stats", atomic.AddUint64(&impl.sequence, 1), &reply, token, uid, limit)
	return
}

// Actions available for the app
func (impl *LambdaAPIClient) Actions(ctx context.Context, token *api.Token, uid string) (reply []string, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "LambdaAPI.Actions", atomic.AddUint64(&impl.sequence, 1), &reply, token, uid)
	return
}

// Invoke action in the app (if make installed)
func (impl *LambdaAPIClient) Invoke(ctx context.Context, token *api.Token, uid string, action string) (reply string, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "LambdaAPI.Invoke", atomic.AddUint64(&impl.sequence, 1), &reply, token, uid, action)
	return
}

// Make link/alias for app
func (impl *LambdaAPIClient) Link(ctx context.Context, token *api.Token, uid string, alias string) (reply *application.App, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "LambdaAPI.Link", atomic.AddUint64(&impl.sequence, 1), &reply, token, uid, alias)
	return
}

// Remove link
func (impl *LambdaAPIClient) Unlink(ctx context.Context, token *api.Token, alias string) (reply *application.App, err error) {
	err = client.CallHTTP(ctx, impl.BaseURL, "LambdaAPI.Unlink", atomic.AddUint64(&impl.sequence, 1), &reply, token, alias)
	return
}
