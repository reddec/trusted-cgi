package server

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

type File struct {
	Dir  bool   `json:"is_dir"`
	Name string `json:"name"`
}

type TemplateStatus struct {
	Template
	Available bool `json:"available"`
}

type API interface {
	// Login user by username and password. Returns signed JWT
	Login(ctx context.Context, login, password string) (*Token, error)
	// Change password for the user
	ChangePassword(ctx context.Context, token *Token, password string) (bool, error)
	// Create new app (lambda)
	Create(ctx context.Context, token *Token) (*application.App, error)
	// Project configuration
	Config(ctx context.Context, token *Token) (*application.ProjectConfig, error)
	// Apply new configuration and save it
	Apply(ctx context.Context, token *Token, config application.ProjectConfig) (bool, error)
	// Get all templates without filtering
	AllTemplates(ctx context.Context, token *Token) ([]*TemplateStatus, error)
	// Create new app/lambda/function using pre-defined template
	CreateFromTemplate(ctx context.Context, token *Token, templateName string) (*application.App, error)
	// Upload content from .tar.gz archive to app and call Install handler (if defined)
	Upload(ctx context.Context, token *Token, uid string, tarGz []byte) (bool, error)
	// Download content as .tar.gz archive from app
	Download(ctx context.Context, token *Token, uid string) ([]byte, error)
	// Push single file to app
	Push(ctx context.Context, token *Token, uid string, file string, content []byte) (bool, error)
	// Pull single file from app
	Pull(ctx context.Context, token *Token, uid string, file string) ([]byte, error)
	// List available apps (lambdas) in a project
	List(ctx context.Context, token *Token) ([]*application.App, error)
	// Remove app and call Uninstall handler (if defined)
	Remove(ctx context.Context, token *Token, uid string) (bool, error)
	// Templates with filter by availability including embedded
	Templates(ctx context.Context, token *Token) ([]*Template, error)
	// Files in func dir
	Files(ctx context.Context, token *Token, name string, dir string) ([]*File, error)
	// Info about application
	Info(ctx context.Context, token *Token, uid string) (*application.App, error)
	// Update application manifest
	Update(ctx context.Context, token *Token, uid string, manifest types.Manifest) (*application.App, error)
	// Create file or directory inside app
	CreateFile(ctx context.Context, token *Token, uid string, path string, dir bool) (bool, error)
	// Remove file or directory
	RemoveFile(ctx context.Context, token *Token, uid string, path string) (bool, error)
	// Rename file or directory
	RenameFile(ctx context.Context, token *Token, uid string, oldPath, newPath string) (bool, error)
	// Global last records
	GlobalStats(ctx context.Context, token *Token, limit int) ([]stats.Record, error)
	// Stats for the app
	Stats(ctx context.Context, token *Token, uid string, limit int) ([]stats.Record, error)
	// Actions available for the app
	Actions(ctx context.Context, token *Token, uid string) ([]string, error)
	// Invoke action in the app (if make installed)
	Invoke(ctx context.Context, token *Token, uid string, action string) (string, error)
	// Make link/alias for app
	Link(ctx context.Context, token *Token, uid string, alias string) (*application.App, error)
	// Remove link
	Unlink(ctx context.Context, token *Token, alias string) (*application.App, error)
}
