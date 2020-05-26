package server

import (
	"bytes"
	"context"
	"fmt"
	"github.com/reddec/trusted-cgi/application"
	"github.com/reddec/trusted-cgi/stats"
	"github.com/reddec/trusted-cgi/templates"
	"github.com/reddec/trusted-cgi/types"
	"io/ioutil"
	"os"
	"sync"
)

type apiImpl struct {
	lock         sync.Mutex
	server       *Server
	project      *application.Project
	tracker      stats.Reader
	templatesDir string
}

func (srv *apiImpl) Login(ctx context.Context, login, password string) (*Token, error) {
	return srv.server.Login(login, password)
}

func (srv *apiImpl) ChangePassword(ctx context.Context, token *Token, password string) (bool, error) {
	if srv.server.Admin != token.Login {
		return false, fmt.Errorf("multi-seat not yet supported")
	}
	srv.server.SetPassword(password)
	err := srv.server.Save()
	return err == nil, err
}

func (srv *apiImpl) Config(ctx context.Context, token *Token) (*application.ProjectConfig, error) {
	return &srv.project.ProjectConfig, nil
}

func (srv *apiImpl) Apply(ctx context.Context, token *Token, config application.ProjectConfig) (bool, error) {
	err := srv.project.ApplyConfig(config)
	return err == nil, err
}

func (srv *apiImpl) AllTemplates(ctx context.Context, token *Token) ([]*TemplateStatus, error) {
	list, err := templates.List(srv.templatesDir)
	if err != nil {
		return nil, err
	}
	var ans = make([]*TemplateStatus, 0, len(list))
	for name, t := range list {
		ans = append(ans, &TemplateStatus{
			Template: Template{
				Name:        name,
				Description: t.Description,
			},
			Available: t.IsAvailable(ctx),
		})
	}

	return ans, nil
}

func (srv *apiImpl) Create(ctx context.Context, token *Token) (*application.App, error) {
	return srv.project.Create(ctx)
}

func (srv *apiImpl) List(ctx context.Context, token *Token) ([]*application.App, error) {
	return srv.project.List(), nil
}

func (srv *apiImpl) Remove(ctx context.Context, token *Token, uid string) (bool, error) {
	err := srv.project.Remove(ctx, uid)
	return err == nil, err
}

func (srv *apiImpl) Upload(ctx context.Context, token *Token, uid string, tarGz []byte) (bool, error) {
	err := srv.project.Upload(ctx, uid, bytes.NewReader(tarGz))
	if err != nil {
		return false, err
	}
	return true, nil
}

func (srv *apiImpl) Download(ctx context.Context, token *Token, uid string) ([]byte, error) {
	var out bytes.Buffer
	err := srv.project.Download(ctx, uid, &out)
	return out.Bytes(), err
}

func (srv *apiImpl) Push(ctx context.Context, token *Token, uid string, file string, content []byte) (bool, error) {
	app := srv.project.FindApp(uid)
	if app == nil {
		return false, fmt.Errorf("unknown app")
	}

	err := app.WriteFile(file, content)
	return err == nil, err
}

func (srv *apiImpl) Pull(ctx context.Context, token *Token, uid string, file string) ([]byte, error) {
	app := srv.project.FindApp(uid)
	if app == nil {
		return nil, fmt.Errorf("unknown app")
	}
	return app.ReadFile(file)
}

func (srv *apiImpl) CreateFromTemplate(ctx context.Context, token *Token, templateName string) (*application.App, error) {
	possible, err := templates.List(srv.templatesDir)
	if err != nil {
		return nil, err
	}
	tpl, ok := possible[templateName]
	if !ok {
		return nil, fmt.Errorf("unknown tempalte %s", templateName)
	}
	if !tpl.IsAvailable(ctx) {
		return nil, fmt.Errorf("template %s is not supported", templateName)
	}
	return srv.project.CreateFromTemplate(ctx, tpl)
}

func (srv *apiImpl) Templates(ctx context.Context, token *Token) ([]*Template, error) {
	possible, err := templates.List(srv.templatesDir)
	if err != nil {
		return nil, err
	}
	var ans = make([]*Template, 0, len(possible))
	for name, info := range possible {
		if info.IsAvailable(ctx) {
			ans = append(ans, &Template{
				Name:        name,
				Description: info.Description,
			})
		}
	}
	return ans, nil
}

func (srv *apiImpl) Files(ctx context.Context, token *Token, uid string, dir string) ([]*File, error) {
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
	var ans = make([]*File, 0, len(list))
	for _, item := range list {
		ans = append(ans, &File{
			Dir:  item.IsDir(),
			Name: item.Name(),
		})
	}
	return ans, nil
}

func (srv *apiImpl) Info(ctx context.Context, token *Token, uid string) (*application.App, error) {
	app := srv.project.FindApp(uid)
	if app == nil {
		return nil, fmt.Errorf("unknown app")
	}
	return app, nil
}

func (srv *apiImpl) CreateFile(ctx context.Context, token *Token, uid string, path string, dir bool) (bool, error) {
	app := srv.project.FindApp(uid)
	if app == nil {
		return false, fmt.Errorf("unknown app")
	}
	err := app.Touch(path, dir)
	return err == nil, err
}

func (srv *apiImpl) RemoveFile(ctx context.Context, token *Token, uid string, path string) (bool, error) {
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

func (srv *apiImpl) RenameFile(ctx context.Context, token *Token, uid string, oldPath, newPath string) (bool, error) {
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

func (srv *apiImpl) Update(ctx context.Context, token *Token, uid string, manifest types.Manifest) (*application.App, error) {
	app := srv.project.FindApp(uid)
	if app == nil {
		return nil, fmt.Errorf("unknown app")
	}
	app.Manifest = manifest
	return app, app.Manifest.SaveAs(app.ManifestFile())
}

func (srv *apiImpl) GlobalStats(ctx context.Context, token *Token, limit int) ([]stats.Record, error) {
	return srv.tracker.Last(limit)
}

func (srv *apiImpl) Stats(ctx context.Context, token *Token, uid string, limit int) ([]stats.Record, error) {
	return srv.tracker.LastByUID(uid, limit)
}

func (srv *apiImpl) Actions(ctx context.Context, token *Token, uid string) ([]string, error) {
	app := srv.project.FindApp(uid)
	if app == nil {
		return nil, fmt.Errorf("unknown app")
	}
	return app.ListActions()
}

func (srv *apiImpl) Invoke(ctx context.Context, token *Token, uid string, action string) (string, error) {
	app := srv.project.FindApp(uid)
	if app == nil {
		return "", fmt.Errorf("unknown app")
	}
	return app.InvokeAction(ctx, action)
}

func (srv *apiImpl) Link(ctx context.Context, token *Token, uid string, alias string) (*application.App, error) {
	return srv.project.Link(uid, alias)
}

func (srv *apiImpl) Unlink(ctx context.Context, token *Token, alias string) (*application.App, error) {
	panic("implement me")
}
