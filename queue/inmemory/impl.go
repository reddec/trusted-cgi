package inmemory

import (
	"bytes"
	"context"
	"fmt"
	"github.com/reddec/trusted-cgi/types"
	"io/ioutil"
	"sync"
	"sync/atomic"
)

type item struct {
	payload types.Request
	data    []byte
}

func (it *item) makeRequest() *types.Request {
	cp := it.payload // shallow copy
	cp.Body = ioutil.NopCloser(bytes.NewReader(it.data))
	return &cp
}

// Dummy implementation of in-memory queue with pre-allocated channel for buffering.
// It's safe to call Close several times
func New(size int) *memoryQueue {
	return &memoryQueue{
		closed: make(chan struct{}),
		stream: make(chan item, size),
	}
}

type memoryQueue struct {
	closed chan struct{}
	stream chan item
	peeked struct {
		value     item
		available bool
	}
	rlock   sync.Mutex
	closing int32
}

func (queue *memoryQueue) Put(ctx context.Context, request *types.Request) error {
	defer request.Body.Close()
	select {
	case <-queue.closed:
		return fmt.Errorf("put: queue is closed")
	default:
	}
	data, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return fmt.Errorf("put: read body: %w", err)
	}
	select {
	case <-queue.closed:
		return fmt.Errorf("put: queue is closed")
	case <-ctx.Done():
		return fmt.Errorf("put: context closed: %w", ctx.Err())
	case queue.stream <- item{
		payload: *request,
		data:    data,
	}:
		return nil
	}
}

func (queue *memoryQueue) Peek(ctx context.Context) (*types.Request, error) {
	select {
	case <-queue.closed:
		return nil, fmt.Errorf("peek: queue is closed")
	default:
	}
	queue.rlock.Lock()
	defer queue.rlock.Unlock()
	if !queue.peeked.available {
		select {
		case <-queue.closed:
			return nil, fmt.Errorf("peek: queue is closed")
		case <-ctx.Done():
			return nil, fmt.Errorf("peek: context closed: %w", ctx.Err())
		case item := <-queue.stream:
			queue.peeked.value = item
			queue.peeked.available = true
		}
	}
	return queue.peeked.value.makeRequest(), nil
}

func (queue *memoryQueue) Commit(ctx context.Context) error {
	select {
	case <-queue.closed:
		return fmt.Errorf("commit: queue is closed")
	default:
	}
	queue.rlock.Lock()
	defer queue.rlock.Unlock()
	queue.peeked.available = false
	return nil
}

func (queue *memoryQueue) Done() <-chan struct{} { return queue.closed }

func (queue *memoryQueue) Close() {
	if atomic.CompareAndSwapInt32(&queue.closing, 0, 1) {
		close(queue.closed)
		close(queue.stream)
	}
}

func (queue *memoryQueue) Destroy() error {
	queue.Close()
	return nil
}
