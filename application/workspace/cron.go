package workspace

import (
	"context"
	"fmt"
	"github.com/reddec/trusted-cgi/application/config"
	"github.com/reddec/trusted-cgi/trace"
	"github.com/robfig/cron"
	"io"
)

func NewCron(project *Project, cfg *config.Cron) (cron.Job, error) {
	script, err := NewScript(project, &cfg.Script)
	if err != nil {
		return nil, fmt.Errorf("create script: %w", err)
	}
	return &cronJob{
		project: project,
		script:  script,
		config:  cfg,
	}, nil
}

type cronJob struct {
	project *Project
	script  *Script
	config  *config.Cron
}

func (cj *cronJob) Run() {
	tracer := cj.project.Trace()
	tracer.Set("cron", cj.config.Schedule)
	defer tracer.Close()

	ctx := trace.WithTrace(context.Background(), tracer)
	out, err := cj.script.Call(ctx, nil, nil)
	if err != nil {
		tracer.Set("error", err.Error())
		return
	}
	_, err = io.Copy(io.Discard, out)
	if err != nil {
		_ = out.Close()
		tracer.Set("copy_error", err.Error())
	}
}
