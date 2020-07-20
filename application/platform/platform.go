package platform

import (
	"context"
	"fmt"
	"github.com/reddec/trusted-cgi/application"
	"github.com/reddec/trusted-cgi/types"
	"io"
	"os"
	"os/user"
	"regexp"
	"strconv"
	"sync"
	"time"
)

var allowedName = regexp.MustCompile("^[a-zA-Z0-9._-]{1,255}$")

func New(configFile string, validator application.Validator) (*platform, error) {
	var config application.Config
	err := config.ReadFile(configFile)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("read config: %w", err)
	}
	pl := &platform{
		configLocation: configFile,
		config:         config,
		validator:      validator,
	}
	return pl, pl.SetConfig(config)
}

type platform struct {
	creds          *types.Credential
	lock           sync.RWMutex
	config         application.Config
	configLocation string
	byUID          map[string]record
	validator      application.Validator
}

type record struct {
	lambda  application.Lambda
	aliases types.JsonStringSet
}

func (platform *platform) Credentials() *types.Credential {
	return platform.creds
}

func (platform *platform) Config() application.Config {
	platform.lock.RLock()
	defer platform.lock.RUnlock()
	return platform.config
}

func (platform *platform) SetConfig(config application.Config) error {
	platform.lock.Lock()
	creds, err := resolveUserCreds(config.User)
	if err != nil {
		platform.lock.Unlock()
		return fmt.Errorf("resolve user %s: %w", config.User, err)
	}
	err = config.WriteFile(platform.configLocation)
	if err != nil {
		platform.lock.Unlock()
		return fmt.Errorf("save config file: %w", err)
	}
	platform.creds = creds
	platform.config = config
	platform.lock.Unlock()
	platform.lock.Lock()
	defer platform.lock.Unlock()
	return platform.applyConfig()
}

func (platform *platform) Link(targetUID string, linkName string) (*application.Definition, error) {
	if !allowedName.MatchString(targetUID) {
		return nil, fmt.Errorf("target UID is not valid name - %s", allowedName.String())
	}
	if !allowedName.MatchString(linkName) {
		return nil, fmt.Errorf("link name is not valid name - %s", allowedName.String())
	}
	platform.lock.Lock()
	defer platform.lock.Unlock()
	target, ok := platform.byUID[targetUID]
	if !ok {
		return nil, fmt.Errorf("unknown target lambda %s", targetUID)
	}
	linked, exists := platform.config.Links[linkName]
	if exists && linked != targetUID {
		return nil, fmt.Errorf("link %s already pointed to another lambda %s", linkName, linked)
	}
	if platform.config.Links == nil {
		platform.config.Links = make(map[string]string)
	}
	target.aliases.Set(linkName)
	platform.config.Links[linkName] = targetUID
	return target.toDefinition(targetUID), platform.unsafeSaveConfig()
}

func (platform *platform) Unlink(linkName string) (*application.Definition, error) {
	if !allowedName.MatchString(linkName) {
		return nil, fmt.Errorf("link name is not valid name - %s", allowedName.String())
	}
	platform.lock.Lock()
	defer platform.lock.Unlock()
	uid, ok := platform.config.Links[linkName]
	delete(platform.config.Links, linkName)
	target, tOk := platform.byUID[uid]
	if tOk && ok {
		target.aliases.Del(linkName)
	}
	return target.toDefinition(uid), platform.unsafeSaveConfig()
}

func (platform *platform) List() []application.Definition {
	platform.lock.RLock()
	defer platform.lock.RUnlock()
	var ans = make([]application.Definition, 0, len(platform.byUID))
	for uid, record := range platform.byUID {
		def := record.toDefinition(uid)
		ans = append(ans, *def)
	}
	return ans
}

func (platform *platform) FindByUID(uid string) (*application.Definition, error) {
	platform.lock.RLock()
	defer platform.lock.RUnlock()
	lambda, ok := platform.byUID[uid]
	if !ok {
		return nil, fmt.Errorf("unkown lambda with UID %s", uid)
	}
	return lambda.toDefinition(uid), nil
}

