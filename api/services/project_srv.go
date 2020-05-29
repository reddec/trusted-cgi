package services

import (
	"context"
	"fmt"
	"github.com/reddec/trusted-cgi/api"
	"github.com/reddec/trusted-cgi/application"
	"github.com/reddec/trusted-cgi/stats"
	"github.com/reddec/trusted-cgi/templates"
)

func NewProjectSrv(project *application.Project, tracker stats.Reader, templatesDir string) *projectSrv {
	return &projectSrv{
		project:      project,
		tracker:      tracker,
		templatesDir: templatesDir,
	}
}

type projectSrv struct {
	project      *application.Project
	tracker      stats.Reader // for stats
	templatesDir string
}

func (srv *projectSrv) Create(ctx context.Context, token *api.Token) (*application.App, error) {
	return srv.project.Create(ctx)
}

func (srv *projectSrv) CreateFromGit(ctx context.Context, token *api.Token, repo string) (*application.App, error) {
	return srv.project.CreateFromGit(ctx, repo)
}

func (srv *projectSrv) CreateFromTemplate(ctx context.Context, token *api.Token, templateName string) (*application.App, error) {
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

func (srv *projectSrv) Config(ctx context.Context, token *api.Token) (*api.Settings, error) {
	return &api.Settings{
		User:      srv.project.RunnerUser(),
		PublicKey: string(srv.project.PublicKey()),
	}, nil
}

func (srv *projectSrv) SetUser(ctx context.Context, token *api.Token, user string) (*api.Settings, error) {
	err := srv.project.ChangeUser(user)
	if err != nil {
		return nil, err
	}
	return srv.Config(ctx, token)
}

func (srv *projectSrv) AllTemplates(ctx context.Context, token *api.Token) ([]*api.TemplateStatus, error) {
	list, err := templates.List(srv.templatesDir)
	if err != nil {
		return nil, err
	}
	var ans = make([]*api.TemplateStatus, 0, len(list))
	for name, t := range list {
		ans = append(ans, &api.TemplateStatus{
			Name:        name,
			Description: t.Description,
			Available:   t.IsAvailable(ctx),
		})
	}

	return ans, nil
}

func (srv *projectSrv) List(ctx context.Context, token *api.Token) ([]*application.App, error) {
	return srv.project.List(), nil
}

func (srv *projectSrv) Templates(ctx context.Context, token *api.Token) ([]*api.Template, error) {
	possible, err := templates.List(srv.templatesDir)
	if err != nil {
		return nil, err
	}
	var ans = make([]*api.Template, 0, len(possible))
	for name, info := range possible {
		if info.IsAvailable(ctx) {
			ans = append(ans, &api.Template{
				Name:        name,
				Description: info.Description,
			})
		}
	}
	return ans, nil
}

func (srv *projectSrv) Stats(ctx context.Context, token *api.Token, limit int) ([]stats.Record, error) {
	return srv.tracker.Last(limit)
}
