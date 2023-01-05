package workspace

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/go-chi/chi/v5"
	"github.com/reddec/trusted-cgi/internal"
	"github.com/reddec/trusted-cgi/trace"
	"github.com/reddec/trusted-cgi/types"
	"github.com/robfig/cron"
)

var ProjectFiles = []string{
	"cgi.yaml",
	"cgi.yml",
	"cgi.json",
}

type Config struct {
	Creds     *types.Credential // optional, which credentials to use for executing (su)
	QueueDir  string            // optional, default is 'queues'. Place where to store queues
	SniffSize int64             // optional, max size of input and output to be stored in stats
}

func New(cfg Config, dir string) (*Workspace, error) {
	list, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("list dir %s: %w", dir, err)
	}

	if cfg.QueueDir == "" {
		cfg.QueueDir = "queues"
	}

	wp := &Workspace{
		settings:  &cfg,
		scheduler: cron.New(),
		router:    chi.NewRouter(),
		queues:    internal.NewDaemonSet(true),
	}
	for _, entry := range list {
		if !entry.IsDir() {
			continue
		}
		for _, projectFile := range ProjectFiles {
			file := filepath.Join(dir, entry.Name(), projectFile)
			if stat, err := os.Stat(file); err != nil || stat.IsDir() {
				continue
			}
			_, err := NewProject(wp, file)
			if err != nil {
				return nil, fmt.Errorf("add file %s: %w", file, err)
			}
			break
		}
	}

	return wp, nil
}

type Workspace struct {
	scheduler *cron.Cron
	router    chi.Router
	queues    *internal.DaemonSet
	creds     *types.Credential
	settings  *Config
	traces    trace.TraceStorage
}

func (wrk *Workspace) Trace() *trace.Trace {
	//TODO: workspace name
	tp := trace.NewTrace(wrk.traces)
	tp.Set("workspace", "default")
	return tp
}

func (wrk *Workspace) QueuesDir() string {
	return wrk.settings.QueueDir
}

func (wrk *Workspace) SniffSize() int64 {
	return wrk.settings.SniffSize
}

func (wrk *Workspace) Run(ctx context.Context) error {
	wrk.scheduler.Start()
	defer wrk.scheduler.Stop()
	return wrk.queues.Run(ctx)
}

func (wrk *Workspace) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	wrk.router.ServeHTTP(writer, request)
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
	for global.Err() == nil {
		current := r.workspace.Load()
		err := r.runCurrent(global, current)
		if err == nil || errors.Is(err, context.Canceled) {
			continue
		}
		return err
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
