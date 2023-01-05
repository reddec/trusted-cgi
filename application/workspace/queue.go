package workspace

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"path/filepath"
	"time"

	"github.com/reddec/dfq"
	"github.com/reddec/trusted-cgi/application/config"
	"github.com/reddec/trusted-cgi/trace"
	"github.com/tinylib/msgp/msgp"
)

func NewQueue(project *Project, cfg *config.Queue) (*Queue, error) {
	queueDir := filepath.Join(project.QueuesDir(), cfg.Name())
	q, err := dfq.Open(queueDir)
	if err != nil {
		return nil, fmt.Errorf("open queue in %s: %w", queueDir, err)
	}
	script, err := NewScript(project, &cfg.Script)
	if err != nil {
		return nil, fmt.Errorf("create script: %w", err)
	}
	return &Queue{
		project: project,
		backend: q,
		config:  cfg,
		script:  script,
	}, nil
}

//msgp:ignore Queue
type Queue struct {
	project *Project
	script  *Script
	config  *config.Queue
	backend dfq.Queue
}

func (q *Queue) interval() time.Duration {
	interval := q.config.Interval
	if interval <= 0 {
		return time.Second
	}
	return interval
}

func (q *Queue) retries() int64 {
	retries := q.config.Retry
	if retries < 0 {
		return math.MaxInt64 - 1
	}
	return retries
}

func (q *Queue) maxMessages() int64 {
	max := q.config.Size
	if max <= 0 {
		max = math.MaxInt16
	}
	return max
}

func (q *Queue) Push(env map[string]string, data io.Reader) error {
	if q.backend.Len() > q.maxMessages() {
		return fmt.Errorf("queue is full")
	}

	return q.backend.Stream(func(out io.Writer) error {
		return NewTask(env, data).Write(out)
	})
}

func (q *Queue) Call(_ context.Context, renderCtx any, payload io.Reader) (io.ReadCloser, error) {
	env, pd, err := q.script.Render(renderCtx, payload)
	if err != nil {
		return nil, err
	}
	return io.NopCloser(bytes.NewReader([]byte(""))), q.Push(env, pd)
}

func (q *Queue) Run(ctx context.Context) error {
	for {
		var i int64
		retries := q.retries()
		for i = 0; i <= retries; i++ {
			err := q.processMessage(ctx)
			if err == nil {
				break
			}
			if errors.Is(err, context.Canceled) {
				return nil
			}
			log.Println("failed to process message in queue:", err)
			if i < retries {
				select {
				case <-time.After(q.interval()):
				case <-ctx.Done():
					return nil
				}
			}
		}
		if err := q.backend.Commit(); err != nil {
			return fmt.Errorf("commit message: %w", err)
		}
	}
}

func (q *Queue) processMessage(ctx context.Context) error {
	t, s, err := q.peek(ctx)
	if err != nil {
		return fmt.Errorf("peek message: %w", err)
	}
	defer s.Close()

	tracer := q.project.Trace()
	tracer.Set("queue", q.config.Name)
	defer tracer.Close()

	out, err := q.script.Invoke(trace.WithTrace(ctx, tracer), t.Environment, t.Body)
	if err != nil {
		tracer.Set("error", err.Error())
		return fmt.Errorf("call lambda: %w", err)
	}
	if _, err := io.Copy(io.Discard, out); err != nil {
		tracer.Set("copy_error", err.Error())
		_ = out.Close()
		return fmt.Errorf("read output: %w", err)
	}
	return out.Close()
}

func (q *Queue) peek(ctx context.Context) (*Task, io.Closer, error) {
	rw, err := q.backend.Wait(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("wait message: %w", err)
	}
	t, err := NewTaskFromStream(rw)
	if err != nil {
		_ = rw.Close()
		return nil, nil, fmt.Errorf("parse task: %w", err)
	}
	return t, rw, nil
}

type readCloser struct {
	reader io.Reader
	closer io.Closer
}

func (rc *readCloser) Read(p []byte) (n int, err error) {
	return rc.reader.Read(p)
}

func (rc *readCloser) Close() error {
	return rc.closer.Close()
}

func NewTask(env map[string]string, payload io.Reader) *Task {
	return &Task{
		Environment: env,
		Body:        payload,
	}
}

func NewTaskFromStream(stream io.Reader) (*Task, error) {
	r := msgp.NewReader(stream)
	var t Task
	if err := t.DecodeMsg(r); err != nil {
		return nil, fmt.Errorf("decode header: %w", err)
	}
	t.Body = r
	return &t, nil
}

//go:generate msgp -tests=false -marshal=false
type Task struct {
	Environment map[string]string `msg:"environment"`
	Body        io.Reader         `msg:"-"`
}

func (pl *Task) Write(out io.Writer) error {
	writer := msgp.NewWriter(out)
	if err := pl.EncodeMsg(writer); err != nil {
		return fmt.Errorf("write header: %w", err)
	}
	if pl.Body == nil {
		return writer.Flush()
	}
	if _, err := io.Copy(writer, pl.Body); err != nil {
		return fmt.Errorf("write body: %w", err)
	}
	return writer.Flush()
}
