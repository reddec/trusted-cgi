package queuemanager

import (
	"context"
	"fmt"
	"github.com/reddec/trusted-cgi/application"
	"github.com/reddec/trusted-cgi/queue"
	"github.com/reddec/trusted-cgi/types"
	"io"
	"log"
	"os"
	"sort"
	"sync"
	"time"
)

// Store contains queues configuration for reload
type Store interface {
	// Save queues list
	SetQueues(queues []application.Queue) error
	// Load queues list
	GetQueues() ([]application.Queue, error)
}

// Minimal required platform features
type Platform interface {
	InvokeByUID(ctx context.Context, uid string, request types.Request, out io.Writer) error
}

type QueueFactory func(name string) (queue.Queue, error)

func New(ctx context.Context, config Store, platform Platform, factory QueueFactory, validator application.Validator) (*queueManager, error) {
	qm := &queueManager{
		ctx:          ctx,
		platform:     platform,
		validator:    validator,
		queues:       map[string]*queueDefinition{},
		queueFactory: factory,
		config:       config,
	}
	return qm, qm.init()
}

type queueManager struct {
	ctx          context.Context
	lock         sync.RWMutex
	platform     Platform
	validator    application.Validator
	queues       map[string]*queueDefinition
	queueFactory QueueFactory
	config       Store
	wg           sync.WaitGroup
}

func (qm *queueManager) init() error {
	list, err := qm.config.GetQueues()
	if err != nil {
		return err
	}
	for _, def := range list {
		err := qm.addQueueUnsafe(def)
		if err != nil {
			return err
		}
	}
	return nil
}

func (qm *queueManager) Put(queue string, request *types.Request) error {
	qm.lock.RLock()
	defer qm.lock.RUnlock()
	defer request.Body.Close()
	q, ok := qm.queues[queue]
	if !ok {
		return fmt.Errorf("queue %s does not exist", queue)
	}
	if err := qm.validator.Inspect(q.Target, request); err != nil {
		return fmt.Errorf("put: security validation failed: %w", err)
	}
	return q.queue.Put(qm.ctx, request)
}

func (qm *queueManager) Add(queue application.Queue) error {
	qm.lock.Lock()
	defer qm.lock.Unlock()
	err := qm.addQueueUnsafe(queue)
	if err != nil {
		return err
	}
	return qm.config.SetQueues(qm.listUnsafe())
}

func (qm *queueManager) addQueueUnsafe(queue application.Queue) error {
	if !application.QueueNameReg.MatchString(queue.Name) {
		return fmt.Errorf("invalid queue name: should be %v", application.QueueNameReg)
	}
	q, ok := qm.queues[queue.Name]
	if ok {
		return fmt.Errorf("queue %s already exists", queue.Name)
	}

	back, err := qm.queueFactory(queue.Name)
	if err != nil {
		return fmt.Errorf("add queue %s - create backend for queue: %w", queue.Name, err)
	}

	q = &queueDefinition{
		Queue:  queue,
		worker: startWorker(qm.ctx, back, queue, qm.platform, &qm.wg),
		queue:  back,
	}
	if qm.queues == nil {
		qm.queues = make(map[string]*queueDefinition)
	}
	qm.queues[queue.Name] = q
	return nil
}

func (qm *queueManager) Remove(queue string) error {
	qm.lock.Lock()
	defer qm.lock.Unlock()
	q, ok := qm.queues[queue]
	if !ok {
		return nil
	}
	q.worker.stop()
	<-q.worker.done
	delete(qm.queues, queue)
	err := q.queue.Destroy()
	if err != nil {
		return err
	}
	return qm.config.SetQueues(qm.listUnsafe())
}

func (qm *queueManager) Assign(queue string, targetLambda string) error {
	qm.lock.Lock()
	defer qm.lock.Unlock()
	q, ok := qm.queues[queue]
	if !ok {
		return fmt.Errorf("queue %s does not exist", queue)
	}
	q.worker.stop()
	<-q.worker.done
	q.Target = targetLambda
	q.worker = startWorker(qm.ctx, q.queue, q.Queue, qm.platform, &qm.wg)
	return qm.config.SetQueues(qm.listUnsafe())
}

func (qm *queueManager) List() []application.Queue {
	var ans = qm.listUnsafe()
	sort.Slice(ans, func(i, j int) bool {
		return ans[i].Name < ans[j].Name
	})
	return ans
}

func (qm *queueManager) listUnsafe() []application.Queue {
	var ans = make([]application.Queue, 0, len(qm.queues))
	for _, q := range qm.queues {
		ans = append(ans, q.Queue)
	}
	return ans
}

func (qm *queueManager) Find(targetLambda string) []application.Queue {
	var ans = make([]application.Queue, 0)
	qm.lock.RLock()
	defer qm.lock.RUnlock()
	for _, q := range qm.queues {
		if q.Target == targetLambda {
			ans = append(ans, q.Queue)
		}
	}
	return ans
}

func (qm *queueManager) Wait() {
	qm.wg.Wait()
}

type queueDefinition struct {
	application.Queue
	worker *worker
	queue  queue.Queue
}

type worker struct {
	stop func()
	done chan struct{}
}

func startWorker(gctx context.Context, queue queue.Queue, definition application.Queue, plt Platform, wg *sync.WaitGroup) *worker {
	ctx, cancel := context.WithCancel(gctx)
	w := &worker{
		stop: cancel,
		done: make(chan struct{}),
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(w.done)
		for {
			err := doTask(ctx, plt, definition, queue)
			if err != nil {
				log.Println("queues: queue", definition.Name, "failed process task:", err)
			}
			select {
			case <-ctx.Done():
				return
			default:
			}
			err = queue.Commit(ctx)
			if err != nil {
				log.Println("queues: failed commit - waiting", commitFailedDelay)
				select {
				case <-time.After(commitFailedDelay):
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return w
}

func doTask(ctx context.Context, plt Platform, definition application.Queue, queue queue.Queue) error {
	for i := 0; i <= definition.Retry; i++ {
		req, err := queue.Peek(ctx)

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			log.Println("queues: failed peek", definition.Name, ":", err)
		} else if err = plt.InvokeByUID(ctx, definition.Target, *req, os.Stderr); err != nil {
			log.Println("queues: failed invoke by uid", definition.Target, "from queue", definition.Name, ":", err)
		} else {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(definition.Interval)):
		}
	}
	return fmt.Errorf("failed to process task for queue %s after all attempts", definition.Name)
}

const commitFailedDelay = 3 * time.Second
