package main

import (
	"context"
	"github.com/jessevdk/go-flags"
	"github.com/reddec/trusted-cgi/api/services"
	"github.com/reddec/trusted-cgi/application"
	"github.com/reddec/trusted-cgi/application/cases"
	"github.com/reddec/trusted-cgi/application/platform"
	"github.com/reddec/trusted-cgi/cmd/internal"
	internal2 "github.com/reddec/trusted-cgi/internal"
	"github.com/reddec/trusted-cgi/server"
	"github.com/reddec/trusted-cgi/stats/impl/memlog"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const version = "dev"

type Config struct {
	HttpServer
	Config    string `short:"c" long:"config" env:"CONFIG" description:"Location of server configuration" default:"server.json"`
	Dir       string `short:"d" long:"dir" env:"DIR" description:"Project directory" default:"."`
	Templates string `long:"templates" env:"TEMPLATES" description:"Templates directory" default:".templates"`
	//
	InitialAdminPassword string        `long:"initial-admin-password" env:"INITIAL_ADMIN_PASSWORD" description:"Initial admin password" default:"admin"`
	InitialChrootUser    string        `long:"initial-chroot-user" env:"INITIAL_CHROOT_USER" description:"Initial user for service" default:""`
	DisableChroot        bool          `long:"disable-chroot" env:"DISABLE_CHROOT" description:"Disable use different user for spawn"`
	SSHKey               string        `long:"ssh-key" env:"SSH_KEY" description:"Path to ssh key. If not empty and not exists - it will be generated" default:".id_rsa"`
	Dev                  bool          `long:"dev" env:"DEV" description:"Enabled dev mode (disables chroot)"`
	StatsCache           uint          `long:"stats-cache" env:"STATS_CACHE" description:"Maximum cache for stats" default:"8192"`
	StatsFile            string        `long:"stats-file" env:"STATS_FILE" description:"Binary file for statistics dump" default:".stats"`
	StatsInterval        time.Duration `long:"stats-interval" env:"STATS_INTERVAL" description:"Interval for dumping stats to file" default:"30s"`
	SchedulerInterval    time.Duration `long:"scheduler-interval" env:"SCHEDULER_INTERVAL" description:"Interval to check cron records" default:"30s"`
}

type HttpServer struct {
	GracefulShutdown time.Duration `long:"graceful-shutdown" env:"GRACEFUL_SHUTDOWN" description:"Interval before server shutdown" default:"15s" json:"graceful_shutdown"`
	Bind             string        `long:"bind" env:"BIND" description:"Address to where bind HTTP server" default:"127.0.0.1:3434" json:"bind"`
	TLS              bool          `long:"tls" env:"TLS" description:"Enable HTTPS serving with TLS" json:"tls"`
	CertFile         string        `long:"cert-file" env:"CERT_FILE" description:"Path to certificate for TLS" default:"server.crt" json:"crt_file"`
	KeyFile          string        `long:"key-file" env:"KEY_FILE" description:"Path to private key for TLS" default:"server.key" json:"key_file"`
}

func (qs *HttpServer) Serve(globalCtx context.Context, handler http.Handler) error {

	srv := http.Server{
		Addr:    qs.Bind,
		Handler: handler,
	}

	go func() {
		<-globalCtx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), qs.GracefulShutdown)
		defer cancel()
		srv.Shutdown(ctx)
	}()
	log.Println("REST server is on", qs.Bind)
	if qs.TLS {
		return srv.ListenAndServeTLS(qs.CertFile, qs.KeyFile)
	}
	return srv.ListenAndServe()
}

func main() {
	var config Config
	parser := flags.NewParser(&config, flags.Default)
	parser.LongDescription = "Easy CGI-like server for development\nAuthor: Baryshnikov Aleksandr <dev@baryshnikov.net>\nVersion: " + version
	_, err := parser.Parse()
	if err != nil {
		os.Exit(1)
	}

	gctx, closer := internal.SignalContext()
	defer closer()
	err = run(gctx, config)
	if err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, config Config) error {
	tracker, err := memlog.NewDumped(config.StatsFile, config.StatsCache)
	if err != nil {
		return err
	}

	var defCfg application.Config
	defCfg.User = config.InitialChrootUser

	if config.Dev {
		log.Println("Warning! Development mode enabled")
		defCfg.User = ""
	} else if config.DisableChroot {
		defCfg.User = ""
	}

	basePlatform, err := platform.New(filepath.Join(config.Dir, internal2.ProjectManifest))
	if err != nil {
		return err
	}

	useCases, err := cases.New(basePlatform, config.Dir, config.Templates)
	if err != nil {
		return err
	}

	if config.SSHKey != "" {
		err = useCases.SetOrCreatePrivateSSHKeyFile(config.SSHKey)
		if err != nil {
			return err
		}
	}

	projectApi := services.NewProjectSrv(useCases, tracker)
	lambdaApi := services.NewLambdaSrv(useCases, tracker)
	userApi, err := services.CreateUserSrv(config.Config, config.InitialAdminPassword)
	if err != nil {
		return err
	}

	go runScheduler(ctx, config.SchedulerInterval, useCases)

	defer tracker.Dump()
	go dumpTracker(ctx, config.StatsInterval, tracker)

	handler, err := server.Handler(ctx, config.Dev, basePlatform, tracker, userApi, projectApi, lambdaApi, userApi)
	if err != nil {
		return err
	}
	log.Println("running on", config.Bind)
	return config.Serve(ctx, handler)
}

func dumpTracker(ctx context.Context, each time.Duration, tracker interface {
	Dump() error
}) {
	t := time.NewTicker(each)
	defer t.Stop()
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
