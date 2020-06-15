package application

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/google/uuid"
	"github.com/reddec/trusted-cgi/templates"
	"github.com/reddec/trusted-cgi/types"
	"golang.org/x/crypto/ssh"
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
	"time"
)

const (
	ProjectManifest = "project.json"
	CGIIgnore       = ".cgiignore"
	SSHKeySize      = 3072
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
	User        string            `json:"user"`                  // user that will be used for jobs
	UnTar       []string          `json:"untar,omitempty"`       // custom tar zxf command
	Tar         []string          `json:"tar,omitempty"`         // custom tar zcf command
	Environment map[string]string `json:"environment,omitempty"` // global environment
}

func (cfg *ProjectConfig) UnTarCommand() []string {
	if len(cfg.UnTar) > 0 {
		return cfg.UnTar
	}
	return []string{"tar", "zxf", "-"}
}

func (cfg *ProjectConfig) TarCommand() []string {
	if len(cfg.Tar) > 0 {
		return cfg.Tar
	}
	return []string{"tar", "zcf", "-", "."}
}

func (cfg *ProjectConfig) Credentials() (*syscall.Credential, error) {
	mappedUser := cfg.User
	if cfg.User == "" {
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

// Open project: load or create project based on config, defined parameters will be used as default
func (cfg *ProjectConfig) LoadOrCreate(file string) (*Project, error) {
	var saved ProjectConfig
	err := saved.ReadFile(file)
	if err == nil {
		*cfg = saved
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	var srv = &Project{
		config:     *cfg,
		configFile: file,
		apps:       map[string]*App{},
		links:      map[string]*App{},
	}
	creds, err := cfg.Credentials()
	if err != nil {
		return nil, err
	}
	srv.creds = creds
	err = srv.unsafeScanAppsToCache()
	if err != nil {
		return nil, err
	}
	return srv, srv.Save()
}

func (cfg *ProjectConfig) WriteFile(file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(cfg)
}

func (cfg *ProjectConfig) ReadFile(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(cfg)
}

type Project struct {
	configFile    string
	config        ProjectConfig
	creds         *syscall.Credential
	keyFile       string
	publicKey     ssh.PublicKey
	appsLock      sync.RWMutex
	lastScheduler time.Time
	apps          map[string]*App
	links         map[string]*App // custom path to UID
	configLock    sync.Mutex
}

func (project *Project) RunnerUser() string { return project.config.User }

func (project *Project) ChangeUser(user string) error {
	project.appsLock.Lock()
	defer project.appsLock.Unlock()

	project.config.User = user
	creds, err := project.config.Credentials()
	if err != nil {
		return err
	}
	project.creds = creds
	for _, app := range project.apps {
		app.creds = creds
		err = app.ApplyOwner()
		if err != nil {
			return err
		}
	}

	return project.config.WriteFile(project.configFile)
}

func (project *Project) GlobalEnvironment() map[string]string {
	project.configLock.Lock()
	defer project.configLock.Unlock()
	return project.config.Environment
}

func (project *Project) SetGlobalEnvironment(env map[string]string) error {
	project.configLock.Lock()
	defer project.configLock.Unlock()
	project.config.Environment = env
	return project.config.WriteFile(project.configFile)
}

// Gets encoded public key if exists. Otherwise returns nil
func (project *Project) PublicKey() []byte {
	if project.publicKey == nil {
		return nil
	}
	return ssh.MarshalAuthorizedKey(project.publicKey)
}

// root directory to search for applications
func (project *Project) Root() string {
	return filepath.Dir(project.configFile)
}

func (project *Project) Save() error {
	f, err := os.Create(project.configFile)
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

func (project *Project) CloneApps() []*App {
	project.appsLock.RLock()
	defer project.appsLock.RUnlock()
	cp := make([]*App, 0, len(project.apps))
	for _, app := range project.apps {
		cp = append(cp, app)
	}
	return cp
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

func (project *Project) CreateFromGit(ctx context.Context, repo string) (*App, error) {
	uid := uuid.New().String()
	creds := project.Credentials()
	root := filepath.Join(project.Root(), uid)

	app, err := CreateAppGit(ctx, root, repo, project.keyFile, creds)
	if err != nil {
		_ = os.RemoveAll(root)
		return nil, err
	}

	project.appsLock.Lock()
	defer project.appsLock.Unlock()
	project.unsafeAttachApp(app)
	return app, nil
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

	if template.PostClone != "" {
		text, err := app.InvokeAction(ctx, template.PostClone, 0, project.config.Environment)
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

	project.unsafeAttachApp(app)
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

	args := project.config.UnTarCommand()
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
	args := project.config.TarCommand()
	ignoreFile := filepath.Join(project.Root(), CGIIgnore)
	if _, err := os.Stat(ignoreFile); err == nil {
		args = append(args, "--exclude-from", ignoreFile)
	}
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

func (project *Project) unsafeAttachApp(app *App) {
	project.apps[app.UID] = app
	for alias := range app.Manifest.Aliases {
		project.links[alias] = app
	}
}

func (project *Project) unsafeScanAppsToCache() error {
	list, err := ioutil.ReadDir(project.Root())
	if err != nil {
		return err
	}

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

			project.unsafeAttachApp(app)
		}
	}
	return nil
}

func (project *Project) SetupSSHKey(file string) error {
	if file == "" {
		log.Println("GIT disabled - no ssh key defined")
		return nil
	} else if pmdata, err := ioutil.ReadFile(file); err == nil {
		// exists
		info, _ := pem.Decode(pmdata)
		priv, err := x509.ParsePKCS1PrivateKey(info.Bytes)
		if err != nil {
			return err
		}
		project.keyFile = file
		project.publicKey, err = ssh.NewPublicKey(priv.Public())
		if err != nil {
			return err
		}
	} else if os.IsNotExist(err) {
		// generate SSH keys
		privateKey, err := project.generateSSHKeys(file)
		if err != nil {
			return err
		}
		project.keyFile = file
		project.publicKey, err = ssh.NewPublicKey(privateKey.Public())
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

func (project *Project) generateSSHKeys(file string) (*rsa.PrivateKey, error) {
	log.Println("generating ssh key to", file)
	privateKey, err := rsa.GenerateKey(rand.Reader, SSHKeySize)
	if err != nil {
		return nil, err
	}
	pemdata := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		},
	)
	err = ioutil.WriteFile(file, pemdata, 0600)
	if err != nil {
		return privateKey, err
	}

	if project.creds == nil {
		return privateKey, nil
	}
	return privateKey, os.Chown(file, int(project.creds.Uid), int(project.creds.Gid))
}

func isValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}
