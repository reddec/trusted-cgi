package stats

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const DatabaseTimeout = 5 * time.Second // we can not use global ctx or we can not record cancellation event

func NewMonitor(dbtx DBTX, sniffBodySize int64) *WorkspaceMonitor {
	return &WorkspaceMonitor{q: New(dbtx), maxBody: sniffBodySize}
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
		outputTrack := &responseTrack{wrap: writer, sniff: sniffer{max: pm.w.maxBody}, headers: writer.Header()}
		inputTrack := &requestBodyTrack{reader: request.Body, sniff: sniffer{max: pm.w.maxBody}}
		request.Body = inputTrack

		handler.ServeHTTP(outputTrack, request)
		finished := time.Now()
		ctx, cancel := dbContext()
		defer cancel()

		requestHeaders, err := json.Marshal(request.Header)
		if err != nil {
			log.Println("failed marshal headers:", err)
			return
		}

		responseHeaders, err := json.Marshal(outputTrack.headers)
		if err != nil {
			log.Println("failed marshal response headers:", err)
			return
		}

		err = pm.w.q.AddEndpointStat(ctx, AddEndpointStatParams{
			Project:         pm.project,
			Method:          method,
			Path:            path,
			RequestUrl:      request.RequestURI,
			StartedAt:       started,
			FinishedAt:      finished,
			RequestSize:     inputTrack.size,
			RequestHeaders:  string(requestHeaders),
			RequestBody:     inputTrack.sniff.data,
			ResponseSize:    outputTrack.size,
			ResponseHeaders: string(responseHeaders),
			ResponseBody:    outputTrack.sniff.data,
			Status:          int64(outputTrack.status),
		})
		if err != nil {
			log.Println("failed save stats:", err)
		}
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
	sniff  sniffer
}

func (rt *requestBodyTrack) Read(p []byte) (n int, err error) {
	n, err = rt.reader.Read(p)
	rt.size += int64(n)
	rt.sniff.Add(n, p)
	return
}

func (rt *requestBodyTrack) Close() error {
	return rt.reader.Close()
}

type responseTrack struct {
	wrap    http.ResponseWriter
	status  int
	size    int64
	headers http.Header
	sniff   sniffer
}

func (rt *responseTrack) Header() http.Header {
	return rt.headers
}

func (rt *responseTrack) Write(bytes []byte) (int, error) {
	n, err := rt.wrap.Write(bytes)
	if rt.status == 0 {
		rt.status = http.StatusOK
	}
	rt.size += int64(n)
	rt.sniff.Add(n, bytes)
	return n, err
}

func (rt *responseTrack) WriteHeader(statusCode int) {
	if rt.status == 0 {
		rt.status = statusCode
	}
	rt.wrap.WriteHeader(statusCode)
}

type sniffer struct {
	max  int64
	data []byte
}

func (s *sniffer) Add(n int, data []byte) {
	if s.max <= 0 {
		return
	}
	l := int64(n)
	if l > s.max {
		l = s.max
	}
	s.max -= l
	s.data = append(s.data, data[:l]...)
}