func (platform *platform) FindByLink(link string) (*application.Definition, error) {
	platform.lock.RLock()
	defer platform.lock.RUnlock()
	uid, ok := platform.config.Links[link]
	if !ok {
		return nil, fmt.Errorf("unknown lambda with alias %s", link)
	}
	lambda, ok := platform.byUID[uid]
	if !ok {
		return nil, fmt.Errorf("broken link %s - unknown lambda %s", link, uid)
	}
	return lambda.toDefinition(uid), nil
}

func (platform *platform) Add(uid string, lambda application.Lambda) error {
	platform.lock.Lock()
	savedLambda, exists := platform.byUID[uid]
	if exists && savedLambda.lambda != lambda {
		platform.lock.Unlock()
		return fmt.Errorf("lambda %s already exists and different", uid)
	}
	if platform.byUID == nil {
		platform.byUID = make(map[string]record)
	}
	rec := record{lambda: lambda, aliases: make(types.JsonStringSet)}
	// search for already existent links
	for alias, target := range platform.config.Links {
		if target == uid {
			rec.aliases.Set(alias)
		}
	}
	platform.byUID[uid] = rec
	platform.lock.Unlock()

	if !exists {
		err := platform.setupLambda(lambda)
		if err != nil {
			return fmt.Errorf("setup new lambda %s: %w", uid, err)
		}
	}
	return nil
}

func (platform *platform) Remove(uid string) {
	platform.lock.Lock()
	defer platform.lock.Unlock()
	rec, ok := platform.byUID[uid]
	delete(platform.byUID, uid)
	if ok {
		for alias := range rec.aliases {
			delete(platform.config.Links, alias)
		}
	}
	_ = platform.unsafeSaveConfig()
}

func (platform *platform) InvokeByUID(ctx context.Context, uid string, request types.Request, out io.Writer) error {
	lambda, err := platform.FindByUID(uid)
	if err != nil {
		_ = request.Body.Close()
		return err
	}
	return platform.Invoke(ctx, lambda.Lambda, request, out)
}

func (platform *platform) Invoke(ctx context.Context, lambda application.Invokable, request types.Request, out io.Writer) error {
	err := platform.validator.Inspect(lambda.UID(), &request)
	if err != nil {
		return err
	}
	return lambda.Invoke(ctx, request, out, platform.config.Environment)
}

func (platform *platform) Do(ctx context.Context, lambda application.Lambda, action string, timeLimit time.Duration, out io.Writer) error {
	return lambda.Do(ctx, action, timeLimit, platform.config.Environment, out)
}

// apply configuration for lambda
func (platform *platform) setupLambda(lambda application.Lambda) error {
	err := lambda.SetCredentials(platform.creds)
	if err != nil {
		return fmt.Errorf("set credentials: %w", err)
	}
	return nil
}

func (platform *platform) applyConfig() error {
	for uid, record := range platform.byUID {
		err := record.lambda.SetCredentials(platform.creds)
		if err != nil {
			return fmt.Errorf("set credentials %s: %w", uid, err)
		}
	}
	return nil
}

func (platform *platform) unsafeSaveConfig() error {
	err := platform.config.WriteFile(platform.configLocation)
	if err != nil {

		return fmt.Errorf("save config file: %w", err)
	}
	return nil
}

func resolveUserCreds(name string) (*types.Credential, error) {
	if name == "" {
		return nil, nil
	}
	info, err := user.Lookup(name)
	if err != nil {
		return nil, err
	}
	uid, err := strconv.Atoi(info.Uid)
	if err != nil {
		return nil, err
	}
	gid, err := strconv.Atoi(info.Gid)
	if err != nil {
		return nil, err
	}
	return &types.Credential{
		User:  uid,
		Group: gid,
	}, nil
}

func (record *record) toDefinition(uid string) *application.Definition {
	if record == nil {
		return nil
	}
	return &application.Definition{
		UID:      uid,
		Aliases:  record.aliases.Dup(),
		Manifest: record.lambda.Manifest(),
		Lambda:   record.lambda,
	}
}
