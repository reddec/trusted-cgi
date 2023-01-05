package config

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/reddec/trusted-cgi/types"
	"gopkg.in/yaml.v3"
)

type Script struct {
	Exec        string            `json:"exec" yaml:"exec"`
	Timeout     Seconds           `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	WorkDir     string            `json:"work_dir,omitempty" yaml:"work_dir,omitempty"`
	Payload     *string           `json:"payload,omitempty" yaml:"payload,omitempty"`
	Environment map[string]string `json:"environment,omitempty" yaml:"environment,omitempty"`
}

type Cron struct {
	Schedule string `json:"schedule"`
	Script   `yaml:",inline"`
}

type Queue struct {
	HTTP     `yaml:",inline"`
	Script   `yaml:",inline"`
	Interval time.Duration `json:"interval,omitempty" yaml:"interval,omitempty"`
	Size     int64         `json:"size,omitempty" yaml:"size,omitempty"`
	Retry    int64         `json:"retry,omitempty" yaml:"retry,omitempty"`
}

type Policy struct {
	Name    string              `json:"name" yaml:"name"`
	IPs     types.JsonStringSet `json:"ips,omitempty" yaml:"ips,omitempty"`         // limit incoming connections from list of IP
	Origins types.JsonStringSet `json:"origins,omitempty" yaml:"origins,omitempty"` // limit incoming connections by origin header
	Tokens  []Token             `json:"tokens,omitempty" yaml:"tokens,omitempty"`   // limit request by value in Authorization header (token => title)
}

type Token struct {
	Title string `json:"title,omitempty" yaml:"title,omitempty"`
	Hash  string `json:"hash" yaml:"hash"`
}

func (q *Queue) Name() string {
	return q.Method + " " + url.PathEscape(q.Path)
}

type HTTP struct {
	Method   string            `json:"method" yaml:"method"`
	Path     string            `json:"path" yaml:"path"`
	Alias    []string          `json:"aliases,omitempty" yaml:"aliases,omitempty"`
	Body     int64             `json:"body,omitempty" yaml:"body,omitempty"`
	Status   int               `json:"status,omitempty" yaml:"status,omitempty"`
	Vars     map[string]string `json:"vars,omitempty" yaml:"vars,omitempty"` // parsed and stored before headers and calls
	Headers  map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
	Policies []string          `json:"policies,omitempty" yaml:"policies,omitempty"`
}

type Endpoint struct {
	HTTP   `yaml:",inline"`
	Script `yaml:",inline"`
}

type Project struct {
	Name      string     `json:"-" yaml:"-"` // to be filled by directory name
	Static    string     `json:"static,omitempty" yaml:"static,omitempty"`
	Endpoints []Endpoint `json:"endpoints,omitempty" yaml:"endpoints,omitempty"`
	Crons     []Cron     `json:"crons,omitempty" yaml:"crons,omitempty"`
	Queues    []Queue    `json:"queues,omitempty" yaml:"queues,omitempty"`
	Policies  []Policy   `json:"policies,omitempty" yaml:"policies,omitempty"`
}

type Seconds int64

func (s Seconds) Duration() time.Duration {
	return time.Duration(s) * time.Second
}

// ParseFile scans all configuration from the file and updates paths in [Lambda.WorkDir].
func ParseFile(file string) (*Project, error) {
	var p Project
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if err := yaml.NewDecoder(f).Decode(&p); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	rootFilePath, err := filepath.Abs(file)
	if err != nil {
		return nil, fmt.Errorf("detect abs path: %w", err)
	}

	rootPath := filepath.Dir(rootFilePath)

	// check policies
	var usedPolicy = make(map[string]bool, len(p.Policies))
	for _, pl := range p.Policies {
		if usedPolicy[pl.Name] {
			return nil, fmt.Errorf("policy %s declared more than once", pl.Name)
		}
		usedPolicy[pl.Name] = true
	}

	// check queues
	var usedQueues = make(map[string]bool, len(p.Queues))
	for i, q := range p.Queues {
		q.Method = strings.ToUpper(q.Method)
		if q.Status <= 0 {
			q.Status = http.StatusAccepted
		}
		p.Queues[i] = q
		key := q.Method + " " + q.Path
		if usedQueues[key] {
			return nil, fmt.Errorf("queue %s %s declared more than once", q.Method, q.Path)
		}
		usedQueues[key] = true
	}

	// check endpoints
	var usedEndpoint = make(map[string]bool, len(p.Endpoints))
	for i, ep := range p.Endpoints {
		ep.Method = strings.ToUpper(ep.Method)
		p.Endpoints[i] = ep
		key := ep.Method + " " + ep.Path
		if usedEndpoint[key] {
			return nil, fmt.Errorf("endpoint %s %s declared more than once", ep.Method, ep.Path)
		}
		usedEndpoint[key] = true
	}

	p.Name = filepath.Base(rootPath)
	return &p, nil
}
