package queues

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

type platform interface {
	FindByUID(uid string) (*application.Definition, error)
	Invoke(ctx context.Context, lambda application.Lambda, request types.Request, out io.Writer) error
}

type QueueFactory func(name string) (queue.Queue, error)

func New(ctx context.Context, platform platform, factory QueueFactory) *queueManager {
	return &queueManager{
		ctx:          ctx,
		platform:     platform,
		queues:       map[string]*queueDefinition{},
		queueFactory: factory,
	}
}

type queueManager struct {
	ctx          context.Context
	lock         sync.RWMutex
	platform     platform
	queues       map[string]*queueDefinition
	queueFactory QueueFactory
	wg           sync.WaitGroup
}

func (qm *queueManager) Put(queue string, request *types.Request) error {
	qm.lock.RLock()
	defer qm.lock.RUnlock()
	defer request.Body.Close()
	q, ok := qm.queues[queue]
	if !ok {
		return fmt.Errorf("queue %s does not exist", queue)
	}
	return q.queue.Put(qm.ctx, request)
}

func (qm *queueManager) Add(queue application.Queue) error {
	if !application.QueueNameReg.MatchString(queue.Name) {
		return fmt.Errorf("invalid queue name: should be %v", application.QueueNameReg)
	}
	qm.lock.Lock()
	defer qm.lock.Unlock()
	q, ok := qm.queues[queue.Name]
	if ok {
		return fmt.Errorf("queue %s already exists", queue.Name)
	}

	lambda, err := qm.platform.FindByUID(queue.Target)
	if err != nil {
		return err
	}

	back, err := qm.queueFactory(queue.Name)
	if err != nil {
		return err
	}

	q = &queueDefinition{
		Queue:  application.Queue{},
		worker: startWorker(qm.ctx, back, lambda.Lambda, qm.platform, &qm.wg),
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
	return q.queue.Destroy()
}

func (qm *queueManager) Assign(queue string, targetLambda string) error {
	lambda, err := qm.platform.FindByUID(targetLambda)
	if err != nil {
		return err
	}
	qm.lock.Lock()
	defer qm.lock.Unlock()
	q, ok := qm.queues[queue]
	if !ok {
		return fmt.Errorf("queue %s does not exist", queue)
	}
	q.worker.stop()
	<-q.worker.done
	q.Target = targetLambda
	q.worker = startWorker(qm.ctx, q.queue, lambda.Lambda, qm.platform, &qm.wg)
	return nil
}

func (qm *queueManager) List() []application.Queue {
	var ans = make([]application.Queue, 0, len(qm.queues))
	qm.lock.RLock()
	for _, q := range qm.queues {
		ans = append(ans, q.Queue)
	}
	qm.lock.RUnlock()
	sort.Slice(ans, func(i, j int) bool {
		return ans[i].Name < ans[j].Name
	})
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

func startWorker(gctx context.Context, queue queue.Queue, lambda application.Lambda, plt platform, wg *sync.WaitGroup) *worker {
	ctx, cancel := context.WithCancel(gctx)
	w := &worker{
		stop: cancel,
		done: make(chan struct{}),
	}
	go func() {
		defer wg.Done()
		defer close(w.done)
		for {
			req, err := queue.Peek(ctx)

			select {
			case <-ctx.Done():
				return
			default:
			}

			if err != nil {
				log.Println("queues: failed peek")
			} else if err = plt.Invoke(ctx, lambda, *req, os.Stderr); err != nil {
				log.Println("queues: failed invoke:", err)
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

const commitFailedDelay = 3 * time.Second
