package application

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/reddec/trusted-cgi/templates"
	"github.com/reddec/trusted-cgi/types"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
)

const (
	ProjectManifest = "project.json"
)

func OpenProject(location string, defaultConfig ProjectConfig) (*Project, error) {
	rootDir, err := filepath.Abs(location)
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(rootDir, 0755)
	if err != nil {
		return nil, err
	}
	return defaultConfig.LoadOrCreate(filepath.Join(rootDir, ProjectManifest))
}

type ProjectConfig struct {
	User  string   `json:"user"`            // user that will be used for jobs
	UnTar []string `json:"untar,omitempty"` // custom tar zxf command
	Tar   []string `json:"tar,omitempty"`   // custom tar zcf command
}

func (project *ProjectConfig) UnTarCommand() []string {
	if len(project.UnTar) > 0 {
		return project.UnTar
	}
	return []string{"tar", "zxf", "-"}
}

func (project *ProjectConfig) TarCommand() []string {
	if len(project.Tar) > 0 {
		return project.Tar
	}
	return []string{"tar", "zcf", "-", "."}
}

func (project *ProjectConfig) Credentials() (*syscall.Credential, error) {
	mappedUser := project.User
	if project.User == "" {
		return nil, nil
	}
	cred, err := user.Lookup(mappedUser)
	if err != nil {
		return nil, err
	}
	uid, err := strconv.ParseUint(cred.Uid, 10, 32)
	if err != nil {
		return nil, err
	}
	gid, err := strconv.ParseUint(cred.Gid, 10, 32)
	if err != nil {
		return nil, err
	}
	return &syscall.Credential{
		Uid: uint32(uid),
		Gid: uint32(gid),
	}, nil
}

func (project *ProjectConfig) LoadOrCreate(file string) (*Project, error) {
	var srv = &Project{
		ProjectConfig: *project,
		file:          file,
		apps:          map[string]*App{},
	}
	err := srv.read()
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	creds, err := srv.ProjectConfig.Credentials()
	if err != nil {
		return nil, err
	}
	srv.creds = creds
	err = srv.scanAppsToCache()
	if err != nil {
		return nil, err
	}
	return srv, srv.Save()
}

type Project struct {
	ProjectConfig
	creds      *syscall.Credential
	file       string
	appsLock   sync.RWMutex
	apps       map[string]*App
	links      map[string]*App // custom path to UID
	configLock sync.Mutex
}

// Replace project config and do necessary updates.
//
// If user changed - update all credentials in project and in apps, apply ownership for all files
func (project *Project) ApplyConfig(cfg ProjectConfig) error {
	project.appsLock.Lock()
	defer project.appsLock.Unlock()
	if cfg.User != project.User {
		creds, err := cfg.Credentials()
		if err != nil {
			return err
		}

		for _, app := range project.apps {
			app.creds = creds
			err = app.ApplyOwner()
			if err != nil {
				return err
			}
		}

		project.creds = creds
	}

	project.ProjectConfig = cfg
	return nil
}

// root directory to search for applications
func (project *Project) Root() string {
	return filepath.Dir(project.file)
}

func (project *Project) Save() error {
	f, err := os.Create(project.file)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(project)
}

func (project *Project) Credentials() *syscall.Credential {
	return project.creds
}

func (project *Project) Create(ctx context.Context) (*App, error) {
	return project.CreateFromTemplate(ctx, &templates.Template{
		Manifest: types.Manifest{},
	})
}

func (project *Project) Link(uid string, alias string) (*App, error) {
	project.appsLock.Lock()
	defer project.appsLock.Unlock()
	app := project.apps[uid]
	if app == nil {
		return nil, fmt.Errorf("app %s not found", uid)
	}
	if anotherApp, ok := project.links[alias]; ok {
		return nil, fmt.Errorf("alias %s already used by %s (%s)", alias, anotherApp.UID, anotherApp.Manifest.Name)
	}
	if app.Manifest.Aliases == nil {
		app.Manifest.Aliases = make(types.JsonStringSet)
	}
	if project.links == nil {
		project.links = make(map[string]*App)
	}
	app.Manifest.Aliases.Set(alias)
	project.links[alias] = app
	return app, app.Manifest.SaveAs(app.ManifestFile())
}

