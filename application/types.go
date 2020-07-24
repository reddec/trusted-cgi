package application

import (
	"encoding/json"
	"os"

	"github.com/reddec/trusted-cgi/types"
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

type Queue struct {
	Name           string             `json:"name"`
	Target         string             `json:"target"`
	Retry          int                `json:"retry"`            // number of additional attempts
	MaxElementSize int64              `json:"max_element_size"` // max request size
	Interval       types.JsonDuration `json:"interval"`         // delay between attempts
}

type PolicyDefinition struct {
	AllowedIP     types.JsonStringSet `json:"allowed_ip,omitempty"`     // limit incoming connections from list of IP
	AllowedOrigin types.JsonStringSet `json:"allowed_origin,omitempty"` // limit incoming connections by origin header
	Public        bool                `json:"public"`                   // if public, tokens are ignores
	Tokens        map[string]string   `json:"tokens,omitempty"`         // limit request by value in Authorization header (token => title)
}

type Policy struct {
	ID         string              `json:"id"`
	Definition PolicyDefinition    `json:"definition"`
	Lambdas    types.JsonStringSet `json:"lambdas"`
}
