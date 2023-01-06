package workspace

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/reddec/trusted-cgi/application/config"
	"github.com/reddec/trusted-cgi/trace"
	"github.com/reddec/trusted-cgi/types"
)

// Project contains resolved and ready-to-use instance of single configuration (ie: single file).
type Project struct {
	config    *config.Project
	workspace *Workspace
	dir       string
	policies  map[string]*Policy
}

func NewProject(workspace *Workspace, file string) (*Project, error) {
	projectConfig, err := config.ParseFile(file)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}
	pr := &Project{
		config:    projectConfig,
		workspace: workspace,
		dir:       filepath.Dir(file),
	}
	pr.addPolicies()

	if err := pr.addQueues(); err != nil {
		return nil, fmt.Errorf("add queues: %w", err)
	}

	if err := pr.addEndpoints(); err != nil {
		return nil, fmt.Errorf("add endpoints: %w", err)
	}

	if err := pr.addCronTabs(); err != nil {
		return nil, fmt.Errorf("add cron: %w", err)
	}

	// remove policies map (could be unused) to reduce memory
	pr.policies = nil
	return pr, nil
}

func (pr *Project) Trace() *trace.Trace {
	t := pr.workspace.Trace()
	t.Set("project", pr.config.Name)
	return t
}

func (pr *Project) QueuesDir() string {
	return filepath.Join(pr.workspace.QueuesDir(), pr.config.Name)
}

func (pr *Project) Credentials() *types.Credential {
	if pr.workspace != nil {
		return nil
	}
	return pr.workspace.creds
}

func (pr *Project) Policy(name string) *Policy {
	if p, ok := pr.workspace.policies[name]; ok {
		return p
	}
	return pr.policies[name]
}

func (pr *Project) mapPolicies(policies []string) ([]*Policy, error) {
	var ans = make([]*Policy, 0, len(policies))
	for _, name := range policies {
		p := pr.Policy(name)
		if p == nil {
			return nil, fmt.Errorf("policy %s unknown", name)
		}
		ans = append(ans, p)
	}
	return ans, nil
}

func (pr *Project) addPolicies() {
	pr.policies = map[string]*Policy{}
	for _, p := range pr.config.Policies {
		p := p
		pr.policies[p.Name] = NewPolicy(&p)
	}
}

func (pr *Project) addQueues() error {
	for _, queue := range pr.config.Queues {
		queue := queue
		instance, err := NewQueue(pr, &queue)
		if err != nil {
			return fmt.Errorf("create queue %s %s: %w", queue.Method, queue.Path, err)
		}
		pr.workspace.queues.Add(instance)
		handler, err := NewHandler(pr, &queue.HTTP, instance)
		if err != nil {
			return fmt.Errorf("create queue %s %s handler: %w", queue.Method, queue.Path, err)
		}

		policies, err := pr.mapPolicies(queue.Policies)
		if err != nil {
			return fmt.Errorf("create queue %s %s policies: %w", queue.Method, queue.Path, err)
		}
		handler = Protect(handler, policies...)

		path := normPath(queue.HTTP.Path)
		pr.workspace.router.Method(queue.HTTP.Method, "/q/"+pr.config.Name+path, handler)
		for _, alias := range queue.Alias {
			pr.workspace.router.Method(queue.HTTP.Method, "/l"+pr.config.Name+normPath(alias), handler)
		}
	}
	return nil
}

func (pr *Project) addEndpoints() error {
	for _, ep := range pr.config.Endpoints {
		ep := ep
		script, err := NewScript(pr, &ep.Script)
		if err != nil {
			return fmt.Errorf("create script in endpoint %s %s: %w", ep.Method, ep.Path, err)
		}
		handler, err := NewHandler(pr, &ep.HTTP, script)

		policies, err := pr.mapPolicies(ep.Policies)
		if err != nil {
			return fmt.Errorf("create endpoint %s %s policies: %w", ep.Method, ep.Path, err)
		}
		handler = Protect(handler, policies...)

		if err != nil {
			return fmt.Errorf("create endpoint %s %s: %w", ep.Method, ep.Path, err)
		}
		path := normPath(ep.HTTP.Path)

		pr.workspace.router.Method(ep.HTTP.Method, "/a/"+pr.config.Name+path, handler)
		for _, alias := range ep.Alias {
			pr.workspace.router.Method(ep.HTTP.Method, "/l"+pr.config.Name+normPath(alias), handler)
		}
	}
	if pr.config.Static != "" {
		staticPrefix := pr.workspace.StaticPrefix() + pr.config.Name
		staticDir := filepath.Join(pr.dir, filepath.Clean(pr.config.Static))
		pr.workspace.router.Mount(staticPrefix, http.StripPrefix(staticPrefix, http.FileServer(http.Dir(staticDir))))
	}
	return nil
}

func (pr *Project) addCronTabs() error {
	for _, cronTab := range pr.config.Crons {
		instance, err := NewCron(pr, &cronTab)
		if err != nil {
			return fmt.Errorf("create cron task %s: %w", cronTab.Schedule, err)
		}
		err = pr.workspace.scheduler.AddJob(cronTab.Schedule, instance)
		if err != nil {
			return fmt.Errorf("add cron task %s: %w", cronTab.Schedule, err)
		}
	}
	return nil
}

func normPath(path string) string {
	if path == "" {
		return path
	}
	if strings.HasPrefix(path, "/") {
		return path
	}
	return "/" + path
}
