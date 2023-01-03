package config

import (
	"fmt"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"path/filepath"
	"regexp"
	"time"
)

var QueueNameReg = regexp.MustCompile(`^[a-z0-9A-Z-]{3,64}$`)

type Lambda struct {
	Name    string   `hcl:"name,label"`
	Exec    []string `hcl:"exec"`
	Timeout Seconds  `hcl:"timeout,optional"`
	WorkDir string   `hcl:"workDir,optional"`
}

type Invoke struct {
	Payload     *string           `hcl:"payload,optional"`
	Environment map[string]string `hcl:"environment,optional"`
}

type Call struct {
	Lambda string `hcl:"lambda,label"`
	Invoke `hcl:",remain"`
}

type Enqueue struct {
	Queue  string `hcl:"queue,label"`
	Invoke `hcl:",remain"`
}

type Cron struct {
	Schedule string    `hcl:"schedule,label"`
	Enqueues []Enqueue `hcl:"enqueue,block"`
	Calls    []Call    `hcl:"call,block"`
}

type Queue struct {
	Name     string  `hcl:"name,label"`
	Call     Call    `hcl:"call,block"` // reserve for future extension: multiple calls, enqueue, etc
	Interval Seconds `hcl:"interval,optional"`
	Size     int64   `hcl:"size,optional"`
	Retry    int64   `hcl:"retry,optional"`
}

type Endpoint struct {
	Path     string            `hcl:"path,label"`
	Body     int64             `hcl:"body,optional"`
	Status   int               `hcl:"status,optional"`
	Headers  map[string]string `hcl:"headers,optional"`
	Enqueues []Enqueue         `hcl:"enqueue,block"`
	Calls    []Call            `hcl:"call,block"`
}

type Project struct {
	Name   string     // to be filled by directory name
	Static string     `hcl:"static,optional"`
	Get    []Endpoint `hcl:"get,block"`
	Post   []Endpoint `hcl:"post,block"`
	Patch  []Endpoint `hcl:"patch,block"`
	Delete []Endpoint `hcl:"delete,block"`
	Put    []Endpoint `hcl:"put,block"`

	Cron   []Cron   `hcl:"cron,block"`
	Queue  []Queue  `hcl:"queue,block"`
	Lambda []Lambda `hcl:"lambda,block"`
}

type Seconds int64

func (s Seconds) Duration() time.Duration {
	return time.Duration(s) * time.Second
}

// ParseFile scans all configuration from the file and updates paths in [Lambda.WorkDir].
func ParseFile(file string) (*Project, error) {
	var p Project
	err := hclsimple.DecodeFile(file, nil, &p)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	rootFilePath, err := filepath.Abs(file)
	if err != nil {
		return nil, fmt.Errorf("detect abs path: %w", err)
	}

	rootPath := filepath.Dir(rootFilePath)

	// calculate work dirs
	for i, l := range p.Lambda {
		l.WorkDir = filepath.Join(rootPath, filepath.Clean(l.WorkDir))
		p.Lambda[i] = l
	}

	// calculate static dir
	if p.Static != "" {
		p.Static = filepath.Join(rootPath, filepath.Clean(p.Static))
	}

	// check queues
	for _, q := range p.Queue {
		if !QueueNameReg.MatchString(q.Name) {
			return nil, fmt.Errorf("queue name '%s' is not allowed", q.Name)
		}
	}
	p.Name = filepath.Base(rootPath)
	return &p, nil
}
