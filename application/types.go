package application

import (
	"encoding/json"
	"github.com/reddec/trusted-cgi/types"
	"os"
)

type Definition struct {
	UID      string              `json:"uid"`
	Aliases  types.JsonStringSet `json:"aliases"`
	Manifest types.Manifest      `json:"manifest"`
	Lambda   Lambda              `json:"-"`
}

type Config struct {
	User        string            `json:"user"`                  // user that will be used for jobs
	Environment map[string]string `json:"environment,omitempty"` // global environment
	Links       map[string]string `json:"links,omitempty"`       // links (alias -> uid)
}

func (cfg Config) WithEnv(env map[string]string) Config {
	cfg.Environment = env
	return cfg
}

func (cfg Config) WithUser(user string) Config {
	cfg.User = user
	return cfg
}

func (cfg *Config) WriteFile(file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(cfg)
}

func (cfg *Config) ReadFile(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(cfg)
}