func (project *Project) Unlink(alias string) (*App, error) {
	project.appsLock.Lock()
	defer project.appsLock.Unlock()
	anotherApp, ok := project.links[alias]
	if !ok {
		return nil, fmt.Errorf("alias %s is uknown", alias)
	}
	delete(project.links, alias)

	anotherApp.Manifest.Aliases.Del(alias)
	return anotherApp, anotherApp.Manifest.SaveAs(anotherApp.ManifestFile())
}

func (project *Project) CreateFromTemplate(ctx context.Context, template *templates.Template) (*App, error) {
	project.appsLock.Lock()
	defer project.appsLock.Unlock()

	uid := uuid.New().String()
	creds := project.Credentials()
	root := filepath.Join(project.Root(), uid)

	app, err := CreateApp(root, creds, template.Manifest)
	if err != nil {
		_ = os.RemoveAll(app.location)
		return nil, err
	}

	for fileName, content := range template.Files {
		err := app.WriteFile(fileName, []byte(content))
		if err != nil {
			_ = os.RemoveAll(app.location)
			return nil, err
		}
	}

	if template.Manifest.PostClone != "" {
		text, err := app.InvokeAction(ctx, template.Manifest.PostClone)
		if err != nil {
			log.Println("action run:", text)
			_ = os.RemoveAll(app.location)
			return nil, err
		}
	}

	err = app.ApplyOwner()
	if err != nil {
		_ = os.RemoveAll(app.location)
		return nil, err
	}

	project.apps[uid] = app

	return app, nil
}

func (project *Project) FindApp(uid string) *App {
	project.appsLock.RLock()
	defer project.appsLock.RUnlock()
	return project.apps[uid]
}

func (project *Project) FindAppByAlias(alias string) *App {
	project.appsLock.RLock()
	defer project.appsLock.RUnlock()
	return project.links[alias]
}

func (project *Project) Upload(ctx context.Context, uid string, tarGzBall io.Reader) error {
	app := project.FindApp(uid)
	if app == nil {
		return fmt.Errorf("no such app")
	}

	args := project.UnTarCommand()
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = tarGzBall
	cmd.Dir = app.location

	err := cmd.Run()
	if err != nil {
		return err
	}

	return app.ApplyOwner()
}

func (project *Project) Download(ctx context.Context, uid string, tarGzBall io.Writer) error {
	app := project.FindApp(uid)
	if app == nil {
		return fmt.Errorf("no such app")
	}
	args := project.TarCommand()
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = tarGzBall
	cmd.Dir = app.location
	return cmd.Run()
}

func (project *Project) List() []*App {
	var ans = make([]*App, 0, len(project.apps))
	project.appsLock.RLock()
	defer project.appsLock.RUnlock()
	for _, app := range project.apps {
		ans = append(ans, app)
	}
	return ans
}

func (project *Project) Remove(ctx context.Context, uid string) error {
	project.appsLock.Lock()
	defer project.appsLock.Unlock()
	app := project.apps[uid]
	delete(project.apps, uid)
	if app == nil {
		return nil
	}
	var links = make(map[string]*App)
	for alias, tapp := range project.links {
		if tapp != app {
			links[alias] = tapp
		}
	}
	project.links = links
	return os.RemoveAll(app.location)
}

func (project *Project) read() error {
	f, err := os.Open(project.file)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(project)
}

func (project *Project) scanAppsToCache() error {
	list, err := ioutil.ReadDir(project.Root())
	if err != nil {
		return err
	}

	project.appsLock.Lock()
	defer project.appsLock.Unlock()

	if project.links == nil {
		project.links = make(map[string]*App)
	}

	for _, item := range list {
		uid := item.Name()
		if item.IsDir() && isValidUUID(uid) {
			app, err := OpenApp(filepath.Join(project.Root(), uid), project.Credentials())
			if err != nil {
				return fmt.Errorf("open app %s: %w", uid, err)
			}
			project.apps[uid] = app

			for link := range app.Manifest.Aliases {
				project.links[link] = app
			}
		}
	}
	return nil
}

func isValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}
