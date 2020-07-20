package cases

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/reddec/trusted-cgi/application"
	"github.com/reddec/trusted-cgi/application/lambda"
	"github.com/reddec/trusted-cgi/templates"
	"github.com/reddec/trusted-cgi/types"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

func New(platform application.Platform, queues application.Queues, policies application.Policies, dir, templateDir string) (*casesImpl, error) {
	aTemplateDir, err := filepath.Abs(templateDir)
	if err != nil {
		return nil, fmt.Errorf("resolve template dir: %w", err)
	}
	aDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("resolve root dir: %w", err)
	}
	cs := &casesImpl{
		directory:    aDir,
		templatesDir: aTemplateDir,
		platform:     platform,
		queues:       queues,
		policies:     policies,
	}
	return cs, cs.Scan()
}

type casesImpl struct {
	sshLoader
	lastScheduler time.Time
	directory     string
	templatesDir  string
	platform      application.Platform
	queues        application.Queues
	policies      application.Policies
}

func (impl *casesImpl) Scan() error {
	list, err := ioutil.ReadDir(impl.directory)
	if err != nil {
		return fmt.Errorf("scan dir for lambdas: %w", err)
	}
	for _, item := range list {
		if item.IsDir() && isValidUUID(item.Name()) {
			uid := item.Name()
			path := filepath.Join(impl.directory, uid)
			fn, err := lambda.FromDir(path)
			if err != nil {
				return fmt.Errorf("load lambda %s: %w", uid, err)
			}
			err = impl.platform.Add(uid, fn)
			if err != nil {
				return fmt.Errorf("add lambda %s to index: %w", uid, err)
			}
			err = impl.applyMigration(uid, path, fn)
			if err != nil {
				return fmt.Errorf("apply migration for lambda %s: %w", uid, err)
			}
		}
	}
	return nil
}

func (impl *casesImpl) CreateFromGit(ctx context.Context, repo string) (string, error) {
	pk := impl.privateKeyFile
	if pk == "" {
		return "", fmt.Errorf("can't clone from Git while SSH key not set")
	}
	uid := uuid.New().String()
	path := filepath.Join(impl.directory, uid)

	err := os.MkdirAll(path, 0755)
	if err != nil {
		return uid, fmt.Errorf("create working directory: %w", err)
	}

	fn, err := lambda.FromGit(ctx, pk, repo, path)
	if err != nil {
		_ = os.RemoveAll(path)
		return uid, fmt.Errorf("clone repo: %w", err)
	}
	err = impl.platform.Add(uid, fn)
	if err != nil {
		_ = os.RemoveAll(path)
		return uid, fmt.Errorf("add cloned lambda to platform: %w", err)
	}
	return uid, nil
}

func (impl *casesImpl) CreateFromTemplate(ctx context.Context, template templates.Template) (string, error) {
	uid := uuid.New().String()
	path := filepath.Join(impl.directory, uid)

	err := os.MkdirAll(path, 0755)
	if err != nil {
		return uid, fmt.Errorf("create working directory: %w", err)
	}

	fn, err := lambda.FromTemplate(ctx, template, path)
	if err != nil {
		_ = os.RemoveAll(path)
		return uid, fmt.Errorf("create lambda: %w", err)
	}
	err = impl.platform.Add(uid, fn)
	if err != nil {
		_ = os.RemoveAll(path)
		return uid, fmt.Errorf("add new lambda to platform: %w", err)
	}
	return uid, nil
}

func (impl *casesImpl) Create(ctx context.Context) (string, error) {
	return impl.CreateFromTemplate(ctx, templates.Template{
		Manifest: types.Manifest{},
	})
}

func (impl *casesImpl) Platform() application.Platform {
	return impl.platform
}

func (impl *casesImpl) RunScheduledActions(ctx context.Context) {
	now := time.Now()
	last := impl.lastScheduler
	impl.lastScheduler = now
	for _, fn := range impl.platform.List() {
		fn.Lambda.DoScheduled(ctx, last, impl.platform.Config().Environment) // FIXME: too much access into platform internals
	}
}

func (impl *casesImpl) Templates() (map[string]*templates.Template, error) {
	return templates.List(impl.templatesDir)
}

func (impl *casesImpl) Remove(uid string) error {
	fn, err := impl.platform.FindByUID(uid)
	if err != nil {
		return fmt.Errorf("remove - find by uid %s: %w", uid, err)
	}
	impl.platform.Remove(uid)
	// unlink queues
	for _, q := range impl.queues.Find(uid) {
		err := impl.queues.Remove(q.Name)
		if err != nil {
			log.Println("[ERROR]", "failed remove queue", q.Name)
		}
	}
	// unlink from policies
	err = impl.policies.Clear(uid)
	if err != nil {
		log.Println("[ERROR]", "failed clear linked policy for lambda", uid, ":", err)
	}
	return fn.Lambda.Remove()
}

func (impl *casesImpl) Queues() application.Queues {
	return impl.queues
}
