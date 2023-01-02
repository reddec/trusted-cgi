package workspace

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/reddec/trusted-cgi/application/config"
	"github.com/reddec/trusted-cgi/types"
	"github.com/robfig/cron"
	"net/http"
	"strings"
)

type Config struct {
	Creds    *types.Credential
	QueueDir string
	Cache    CacheStorage
}

func Load(cfg Config, files ...string) (*Workspace, error) {
	wp := &Workspace{
		lambdas:   map[string]*Lambda{},
		queues:    map[string]*Queue{},
		scheduler: cron.New(),
		router:    chi.NewRouter(),
	}
	for _, file := range files {
		if err := wp.addFile(cfg, file); err != nil {
			wp.Close()
			return nil, fmt.Errorf("add file %s: %w", file, err)
		}
	}
	// cleanup references - everything already linked and resolved
	wp.lambdas = nil
	wp.queues = nil
	// start cron scheduler
	wp.scheduler.Start()
	return wp, nil
}

type Workspace struct {
	lambdas   map[string]*Lambda
	queues    map[string]*Queue
	scheduler *cron.Cron
	router    chi.Router
}

func (wrk *Workspace) Close() {
	wrk.scheduler.Stop()
}

func (wrk *Workspace) addFile(cfg Config, file string) error {
	project, err := config.ParseFile(file)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}
	// load lambdas
	for _, lambda := range project.Lambda {
		instance, err := NewLambda(lambda, cfg.Creds)
		if err != nil {
			return fmt.Errorf("create lambda %s: %w", lambda.Name, err)
		}
		wrk.lambdas[lambda.Name] = instance
	}
	// load queues
	for _, queue := range project.Queue {
		instance, err := NewQueue(cfg.QueueDir, queue)
		if err != nil {
			return fmt.Errorf("create queue %s: %w", queue.Name, err)
		}
		wrk.queues[queue.Name] = instance
	}
	// add endpoints
	group := chi.NewMux()
	if err := wrk.addEndpoints(group, http.MethodGet, project.Get, cfg.Cache); err != nil {
		return fmt.Errorf("add GET endpoints: %w", err)
	}
	if err := wrk.addEndpoints(group, http.MethodPost, project.Post, cfg.Cache); err != nil {
		return fmt.Errorf("add POST endpoints: %w", err)
	}
	if err := wrk.addEndpoints(group, http.MethodPut, project.Put, cfg.Cache); err != nil {
		return fmt.Errorf("add PUT endpoints: %w", err)
	}
	if err := wrk.addEndpoints(group, http.MethodPatch, project.Patch, cfg.Cache); err != nil {
		return fmt.Errorf("add PATCH endpoints: %w", err)
	}
	if err := wrk.addEndpoints(group, http.MethodDelete, project.Delete, cfg.Cache); err != nil {
		return fmt.Errorf("add DELETE endpoints: %w", err)
	}
	if project.Static != "" {
		group.Handle("*", http.FileServer(http.Dir(project.Static)))
	}
	path := "/" + strings.ToLower(project.Name)
	wrk.router.Mount(path, http.StripPrefix(path, group))

	// add cron tasks
	for _, cronTab := range project.Cron {
		queues, calls, err := wrk.resolve(cronTab.Enqueues, cronTab.Calls)
		if err != nil {
			return fmt.Errorf("resolve cron %s: %w", cronTab.Schedule, err)
		}

		err = wrk.scheduler.AddJob(cronTab.Schedule, NewCron(cronTab, calls, queues))
		if err != nil {
			return fmt.Errorf("add cron task %s: %w", cronTab.Schedule, err)
		}
	}

	return nil
}

func (wrk *Workspace) addEndpoints(router chi.Router, method string, endpoints []config.Endpoint, cache CacheStorage) error {
	for _, ep := range endpoints {
		queues, calls, err := wrk.resolve(ep.Enqueues, ep.Calls)
		if err != nil {
			return fmt.Errorf("resolve endpoint %s %s: %w", method, ep.Path, err)
		}

		handler, err := NewEndpoint(ep, cache, calls, queues)
		if err != nil {
			return fmt.Errorf("create endpoint %s %s: %w", method, ep.Path, err)
		}
		router.Method(method, ep.Path, handler)
	}
	return nil
}

func (wrk *Workspace) resolve(toEnqueue []config.Enqueue, toCall []config.Call) ([]*Async, []*Sync, error) {
	var asyncs []*Async
	var syncs []*Sync
	// resolve queues
	for _, queue := range toEnqueue {
		q, ok := wrk.queues[queue.Queue]
		if !ok {
			return nil, nil, fmt.Errorf("refernce to uknown queue '%s'", queue.Queue)
		}
		async, err := NewAsync(queue, q)
		if err != nil {
			return nil, nil, fmt.Errorf("create async link to queue %s: %w", queue.Queue, err)
		}
		asyncs = append(asyncs, async)
	}
	// resolve lambdas (sync call)
	for _, lambda := range toCall {
		l, ok := wrk.lambdas[lambda.Lambda]
		if !ok {
			return nil, nil, fmt.Errorf("refernce to uknown lambda '%s'", lambda.Lambda)
		}
		sync, err := NewSync(lambda, l)
		if err != nil {
			return nil, nil, fmt.Errorf("create sync link to lambda %s: %w", lambda.Lambda, err)
		}
		syncs = append(syncs, sync)
	}
	return asyncs, syncs, nil
}
