package application

import (
	"context"
	"github.com/robfig/cron"
	"log"
	"time"
)

func (project *Project) RunCron(ctx context.Context) {
	now := time.Now()
	last := project.lastScheduler
	for _, app := range project.CloneApps() {
		app.RunScheduled(ctx, last, now)
	}
	project.lastScheduler = now
}

func (app *App) RunScheduled(ctx context.Context, last, now time.Time) {
	for expr, action := range app.Manifest.Cron {
		sched, err := cron.Parse(expr)
		if err != nil {
			log.Println(app.UID, expr, "-", err)
			continue
		}
		if !sched.Next(last).After(now) {
			_, err = app.InvokeAction(ctx, action)
			if err != nil {
				log.Println(app.UID, expr, action, err)
			}
		}
	}
}
