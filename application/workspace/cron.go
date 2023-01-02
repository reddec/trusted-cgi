package workspace

import (
	"github.com/reddec/trusted-cgi/application/config"
	"github.com/robfig/cron"
)

func NewCron(cfg config.Cron, sync []*Sync, async []*Async) cron.Job {
	return nil // TODO
}
