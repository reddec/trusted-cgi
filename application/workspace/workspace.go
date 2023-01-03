package workspace

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/reddec/trusted-cgi/internal"
	"github.com/reddec/trusted-cgi/types"
	"github.com/robfig/cron"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
)

const ProjectFile = "cgi.hcl"

type Config struct {
	Creds    *types.Credential // optional, which credentials to use for executing (su)
	QueueDir string            // optional, default is 'queues'. Place where to store queues
	CacheDir string            // optional, default is 'cache'. Place where to store requests payloads (cache).
}

func New(cfg Config, dir string) (*Workspace, error) {
	list, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("list dir %s: %w", dir, err)
	}

	if cfg.QueueDir == "" {
		cfg.QueueDir = "queues"
	}
	if cfg.CacheDir == "" {
		cfg.CacheDir = "cache"
	}

	cache, err := NewFileCache(cfg.CacheDir)
	if err != nil {
		return nil, fmt.Errorf("create cache in %s: %w", cfg.CacheDir, err)
	}

	wp := &Workspace{
		scheduler: cron.New(),
		router:    chi.NewRouter(),
		queues:    internal.NewDaemonSet(true),
	}
	for _, entry := range list {
		if !entry.IsDir() {
			continue
		}
		file := filepath.Join(dir, entry.Name(), ProjectFile)
		pr, err := newProject(file, cfg, cache)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("add file %s: %w", file, err)
		}
		wp.addProject(pr)
	}

	return wp, nil
}

type Workspace struct {
	scheduler *cron.Cron
	router    chi.Router
	queues    *internal.DaemonSet
}

func (wrk *Workspace) Run(ctx context.Context) error {
	wrk.scheduler.Start()
	defer wrk.scheduler.Stop()
	return wrk.queues.Run(ctx)
}

func (wrk *Workspace) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	wrk.router.ServeHTTP(writer, request)
}

func (wrk *Workspace) addProject(project *Project) {
	path := "/" + strings.ToLower(project.config.Name)
	wrk.router.Mount(path, http.StripPrefix(path, project.router))

	for _, entry := range project.scheduler.Entries() {
		wrk.scheduler.Schedule(entry.Schedule, entry.Job)
	}

	wrk.queues.Add(project.queuesDaemons.Jobs()...)
}

func NewReloadable(cfg Config, rootDir string) (*ReloadableWorkspace, error) {
	wrk, err := New(cfg, rootDir)
	if err != nil {
		return nil, err
	}
	rwrk := &ReloadableWorkspace{
		config:  cfg,
		rootDir: rootDir,
		restart: make(chan struct{}, 1),
	}
	rwrk.workspace.Store(wrk)
	return rwrk, nil
}

type ReloadableWorkspace struct {
	workspace atomic.Pointer[Workspace]
	config    Config
	rootDir   string
	restart   chan struct{}
}

func (r *ReloadableWorkspace) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	r.workspace.Load().ServeHTTP(writer, request)
}

func (r *ReloadableWorkspace) Reload() error {
	wrk, err := New(r.config, r.rootDir)
	if err != nil {
		return err
	}
	r.workspace.Store(wrk)
	// notify
	select {
	case r.restart <- struct{}{}:
	default:
	}
	return nil
}

func (r *ReloadableWorkspace) Run(global context.Context) error {
	for global.Err() != nil {
		current := r.workspace.Load()
		err := r.runCurrent(global, current)
		if err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
	}
	return nil
}

func (r *ReloadableWorkspace) runCurrent(global context.Context, current *Workspace) error {
	ctx, cancel := context.WithCancel(global)
	defer cancel()

	go func() {
		defer cancel()

		select {
		case <-ctx.Done():
		case <-r.restart:
		}
	}()

	return current.Run(ctx)
}
