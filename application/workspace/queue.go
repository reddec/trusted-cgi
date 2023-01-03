package workspace

import (
	"context"
	"fmt"
	"github.com/reddec/trusted-cgi/application/config"
	"io"
)

func NewQueue(rootDir string, cfg config.Queue, sync *Sync) (*Queue, error) {
	return nil, fmt.Errorf("TODO")
}

type Message struct {
	Environment map[string]string
	Payload     io.ReadCloser
}

type Queue struct {
}

func (q *Queue) Push(ctx context.Context, env map[string]string, data io.Reader) error {
	return nil
}

func (q *Queue) Peek(ctx context.Context) (*Message, error) {
	return nil, nil
}

func (q *Queue) Commit(ctx context.Context) error {

}
