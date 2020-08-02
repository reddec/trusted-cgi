package trustedcgi

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/reddec/trusted-cgi/api/services"
	"github.com/reddec/trusted-cgi/application"
	"github.com/reddec/trusted-cgi/application/cases"
	"github.com/reddec/trusted-cgi/application/platform"
	"github.com/reddec/trusted-cgi/application/policy"
	"github.com/reddec/trusted-cgi/application/queuemanager"
	"github.com/reddec/trusted-cgi/queue"
	"github.com/reddec/trusted-cgi/queue/indir"
	"github.com/reddec/trusted-cgi/server"
	"github.com/reddec/trusted-cgi/stats/impl/memlog"
)

const (
	defPoliciesFile         = "policies.json"
	defQueuesFile           = "queues.json"
	defServerFile           = "server.json"
	defProjectFile          = "project.json"
	defStatsFile            = ".stats"
	defTemplatesDir         = ".templates"
	defQueuesDir            = ".queues"
	defSshKey               = ".id_rsa"
	defGracefulShutdown     = 10 * time.Second // time to wait for HTTP connections shutdown (if ListenAndServe were used)
	defCfgPassword          = "admin"
	defCfgStatsDepth        = 8192
	defCfgDumpInterval      = 30 * time.Second
	defCfgSchedulerInterval = 30 * time.Second
)

// Creates default parameters for trusted-cgi instance.
func Default() *Config {
	return &Config{
		dir:               ".",
		password:          defCfgPassword,
		statsDepth:        defCfgStatsDepth,
		dumpInterval:      defCfgDumpInterval,
		schedulerInterval: defCfgSchedulerInterval,
		ssh:               true,
	}
}

// Config description for new trusted-cgi instance.
type Config struct {
	ctx               context.Context
	password          string
	statsDepth        uint
	dumpInterval      time.Duration
	schedulerInterval time.Duration
	dir               string
	ssh               bool
}

// Directory for project files.
func (cfg *Config) Directory(dir string) *Config {
	cfg.dir = dir
	return cfg
}

// Parent context.
func (cfg *Config) Context(ctx context.Context) *Config {
	cfg.ctx = ctx
	return cfg
}

// Password for admin account if not yet set.
func (cfg *Config) Password(password string) *Config {
	cfg.password = password
	return cfg
}

// SSH support enable or disable. By default - enabled.
func (cfg *Config) SSH(enable bool) *Config {
	cfg.ssh = enable
	return cfg
}

