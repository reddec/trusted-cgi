package workspace

import (
	"bytes"
	"context"
	"github.com/hashicorp/go-multierror"
	"github.com/reddec/trusted-cgi/application/stats"
	"github.com/robfig/cron"
	"io"
	"log"
)

var emptyStream = bytes.NewReader([]byte{})
var emptyContext = struct{}{}

func NewCron(sync []*Sync, async []*Async, monitor *stats.CronMonitor) cron.Job {
	return &cronJob{
		monitor: monitor,
		sync:    sync,
		async:   async,
	}
}

type cronJob struct {
	monitor *stats.CronMonitor
	sync    []*Sync
	async   []*Async
}

func (cj *cronJob) Run() {
	ctx := context.Background()
	running := cj.monitor.Started()
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
	err := wg.Wait().ErrorOrNil()
	if err != nil {
		log.Println("cron failed:", err)
	}
	_ = running.Finished(err)
}
