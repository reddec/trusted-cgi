package services

import (
	"bytes"
	"context"
	"github.com/reddec/trusted-cgi/api"
	"github.com/reddec/trusted-cgi/application"
	"github.com/reddec/trusted-cgi/stats"
	"github.com/reddec/trusted-cgi/types"
)

func NewLambdaSrv(cases application.Cases, tracker stats.Reader) *lambdaSrv {
	return &lambdaSrv{
		cases:   cases,
		tracker: tracker,
	}
}

type lambdaSrv struct {
	cases   application.Cases
	tracker stats.Reader
}

func (srv *lambdaSrv) Upload(ctx context.Context, token *api.Token, uid string, tarGz []byte) (bool, error) {
	fn, err := srv.cases.Platform().FindByUID(uid)
	if err != nil {
		return false, err
	}
	err = fn.Lambda.SetContent(bytes.NewReader(tarGz))
	if err != nil {
		return false, err
	}
	return true, nil
}

func (srv *lambdaSrv) Download(ctx context.Context, token *api.Token, uid string) ([]byte, error) {
	var out bytes.Buffer
	fn, err := srv.cases.Platform().FindByUID(uid)
	if err != nil {
		return nil, err
	}
	err = fn.Lambda.Content(&out)
	return out.Bytes(), err
}

func (srv *lambdaSrv) Push(ctx context.Context, token *api.Token, uid string, file string, content []byte) (bool, error) {
	fn, err := srv.cases.Platform().FindByUID(uid)
	if err != nil {
		return false, err
	}
	err = fn.Lambda.WriteFile(file, bytes.NewReader(content))
	return err == nil, err
}

func (srv *lambdaSrv) Pull(ctx context.Context, token *api.Token, uid string, file string) ([]byte, error) {
	fn, err := srv.cases.Platform().FindByUID(uid)
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	err = fn.Lambda.ReadFile(file, &out)
	return out.Bytes(), err
}

func (srv *lambdaSrv) Remove(ctx context.Context, token *api.Token, uid string) (bool, error) {
	err := srv.cases.Remove(uid)
	return err == nil, err
}

func (srv *lambdaSrv) Files(ctx context.Context, token *api.Token, uid string, dir string) ([]types.File, error) {
	fn, err := srv.cases.Platform().FindByUID(uid)
	if err != nil {
		return nil, err
	}
	return fn.Lambda.ListFiles(dir)
}

func (srv *lambdaSrv) Info(ctx context.Context, token *api.Token, uid string) (*application.Definition, error) {
	fn, err := srv.cases.Platform().FindByUID(uid)
	if err != nil {
		return nil, err
	}
	return fn, nil
}

func (srv *lambdaSrv) Update(ctx context.Context, token *api.Token, uid string, manifest types.Manifest) (*application.Definition, error) {
	fn, err := srv.cases.Platform().FindByUID(uid)
	if err != nil {
		return nil, err
	}
	if err := manifest.Validate(); err != nil {
		return nil, err
	}
	err = fn.Lambda.SetManifest(manifest)
	if err != nil {
		return nil, err
	}
	fn.Manifest = manifest
	return fn, nil
}

func (srv *lambdaSrv) CreateFile(ctx context.Context, token *api.Token, uid string, path string, dir bool) (bool, error) {
	fn, err := srv.cases.Platform().FindByUID(uid)
	if err != nil {
		return false, err
	}
	err = fn.Lambda.WriteFile(path, bytes.NewBufferString(""))
	return err == nil, err
}

func (srv *lambdaSrv) RemoveFile(ctx context.Context, token *api.Token, uid string, path string) (bool, error) {
	fn, err := srv.cases.Platform().FindByUID(uid)
	if err != nil {
		return false, err
	}
	err = fn.Lambda.RemoveFile(path)
	return err == nil, err
}

func (srv *lambdaSrv) RenameFile(ctx context.Context, token *api.Token, uid string, oldPath, newPath string) (bool, error) {
	fn, err := srv.cases.Platform().FindByUID(uid)
	if err != nil {
		return false, err
	}
	err = fn.Lambda.RenameFile(oldPath, newPath)
	return err == nil, err
}

func (srv *lambdaSrv) Stats(ctx context.Context, token *api.Token, uid string, limit int) ([]stats.Record, error) {
	return srv.tracker.LastByUID(uid, limit)
}

func (srv *lambdaSrv) Actions(ctx context.Context, token *api.Token, uid string) ([]string, error) {
	fn, err := srv.cases.Platform().FindByUID(uid)
	if err != nil {
		return nil, err
	}
	return fn.Lambda.Actions()
}

func (srv *lambdaSrv) Invoke(ctx context.Context, token *api.Token, uid string, action string) (string, error) {
	fn, err := srv.cases.Platform().FindByUID(uid)
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	err = srv.cases.Platform().Do(ctx, fn.Lambda, action, 0, &out)
	return out.String(), err
}

func (srv *lambdaSrv) Link(ctx context.Context, token *api.Token, uid string, alias string) (*application.Definition, error) {
	return srv.cases.Platform().Link(uid, alias)
}

func (srv *lambdaSrv) Unlink(ctx context.Context, token *api.Token, alias string) (*application.Definition, error) {
	return srv.cases.Platform().Unlink(alias)
}
