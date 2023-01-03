package stats

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const DatabaseTimeout = 5 * time.Second // we can not use global ctx or we can not record cancellation event

func NewMonitor(dbtx DBTX) *WorkspaceMonitor {
	return &WorkspaceMonitor{q: New(dbtx)}
}

type WorkspaceMonitor struct {
	q       *Queries
	maxBody int64
}

func (wm *WorkspaceMonitor) Project(project string) *ProjectMonitor {
	if wm == nil {
		return nil
	}
	return &ProjectMonitor{
		w:       wm,
		project: project,
	}
}

type ProjectMonitor struct {
	w       *WorkspaceMonitor
	project string
}

func (pm *ProjectMonitor) Lambda(name string) *LambdaMonitor {
	if pm == nil {
		return nil
	}
	return &LambdaMonitor{
		pm:   pm,
		name: name,
	}
}

func (pm *ProjectMonitor) Cron(expression string) *CronMonitor {
	if pm == nil {
		return nil
	}
	return &CronMonitor{
		pm:         pm,
		expression: expression,
	}
}

func (pm *ProjectMonitor) Endpoint(method, path string, handler http.Handler) http.Handler {
	if pm == nil {
		return handler
	}
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		started := time.Now()

	})
}

type CronMonitor struct {
	pm         *ProjectMonitor
	expression string
}

func (cm *CronMonitor) Started() *RunningCron {
	if cm == nil {
		return nil
	}
	return &RunningCron{
		cm:      cm,
		started: time.Now(),
	}
}

type RunningCron struct {
	cm      *CronMonitor
	started time.Time
}

func (rc *RunningCron) Finished(err error) error {
	if rc == nil {
		return nil
	}
	now := time.Now()
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	ctx, cancel := dbContext()
	defer cancel()
	return rc.cm.pm.w.q.AddCronStat(ctx, AddCronStatParams{
		Project:    rc.cm.pm.project,
		Expression: rc.cm.expression,
		StartedAt:  rc.started,
		FinishedAt: now,
		Error:      errMsg,
	})
}

type LambdaMonitor struct {
	pm   *ProjectMonitor
	name string
}

func (lm *LambdaMonitor) Started(env []string) *RunningLambda {
	if lm == nil {
		return nil
	}
	return &RunningLambda{
		lm:      lm,
		env:     env,
		started: time.Now(),
	}
}

type RunningLambda struct {
	lm      *LambdaMonitor
	env     []string
	started time.Time
}

func (rm *RunningLambda) Finished(err error) error {
	if rm == nil {
		return nil
	}
	now := time.Now()
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	data, err := json.Marshal(rm.env)
	if err != nil {
		return fmt.Errorf("encode envs: %w", err)
	}
	ctx, cancel := dbContext()
	defer cancel()
	return rm.lm.pm.w.q.AddLambdaStat(ctx, AddLambdaStatParams{
		Project:     rm.lm.pm.project,
		Name:        rm.lm.name,
		StartedAt:   rm.started,
		FinishedAt:  now,
		Environment: string(data),
		Error:       errMsg,
	})
}

func dbContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), DatabaseTimeout)
}

type requestBodyTrack struct {
	reader io.ReadCloser
	size   int64
}

func (rt *requestBodyTrack) Read(p []byte) (n int, err error) {
	//TODO implement me
	panic("implement me")
}

func (rt *requestBodyTrack) Close() error {
	//TODO implement me
	panic("implement me")
}
