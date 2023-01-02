package workspace

import (
	"github.com/reddec/trusted-cgi/application/config"
	"github.com/robfig/cron"
)

func NewCron(cfg config.Cron, calls []*Lambda, queues []*Queue) cron.Job {
	return nil // TODO
}
