package workspace

import (
	"bytes"
	"context"
	"github.com/hashicorp/go-multierror"
	"github.com/robfig/cron"
	"io"
	"log"
)

var emptyStream = bytes.NewReader([]byte{})
var emptyContext = struct{}{}

func NewCron(sync []*Sync, async []*Async) cron.Job {
	return &cronJob{
		sync:  sync,
		async: async,
	}
}

type cronJob struct {
	sync  []*Sync
	async []*Async
}

func (cj *cronJob) Run() {
	ctx := context.Background()

	var wg multierror.Group

	for _, a := range cj.async {
		ref := a
		wg.Go(func() error {
			return ref.Push(nil, emptyStream, &emptyContext)
		})
	}

	for _, s := range cj.sync {
		ref := s
		wg.Go(func() error {
			out, err := ref.Call(ctx, nil, emptyStream, &emptyContext)
			if err != nil {
				return err
			}
			_, err = io.Copy(io.Discard, out)
			if err != nil {
				_ = out.Close()
				return err
			}
			return out.Close()
		})
	}

	if err := wg.Wait().ErrorOrNil(); err != nil {
		log.Println("cron failed:", err)
	}
}
