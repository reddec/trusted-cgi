package services

import (
	"context"
	"fmt"
	"github.com/reddec/trusted-cgi/api"
	"github.com/reddec/trusted-cgi/application"
	"github.com/reddec/trusted-cgi/stats"
)

func NewProjectSrv(cases application.Cases, tracker stats.Reader) *projectSrv {
	return &projectSrv{
		cases:   cases,
		tracker: tracker,
	}
}

type projectSrv struct {
	cases   application.Cases
	tracker stats.Reader // for stats
}

func (srv *projectSrv) Create(ctx context.Context, token *api.Token) (*application.Definition, error) {
	uid, err := srv.cases.Create(ctx)
	if err != nil {
		return nil, err
	}
	return srv.cases.Platform().FindByUID(uid)
}

func (srv *projectSrv) CreateFromGit(ctx context.Context, token *api.Token, repo string) (*application.Definition, error) {
	uid, err := srv.cases.CreateFromGit(ctx, repo)
	if err != nil {
		return nil, err
	}
	return srv.cases.Platform().FindByUID(uid)
}

func (srv *projectSrv) CreateFromTemplate(ctx context.Context, token *api.Token, templateName string) (*application.Definition, error) {
	possible, err := srv.cases.Templates()
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
	uid, err := srv.cases.CreateFromTemplate(ctx, *tpl)
	if err != nil {
		return nil, err
	}
	return srv.cases.Platform().FindByUID(uid)
}

func (srv *projectSrv) Config(ctx context.Context, token *api.Token) (*api.Settings, error) {
	pk, _ := srv.cases.PublicSSHKey()
	return &api.Settings{
		User:        srv.cases.Platform().Config().User,
		PublicKey:   string(pk),
		Environment: srv.cases.Platform().Config().Environment,
	}, nil
}

func (srv *projectSrv) SetEnvironment(ctx context.Context, token *api.Token, env api.Environment) (*api.Settings, error) {
	err := srv.cases.Platform().SetConfig(srv.cases.Platform().Config().WithEnv(env.Environment))
	if err != nil {
		return nil, err
	}
	return srv.Config(ctx, token)
}

func (srv *projectSrv) SetUser(ctx context.Context, token *api.Token, user string) (*api.Settings, error) {
	err := srv.cases.Platform().SetConfig(srv.cases.Platform().Config().WithUser(user))
	if err != nil {
		return nil, err
	}
	return srv.Config(ctx, token)
}

func (srv *projectSrv) AllTemplates(ctx context.Context, token *api.Token) ([]*api.TemplateStatus, error) {
	list, err := srv.cases.Templates()
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

func (srv *projectSrv) List(ctx context.Context, token *api.Token) ([]application.Definition, error) {
	return srv.cases.Platform().List(), nil
}

func (srv *projectSrv) Templates(ctx context.Context, token *api.Token) ([]*api.Template, error) {
	possible, err := srv.cases.Templates()
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