// New instance of trusted-cgi using defaults storages and implementations.
// Also initializes SSH key (if enabled). Starts supporting go-routines that will be stopped when context will be canceled.
// The Done() channel can be used to determinate sub-routine termination.
// Global context could be nil - the Background will be used.
func (cfg *Config) New() (*Instance, error) {
	globalContext := cfg.ctx
	if globalContext == nil {
		globalContext = context.Background()
	}

	policies, err := policy.New(policy.FileConfig(filepath.Join(cfg.dir, defPoliciesFile)))
	if err != nil {
		return nil, fmt.Errorf("initialiaze policies: %w", err)
	}

	basePlatform, err := platform.New(filepath.Join(cfg.dir, defProjectFile))
	if err != nil {
		return nil, fmt.Errorf("initialize base platform: %w", err)
	}

	queueFactory := func(name string) (queue.Queue, error) {
		return indir.New(filepath.Join(cfg.dir, defQueuesDir, name))
	}

	ctx, cancel := context.WithCancel(globalContext)

	queueManager, err := queuemanager.New(ctx, queuemanager.FileConfig(filepath.Join(cfg.dir, defQueuesFile)), basePlatform, queueFactory)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("initialize queues: %w", err)
	}

	useCases, err := cases.New(basePlatform, queueManager, policies, cfg.dir, filepath.Join(cfg.dir, defTemplatesDir))
	if err != nil {
		cancel()
		return nil, fmt.Errorf("initialize use-cases: %w", err)
	}

	if cfg.ssh {
		err = useCases.SetOrCreatePrivateSSHKeyFile(filepath.Join(cfg.dir, defSshKey))
		if err != nil {
			cancel()
			return nil, fmt.Errorf("initialize SSH key: %w", err)
		}
	}

	tracker, err := memlog.NewDumped(filepath.Join(cfg.dir, defStatsFile), cfg.statsDepth)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("initalize stats: %w", err)
	}

	projectApi := services.NewProjectSrv(useCases, tracker)
	lambdaApi := services.NewLambdaSrv(useCases, tracker)
	queuesApi := services.NewQueuesSrv(queueManager)
	policiesApi := services.NewPoliciesSrv(policies)
	userApi, err := services.CreateUserSrv(filepath.Join(cfg.dir, defServerFile), cfg.password)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("initialize admin API (user): %w", err)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		dumpTracker(ctx, cfg.dumpInterval, tracker)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		runScheduler(ctx, cfg.schedulerInterval, useCases)
	}()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	srv := &server.Server{
		Policies:     policies,
		Platform:     basePlatform,
		Cases:        useCases,
		Queues:       queueManager,
		Tracker:      tracker,
		TokenHandler: userApi,
		ProjectAPI:   projectApi,
		LambdaAPI:    lambdaApi,
		UserAPI:      userApi,
		QueuesAPI:    queuesApi,
		PoliciesAPI:  policiesApi,
	}
	return &Instance{
		Location: cfg.dir,
		server:   srv,
		ctx:      ctx,
		done:     done,
		cancel:   cancel,
	}, nil
}

type Instance struct {
	Location string         // location as-is it used during initialization
	server   *server.Server // initialize server with all dependencies
	ctx      context.Context
	cancel   func()
	done     chan struct{}
}

// Cancel underlying context and waits for finish.
func (instance *Instance) Stop() {
	instance.cancel()
	<-instance.done
}

// Returns channel that will be closed once all sub-routine (tracker dump and scheduler) finished.
func (instance *Instance) Done() <-chan struct{} {
	return instance.done
}

// Creates (every time new) server handlers (see Server::Handlers) using local (cancelable) context.
func (instance *Instance) Handler() http.Handler {
	return instance.server.Handler(instance.ctx)
}

// Server initialized with all dependencies.
func (instance *Instance) Server() *server.Server {
	return instance.server
}

// Context used for control instance lifecycle.
func (instance *Instance) Context() context.Context {
	return instance.ctx
}

// Listen and serves using provided binding - simple wrapper around Handler and http.Listen.
// Will shutdown in case of context cancel.
func (instance *Instance) ListenAndServe(binding string) error {
	srv := http.Server{
		Addr:    binding,
		Handler: instance.Handler(),
	}

	go func() {
		<-instance.ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), defGracefulShutdown)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()
	return srv.ListenAndServe()
}

// Listen and serves using provided binding using TLS - simple wrapper around Handler and http.ListenTLS.
// Will shutdown in case of context cancel.
func (instance *Instance) ListenAndServeTLS(binding string, certFile, keyFile string) error {
	srv := http.Server{
		Addr:    binding,
		Handler: instance.Handler(),
	}

	go func() {
		<-instance.ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), defGracefulShutdown)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()
	return srv.ListenAndServeTLS(certFile, keyFile)
}

func dumpTracker(ctx context.Context, each time.Duration, tracker interface {
	Dump() error
}) {
	t := time.NewTicker(each)
	defer t.Stop()
	defer tracker.Dump()
	for {
		select {
		case <-t.C:
		case <-ctx.Done():
			return
		}
		err := tracker.Dump()
		if err != nil {
			log.Println("[ERROR] failed to dump statistics:", err)
		}
	}
}

func runScheduler(ctx context.Context, each time.Duration, runner application.Cases) {
	t := time.NewTicker(each)
	defer t.Stop()
	for {
		select {
		case <-t.C:
		case <-ctx.Done():
			return
		}
		runner.RunScheduledActions(ctx)
	}
}
