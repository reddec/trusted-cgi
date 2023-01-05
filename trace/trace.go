package trace

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type TraceStorage interface {
	Store(started, stopped time.Time, entity, name string, meta map[string]any) error
}

type Trace struct {
	parent   *Trace
	storage  TraceStorage
	entity   string
	name     string
	started  time.Time
	complete int32
	meta     map[string]any
	metaLock sync.RWMutex
}

func (t *Trace) Close() {
	if !atomic.CompareAndSwapInt32(&t.complete, 0, 1) {
		return
	}
	stopped := time.Now()
	t.metaLock.RLock()
	defer t.metaLock.RUnlock()

	if err := t.getStorage().Store(t.started, stopped, t.entity, t.name, t.meta); err != nil {
		log.Printf("failed store trace %s: %v", t.name, err)
	}
}

func (t *Trace) Set(key string, value any) {
	t.metaLock.Lock()
	defer t.metaLock.Unlock()
	t.meta[key] = value
}

func (t *Trace) getStorage() TraceStorage {
	if t.storage != nil {
		return t.storage
	}
	if t.parent != nil {
		return t.parent.getStorage()
	}
	return &LoggingStorage{}
}

func NewTrace(storage TraceStorage) *Trace {
	return &Trace{started: time.Now(), meta: map[string]any{}, storage: storage}
}

type traceContextKey struct{}

func NewTraceFromContext(ctx context.Context) *Trace {
	v := ctx.Value(traceContextKey{})
	if root, ok := v.(*Trace); ok {
		return root
	}
	return NewTrace(nil)
}

func WithTrace(ctx context.Context, trace *Trace) context.Context {
	return context.WithValue(ctx, traceContextKey{}, trace)
}

type LoggingStorage struct{}

func (ls LoggingStorage) Store(started, stopped time.Time, entity, name string, meta map[string]any) error {
	log.Printf("[%s:%s] started: %v, finished: %v, duration: %v, meta: %+v", entity, name, started, stopped, stopped.Sub(started), meta)
	return nil
}
