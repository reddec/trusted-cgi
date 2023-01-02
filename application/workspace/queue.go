package workspace

import (
	"context"
	"fmt"
	"github.com/reddec/trusted-cgi/application/config"
)

func NewQueue(rootDir string, cfg config.Queue) (*Queue, error) {
	return nil, fmt.Errorf("TODO")
}

type Envelope struct {
}

type Queue struct {
}

func (q *Queue) Push(ctx context.Context, message Envelope) error {
	return nil
}

func (q *Queue) Pull(ctx context.Context) (Envelope, error) {
	return Envelope{}, nil
}
