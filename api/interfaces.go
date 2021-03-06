package api

import (
	"context"
	"encoding/json"
	"github.com/reddec/trusted-cgi/application"
	"github.com/reddec/trusted-cgi/stats"
	"github.com/reddec/trusted-cgi/types"
)

// JWT wrapper , should be unmarshalled from string
type Token struct {
	Login string `json:"-"` // parsed by validator
	Data  string `json:"-"` // raw JWT
}

func (t *Token) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, &t.Data)
}

func (t *Token) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Data)
}

type Template struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type TemplateStatus struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Available   bool   `json:"available"`
}

type Settings struct {
	User        string            `json:"user"`                  // effective user (user for run apps)
	PublicKey   string            `json:"public_key,omitempty"`  // optional public RSA key for SSH
	Environment map[string]string `json:"environment,omitempty"` // global environment
}

type Environment struct {
	Environment map[string]string `json:"environment,omitempty"` // global environment
}

// API for lambdas
type LambdaAPI interface {
	// Upload content from .tar.gz archive to app and call Install handler (if defined)
	Upload(ctx context.Context, token *Token, uid string, tarGz []byte) (bool, error)
	// Download content as .tar.gz archive from app
	Download(ctx context.Context, token *Token, uid string) ([]byte, error)
	// Push single file to app
	Push(ctx context.Context, token *Token, uid string, file string, content []byte) (bool, error)
	// Pull single file from app
	Pull(ctx context.Context, token *Token, uid string, file string) ([]byte, error)
	// Remove app and call Uninstall handler (if defined)
	Remove(ctx context.Context, token *Token, uid string) (bool, error)
	// Files in func dir
	Files(ctx context.Context, token *Token, uid string, dir string) ([]types.File, error)
	// Info about application
	Info(ctx context.Context, token *Token, uid string) (*application.Definition, error)
	// Update application manifest
	Update(ctx context.Context, token *Token, uid string, manifest types.Manifest) (*application.Definition, error)
	// Create file or directory inside app
	CreateFile(ctx context.Context, token *Token, uid string, path string, dir bool) (bool, error)
	// Remove file or directory
	RemoveFile(ctx context.Context, token *Token, uid string, path string) (bool, error)
	// Rename file or directory
	RenameFile(ctx context.Context, token *Token, uid string, oldPath, newPath string) (bool, error)
	// Stats for the app
	Stats(ctx context.Context, token *Token, uid string, limit int) ([]stats.Record, error)
	// Actions available for the app
	Actions(ctx context.Context, token *Token, uid string) ([]string, error)
	// Invoke action in the app (if make installed)
	Invoke(ctx context.Context, token *Token, uid string, action string) (string, error)
	// Make link/alias for app
	Link(ctx context.Context, token *Token, uid string, alias string) (*application.Definition, error)
	// Remove link
	Unlink(ctx context.Context, token *Token, alias string) (*application.Definition, error)
}

// API for global project
type ProjectAPI interface {
	// Get global configuration
	Config(ctx context.Context, token *Token) (*Settings, error)
	// Change effective user
	SetUser(ctx context.Context, token *Token, user string) (*Settings, error)
	// Change global environment
	SetEnvironment(ctx context.Context, token *Token, env Environment) (*Settings, error)
	// Get all templates without filtering
	AllTemplates(ctx context.Context, token *Token) ([]*TemplateStatus, error)
	// List available apps (lambdas) in a project
	List(ctx context.Context, token *Token) ([]application.Definition, error)
	// Templates with filter by availability including embedded
	Templates(ctx context.Context, token *Token) ([]*Template, error)
	// Global last records
	Stats(ctx context.Context, token *Token, limit int) ([]stats.Record, error)
	// Create new app (lambda)
	Create(ctx context.Context, token *Token) (*application.Definition, error)
	// Create new app/lambda/function using pre-defined template
	CreateFromTemplate(ctx context.Context, token *Token, templateName string) (*application.Definition, error)
	// Create new app/lambda/function using remote Git repo
	CreateFromGit(ctx context.Context, token *Token, repo string) (*application.Definition, error)
}

// User/admin profile API
type UserAPI interface {
	// Login user by username and password. Returns signed JWT
	Login(ctx context.Context, login, password string) (*Token, error)
	// Change password for the user
	ChangePassword(ctx context.Context, token *Token, password string) (bool, error)
}

// API for managing queues
type QueuesAPI interface {
	// Create queue and link it to lambda and start worker
	Create(ctx context.Context, token *Token, queue application.Queue) (*application.Queue, error)
	// Remove queue and stop worker
	Remove(ctx context.Context, token *Token, name string) (bool, error)
	// Linked queues for lambda
	Linked(ctx context.Context, token *Token, lambda string) ([]application.Queue, error)
	// List of all queues
	List(ctx context.Context, token *Token) ([]application.Queue, error)
	// Assign lambda to queue (re-link)
	Assign(ctx context.Context, token *Token, name string, lambda string) (bool, error)
}

// API for managing policies
type PoliciesAPI interface {
	// List all policies
	List(ctx context.Context, token *Token) ([]application.Policy, error)
	// Create new policy
	Create(ctx context.Context, token *Token, policy string, definition application.PolicyDefinition) (*application.Policy, error)
	// Remove policy
	Remove(ctx context.Context, token *Token, policy string) (bool, error)
	// Update policy definition
	Update(ctx context.Context, token *Token, policy string, definition application.PolicyDefinition) (bool, error)
	// Apply policy for the resource
	Apply(ctx context.Context, token *Token, lambda string, policy string) (bool, error)
	// Clear applied policy for the lambda
	Clear(ctx context.Context, token *Token, lambda string) (bool, error)
}
