package services

import (
	"context"
	"github.com/reddec/trusted-cgi/api"
	"github.com/reddec/trusted-cgi/application"
)

func NewPoliciesSrv(policies application.Policies) *policiesSrv {
	return &policiesSrv{policies: policies}
}

type policiesSrv struct {
	policies application.Policies
}

func (srv *policiesSrv) List(ctx context.Context, token *api.Token) ([]application.Policy, error) {
	return srv.policies.List(), nil
}

func (srv *policiesSrv) Create(ctx context.Context, token *api.Token, policy string, definition application.PolicyDefinition) (*application.Policy, error) {
	return srv.policies.Create(policy, definition)
}

func (srv *policiesSrv) Remove(ctx context.Context, token *api.Token, policy string) (bool, error) {
	err := srv.policies.Remove(policy)
	return err == nil, err
}

func (srv *policiesSrv) Update(ctx context.Context, token *api.Token, policy string, definition application.PolicyDefinition) (bool, error) {
	err := srv.policies.Update(policy, definition)
	return err == nil, err
}

func (srv *policiesSrv) Apply(ctx context.Context, token *api.Token, lambda string, policy string) (bool, error) {
	err := srv.policies.Apply(lambda, policy)
	return err == nil, err
}

func (srv *policiesSrv) Clear(ctx context.Context, token *api.Token, lambda string) (bool, error) {
	err := srv.policies.Clear(lambda)
	return err == nil, err
}
