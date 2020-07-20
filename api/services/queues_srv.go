package services

import (
	"context"
	"github.com/reddec/trusted-cgi/api"
	"github.com/reddec/trusted-cgi/application"
)

func NewQueuesSrv(queues application.Queues) *queuesSrv {
	return &queuesSrv{queues: queues}
}

type queuesSrv struct {
	queues application.Queues
}

func (srv *queuesSrv) Create(ctx context.Context, token *api.Token, queue application.Queue) (*application.Queue, error) {
	return &queue, srv.queues.Add(queue)
}

func (srv *queuesSrv) Remove(ctx context.Context, token *api.Token, name string) (bool, error) {
	err := srv.queues.Remove(name)
	return err == nil, err
}

func (srv *queuesSrv) Linked(ctx context.Context, token *api.Token, lambda string) ([]application.Queue, error) {
	return srv.queues.Find(lambda), nil
}

func (srv *queuesSrv) List(ctx context.Context, token *api.Token) ([]application.Queue, error) {
	return srv.queues.List(), nil
}

func (srv *queuesSrv) Assign(ctx context.Context, token *api.Token, name string, lambda string) (bool, error) {
	err := srv.queues.Assign(name, lambda)
	return err == nil, err
}
