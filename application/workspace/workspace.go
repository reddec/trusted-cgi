package workspace

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/reddec/trusted-cgi/application/config"
	"github.com/reddec/trusted-cgi/types"
	"github.com/robfig/cron"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const ProjectFile = "cgi.hcl"

type Config struct {
	Creds    *types.Credential
	QueueDir string
	Cache    CacheStorage
}

func New(cfg Config, dir string) (*Workspace, error) {
	list, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("list dir %s: %w", dir, err)
	}

	wp := &Workspace{
		scheduler: cron.New(),
		router:    chi.NewRouter(),
	}
	for _, entry := range list {
		if !entry.IsDir() {
			continue
		}
		file := filepath.Join(dir, entry.Name(), ProjectFile)
		pr, err := newProject(file, cfg)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			wp.Close()
			return nil, fmt.Errorf("add file %s: %w", file, err)
		}
		wp.addProject(pr)
	}
	// start cron scheduler
	wp.scheduler.Start()
	return wp, nil
}

type Workspace struct {
	scheduler *cron.Cron
	router    chi.Router
}

func (wrk *Workspace) Close() {
	wrk.scheduler.Stop()
}

func (wrk *Workspace) addProject(project *Project) {
	path := "/" + strings.ToLower(project.config.Name)
	wrk.router.Mount(path, http.StripPrefix(path, project.router))

	for _, entry := range project.scheduler.Entries() {
		wrk.scheduler.Schedule(entry.Schedule, entry.Job)
	}
}

type Project struct {
	lambdas   map[string]*Lambda
	queues    map[string]*Queue
	scheduler *cron.Cron
	router    chi.Router
	config    *config.Project
	settings  Config
}

func newProject(file string, cfg Config) (*Project, error) {
	projectConfig, err := config.ParseFile(file)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}
	pr := &Project{
		lambdas:   map[string]*Lambda{},
		queues:    map[string]*Queue{},
		scheduler: cron.New(),
		router:    chi.NewMux(),
		config:    projectConfig,
		settings:  cfg,
	}

	if err := pr.indexLambdas(); err != nil {
		return nil, fmt.Errorf("index lambdas: %w", err)
	}

	if err := pr.indexQueues(); err != nil {
		return nil, fmt.Errorf("index queues: %w", err)
	}

	if err := pr.addEndpoints(); err != nil {
		return nil, fmt.Errorf("add endpoints: %w", err)
	}

	if err := pr.addCronTabs(); err != nil {
		return nil, fmt.Errorf("add cron: %w", err)
	}
	return pr, nil
}

func (pr *Project) indexLambdas() error {
	for _, lambda := range pr.config.Lambda {
		instance, err := NewLambda(lambda, pr.settings.Creds)
		if err != nil {
			return fmt.Errorf("create lambda %s: %w", lambda.Name, err)
		}
		pr.lambdas[lambda.Name] = instance
	}
	return nil
}

func (pr *Project) indexQueues() error {
	queueDir := filepath.Join(pr.settings.QueueDir, pr.config.Name)
	for _, queue := range pr.config.Queue {
		sync, err := pr.resolveCall(queue.Call)
		if err != nil {
			return fmt.Errorf("resolve call '%s' referenced in queue '%s: %w'", queue.Call.Lambda, queue.Name, err)
		}
		instance, err := NewQueue(queueDir, queue, sync)
		if err != nil {
			return fmt.Errorf("create queue %s: %w", queue.Name, err)
		}
		pr.queues[queue.Name] = instance
	}
	return nil
}

func (pr *Project) addEndpoints() error {
	group := pr.router
	if err := pr.mountEndpoints(group, http.MethodGet, pr.config.Get, pr.settings.Cache); err != nil {
		return fmt.Errorf("add GET endpoints: %w", err)
	}
	if err := pr.mountEndpoints(group, http.MethodPost, pr.config.Post, pr.settings.Cache); err != nil {
		return fmt.Errorf("add POST endpoints: %w", err)
	}
	if err := pr.mountEndpoints(group, http.MethodPut, pr.config.Put, pr.settings.Cache); err != nil {
		return fmt.Errorf("add PUT endpoints: %w", err)
	}
	if err := pr.mountEndpoints(group, http.MethodPatch, pr.config.Patch, pr.settings.Cache); err != nil {
		return fmt.Errorf("add PATCH endpoints: %w", err)
	}
	if err := pr.mountEndpoints(group, http.MethodDelete, pr.config.Delete, pr.settings.Cache); err != nil {
		return fmt.Errorf("add DELETE endpoints: %w", err)
	}
	if pr.config.Static != "" {
		group.Handle("*", http.FileServer(http.Dir(pr.config.Static)))
	}
	return nil
}

func (pr *Project) mountEndpoints(router chi.Router, method string, endpoints []config.Endpoint, cache CacheStorage) error {
	for _, ep := range endpoints {
		queues, calls, err := pr.resolve(ep.Enqueues, ep.Calls)
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

func (pr *Project) addCronTabs() error {
	for _, cronTab := range pr.config.Cron {
		queues, calls, err := pr.resolve(cronTab.Enqueues, cronTab.Calls)
		if err != nil {
			return fmt.Errorf("resolve cron %s: %w", cronTab.Schedule, err)
		}

		err = pr.scheduler.AddJob(cronTab.Schedule, NewCron(calls, queues))
		if err != nil {
			return fmt.Errorf("add cron task %s: %w", cronTab.Schedule, err)
		}
	}
	return nil
}

func (pr *Project) resolve(toEnqueue []config.Enqueue, toCall []config.Call) ([]*Async, []*Sync, error) {
	var asyncs []*Async
	var syncs []*Sync
	// resolve queues
	for _, queue := range toEnqueue {
		q, ok := pr.queues[queue.Queue]
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
		sync, err := pr.resolveCall(lambda)
		if err != nil {
			return nil, nil, err
		}
		syncs = append(syncs, sync)
	}
	return asyncs, syncs, nil
}

func (pr *Project) resolveCall(lambda config.Call) (*Sync, error) {
	l, ok := pr.lambdas[lambda.Lambda]
	if !ok {
		return nil, fmt.Errorf("refernce to uknown lambda '%s'", lambda.Lambda)
	}
	sync, err := NewSync(lambda, l)
	if err != nil {
		return nil, fmt.Errorf("create sync link to lambda %s: %w", lambda.Lambda, err)
	}
	return sync, nil
}
