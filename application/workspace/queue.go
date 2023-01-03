package workspace

import (
	"context"
	"errors"
	"fmt"
	"github.com/reddec/dfq"
	"github.com/reddec/trusted-cgi/application/config"
	"github.com/tinylib/msgp/msgp"
	"io"
	"log"
	"math"
	"path"
	"path/filepath"
	"time"
)

func NewQueue(rootDir string, cfg config.Queue, sync *Sync) (*Queue, error) {
	retries := cfg.Retry
	if retries < 0 {
		retries = math.MaxInt64 - 1
	}
	interval := cfg.Interval.Duration()
	if interval <= 0 {
		interval = time.Second
	}
	max := cfg.Size
	if max <= 0 {
		max = math.MaxInt16
	}
	queueDir := filepath.Join(rootDir, path.Clean(cfg.Name))
	q, err := dfq.Open(queueDir)
	if err != nil {
		return nil, fmt.Errorf("open queue in %s: %w", queueDir, err)
	}

	return &Queue{
		backend:  q,
		call:     sync,
		interval: interval,
		retries:  retries,
		max:      max,
	}, nil
}

//go:generate msgp
type Message struct {
	Environment map[string]string `msg:"env"`
	Payload     io.ReadCloser     `msg:"-"`
}

//msgp:ignore Queue
type Queue struct {
	backend  dfq.Queue
	call     *Sync
	interval time.Duration
	retries  int64
	max      int64
}

func (q *Queue) Push(env map[string]string, data io.Reader) error {
	if q.backend.Len() > q.max {
		return fmt.Errorf("queue is full")
	}
	msg := Message{
		Environment: env,
	}
	return q.backend.Stream(func(out io.Writer) error {
		w := msgp.NewWriter(out)
		if err := msg.EncodeMsg(w); err != nil {
			return fmt.Errorf("encode header: %w", err)
		}
		if _, err := io.Copy(w, data); err != nil {
			return fmt.Errorf("write data: %w", err)
		}
		return w.Flush()
	})
}

func (q *Queue) Run(ctx context.Context) error {
	for {
		var i int64
		for i = 0; i <= q.retries; i++ {
			err := q.processMessage(ctx)
			if err == nil {
				break
			}
			if errors.Is(err, context.Canceled) {
				return nil
			}
			log.Println("failed to process message in queue:", err)
			if i < q.retries {
				select {
				case <-time.After(q.interval):
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
	m, err := q.peek(ctx)
	if err != nil {
		return fmt.Errorf("peek message: %w", err)
	}
	defer m.Payload.Close()
	out, err := q.call.Call(ctx, m.Environment, m.Payload, emptyContext)
	if err != nil {
		return fmt.Errorf("call lambda: %w", err)
	}
	if _, err := io.Copy(io.Discard, out); err != nil {
		_ = out.Close()
		return fmt.Errorf("read output: %w", err)
	}
	return out.Close()
}

func (q *Queue) peek(ctx context.Context) (*Message, error) {
	rw, err := q.backend.Wait(ctx)
	if err != nil {
		return nil, fmt.Errorf("wait message: %w", err)
	}
	r := msgp.NewReader(rw)
	var msg = &Message{
		Payload: &readCloser{
			reader: r,
			closer: rw,
		},
	}
	if err := msg.DecodeMsg(r); err != nil {
		return nil, fmt.Errorf("decode header: %w", err)
	}
	return msg, nil
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
