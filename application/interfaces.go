package application

import (
	"context"
	"github.com/reddec/trusted-cgi/templates"
	"github.com/reddec/trusted-cgi/types"
	"io"
	"regexp"
	"time"
)

type FileSystem interface {
	// List files and dirs
	ListFiles(path string) ([]types.File, error)
	// Get content of file
	ReadFile(path string, output io.Writer) error
	// Write content to file
	WriteFile(path string, input io.Reader) error
	// Remove selected file
	RemoveFile(path string) error
	// Rename local file
	RenameFile(src, dest string) error
	// Pack content of lambda to tar.gz
	Content(tarball io.Writer) error
	// Set content of lambda from tar.gz and apply changes (re-index)
	SetContent(tarball io.Reader) error
}

// Lambda functions
type Actions interface {
	// List of actions defined in Makefile as targets
	Actions() ([]string, error)
	// Do target defined in Makefile. Time limit, global env and out can be nil.
	Do(ctx context.Context, name string, timeLimit time.Duration, globalEnv map[string]string, out io.Writer) error
	// Do scheduled actions based on last run
	DoScheduled(ctx context.Context, lastRun time.Time, globalEnv map[string]string)
}

// Basic invokable entity
//
// Highlights:
//
//    - can be executed
//    - can be destroyed
//    - can manage files
//    - doesn't know about UID and alias
type Lambda interface {
	FileSystem
	Actions
	// Manifest configuration
	Manifest() types.Manifest
	// Update manifest and apply changes (re-index)
	SetManifest(manifest types.Manifest) error
	// Running credentials
	Credentials() *types.Credential
	// Update credentials (could be null) (and apply ownership for files if needed)
	SetCredentials(creds *types.Credential) error
	// Remove lambda
	Remove() error
	// Invoke request, write response. Required header should be set by invoker
	Invoke(ctx context.Context, request types.Request, response io.Writer, globalEnv map[string]string) error
}

// Platform should index lambda, keep shared info (like env) and apply global configuration
//
// Highlights:
//
//    - should manage index lambda by UID
//    - should manage index lambda by Alias
//    - should keep shared info
//    - remove from index doesn't destroy lambda
type Platform interface {
	// Resolved user credentials (could be null)
	Credentials() *types.Credential
	// Platform configuration
	Config() Config
	// Update and apply new configuration
	SetConfig(config Config) error
	// List of all lambdas manifests (unordered) with UID and aliases
	List() []Definition
	// Get lambda by UID (if indexed)
	FindByUID(uid string) (*Definition, error)
	// Get lambda by link/alias (if indexed)
	FindByLink(link string) (*Definition, error)
	// Make link to target UID. Could fail if no target UID exists or link already bound to another lambda. Returns definition of lambda
	Link(targetUID string, linkName string) (*Definition, error)
	// Remove link by name. Returns old linked lambda or null
	Unlink(linkName string) (*Definition, error)
	// Put existent lambda to platform, index it and apply.
	Add(uid string, lambda Lambda) error
	// Remove existent lambda from platform and index (doesn't call underlying Remove() method)
	Remove(uid string)
	// Invoke lambda with platform global environment and logs results to tracker (if set)
	Invoke(ctx context.Context, lambda Lambda, request types.Request, out io.Writer) error
	// Do lambda action target defined in Makefile with platform global environment. Time limit and out can be nil
	Do(ctx context.Context, lambda Lambda, action string, timeLimit time.Duration, out io.Writer) error
}

// High-level use-cases
type Cases interface {
	// Create new lambda from remote Git repository. Will work only if SSH key set
	CreateFromGit(ctx context.Context, repo string) (string, error)
	// Create new lambda using provided template
	CreateFromTemplate(ctx context.Context, template templates.Template) (string, error)
	// Create empty lambda
	Create(ctx context.Context) (string, error)
	// Remove lamdba from index and definition
	Remove(uid string) error
	// Get underlying platform
	Platform() Platform
	// Get underlying queues manager
	Queues() Queues
	// Run scheduled actions from all lambda. Saves last run
	RunScheduledActions(ctx context.Context)
	// List of all templates without availability check
	Templates() (map[string]*templates.Template, error)
	// Content of SSH public key if set
	PublicSSHKey() ([]byte, error)
}

// Queue name limitations
var QueueNameReg = regexp.MustCompile(`^[a-z0-9A-Z-]{3,64}$`)

// Queues manager. Manages queues and linked worker
type Queues interface {
	// Put request to queue. If queue not exists, an error will be thrown
	Put(queue string, request *types.Request) error
	// Add new queue. See QueueNameReg for limitations
	Add(queue Queue) error
	// Remove queue and worker
	Remove(queue string) error
	// Assign queue to another lambda. If lambda is empty, queue will be stopped (but Put will still work)
	Assign(queue string, targetLambda string) error
	// List of all defined queue.
	List() []Queue
	// Find queues linked to lambda
	Find(targetLambda string) []Queue
}
