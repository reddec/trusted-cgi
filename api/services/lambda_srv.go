package services

import (
	"bytes"
	"context"
	"fmt"
	"github.com/reddec/trusted-cgi/api"
	"github.com/reddec/trusted-cgi/application"
	"github.com/reddec/trusted-cgi/stats"
	"github.com/reddec/trusted-cgi/types"
	"io/ioutil"
	"os"
)

func NewLambdaSrv(project *application.Project, tracker stats.Reader) *lambdaSrv {
	return &lambdaSrv{
		project: project,
		tracker: tracker,
	}
}

type lambdaSrv struct {
	project *application.Project
	tracker stats.Reader
}

func (srv *lambdaSrv) Upload(ctx context.Context, token *api.Token, uid string, tarGz []byte) (bool, error) {
	err := srv.project.Upload(ctx, uid, bytes.NewReader(tarGz))
	if err != nil {
		return false, err
	}
	return true, nil
}

func (srv *lambdaSrv) Download(ctx context.Context, token *api.Token, uid string) ([]byte, error) {
	var out bytes.Buffer
	err := srv.project.Download(ctx, uid, &out)
	return out.Bytes(), err
}

func (srv *lambdaSrv) Push(ctx context.Context, token *api.Token, uid string, file string, content []byte) (bool, error) {
	app := srv.project.FindApp(uid)
	if app == nil {
		return false, fmt.Errorf("unknown app")
	}

	err := app.WriteFile(file, content)
	return err == nil, err
}

func (srv *lambdaSrv) Pull(ctx context.Context, token *api.Token, uid string, file string) ([]byte, error) {
	app := srv.project.FindApp(uid)
	if app == nil {
		return nil, fmt.Errorf("unknown app")
	}
	return app.ReadFile(file)
}

func (srv *lambdaSrv) Remove(ctx context.Context, token *api.Token, uid string) (bool, error) {
	err := srv.project.Remove(ctx, uid)
	return err == nil, err
}

func (srv *lambdaSrv) Files(ctx context.Context, token *api.Token, uid string, dir string) ([]*api.File, error) {
	app := srv.project.FindApp(uid)
	if app == nil {
		return nil, fmt.Errorf("unknown app")
	}

	fpath, err := app.File(dir)
	if err != nil {
		return nil, err
	}
	list, err := ioutil.ReadDir(fpath)
	if err != nil {
		return nil, err
	}
	var ans = make([]*api.File, 0, len(list))
	for _, item := range list {
		ans = append(ans, &api.File{
			Dir:  item.IsDir(),
			Name: item.Name(),
		})
	}
	return ans, nil
}

func (srv *lambdaSrv) Info(ctx context.Context, token *api.Token, uid string) (*application.App, error) {
	app := srv.project.FindApp(uid)
	if app == nil {
		return nil, fmt.Errorf("unknown app")
	}
	return app, nil
}

func (srv *lambdaSrv) Update(ctx context.Context, token *api.Token, uid string, manifest types.Manifest) (*application.App, error) {
	app := srv.project.FindApp(uid)
	if app == nil {
		return nil, fmt.Errorf("unknown app")
	}
	if err := manifest.Validate(); err != nil {
		return nil, err
	}
	app.Manifest = manifest
	return app, app.Manifest.SaveAs(app.ManifestFile())
}

func (srv *lambdaSrv) CreateFile(ctx context.Context, token *api.Token, uid string, path string, dir bool) (bool, error) {
	app := srv.project.FindApp(uid)
	if app == nil {
		return false, fmt.Errorf("unknown app")
	}
	err := app.Touch(path, dir)
	return err == nil, err
}

func (srv *lambdaSrv) RemoveFile(ctx context.Context, token *api.Token, uid string, path string) (bool, error) {
	app := srv.project.FindApp(uid)
	if app == nil {
		return false, fmt.Errorf("unknown app")
	}
	fpath, err := app.File(path)
	if err != nil {
		return false, err
	}
	err = os.RemoveAll(fpath)
	return err == nil, err
}

func (srv *lambdaSrv) RenameFile(ctx context.Context, token *api.Token, uid string, oldPath, newPath string) (bool, error) {
	app := srv.project.FindApp(uid)
	if app == nil {
		return false, fmt.Errorf("unknown app")
	}
	opath, err := app.File(oldPath)
	if err != nil {
		return false, err
	}
	npath, err := app.File(newPath)
	if err != nil {
		return false, err
	}
	err = os.Rename(opath, npath)
	return err == nil, err
}

func (srv *lambdaSrv) Stats(ctx context.Context, token *api.Token, uid string, limit int) ([]stats.Record, error) {
	return srv.tracker.LastByUID(uid, limit)
}

func (srv *lambdaSrv) Actions(ctx context.Context, token *api.Token, uid string) ([]string, error) {
	app := srv.project.FindApp(uid)
	if app == nil {
		return nil, fmt.Errorf("unknown app")
	}
	return app.ListActions()
}

func (srv *lambdaSrv) Invoke(ctx context.Context, token *api.Token, uid string, action string) (string, error) {
	app := srv.project.FindApp(uid)
	if app == nil {
		return "", fmt.Errorf("unknown app")
	}
	return app.InvokeAction(ctx, action, 0)
}

func (srv *lambdaSrv) Link(ctx context.Context, token *api.Token, uid string, alias string) (*application.App, error) {
	return srv.project.Link(uid, alias)
}

func (srv *lambdaSrv) Unlink(ctx context.Context, token *api.Token, alias string) (*application.App, error) {
	return srv.project.Unlink(alias)
}
