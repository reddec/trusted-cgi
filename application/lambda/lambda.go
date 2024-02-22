package lambda

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/reddec/trusted-cgi/internal"
	"github.com/reddec/trusted-cgi/types"
)

type localLambda struct {
	rootDir   string
	staticDir string
	uid       string
	manifest  types.Manifest
	creds     *types.Credential
	lock      sync.RWMutex
}

func (local *localLambda) UID() string { return local.uid }

func (local *localLambda) Manifest() types.Manifest {
	local.lock.RLock()
	defer local.lock.RUnlock()
	return local.manifest
}

func (local *localLambda) SetManifest(manifest types.Manifest) error {
	local.lock.Lock()
	defer local.lock.Unlock()
	err := manifest.SaveAs(local.manifestFile())
	if err != nil {
		return fmt.Errorf("save manifest: %w", err)
	}
	local.manifest = manifest
	return nil
}

func (local *localLambda) Credentials() *types.Credential {
	local.lock.RLock()
	defer local.lock.RUnlock()
	return local.creds
}

func (local *localLambda) SetCredentials(creds *types.Credential) error {
	local.lock.Lock()
	defer local.lock.Unlock()
	if !creds.Equal(local.creds) {
		local.creds = creds
		return local.applyFilesOwner()
	}
	return nil
}

func (local *localLambda) Invoke(ctx context.Context, request types.Request, response io.Writer, globalEnv map[string]string) error {
	local.lock.RLock()
	defer local.lock.RUnlock()
	defer request.Body.Close()

	if local.staticDir != "" && request.Method == http.MethodGet {
		return local.serveStaticFile(request, response)
	}

	if len(local.manifest.Run) == 0 {
		return fmt.Errorf("run is not defined in manifest")
	}

	if local.manifest.Method != "" && local.manifest.Method != local.manifest.Method {
		return fmt.Errorf("method not allowed")
	}

	if local.manifest.TimeLimit > 0 {
		cctx, cancel := context.WithTimeout(ctx, time.Duration(local.manifest.TimeLimit))
		defer cancel()
		ctx = cctx
	}

	var input io.Reader = request.Body

	if local.manifest.MaximumPayload > 0 {
		input = io.LimitReader(input, local.manifest.MaximumPayload)
	}

	cmd := exec.CommandContext(ctx, local.manifest.Run[0], local.manifest.Run[1:]...)
	cmd.Dir = local.rootDir
	cmd.Stdin = input
	cmd.Stdout = response
	cmd.Stderr = os.Stderr
	internal.SetCreds(cmd, local.creds)
	internal.SetFlags(cmd)
	var environments = os.Environ()
	for header, mapped := range globalEnv {
		environments = append(environments, header+"="+mapped)
	}
	for header, mapped := range local.manifest.InputHeaders {
		environments = append(environments, mapped+"="+request.Headers[header])
	}
	for query, mapped := range local.manifest.Query {
		environments = append(environments, mapped+"="+request.Form[query])
	}
	if local.manifest.MethodEnv != "" {
		environments = append(environments, local.manifest.MethodEnv+"="+request.Method)
	}
	if local.manifest.PathEnv != "" {
		environments = append(environments, local.manifest.PathEnv+"="+request.Path)
	}
	for k, v := range local.manifest.Environment {
		environments = append(environments, k+"="+v)
	}
	cmd.Env = environments
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("run failed: %w", err)
	}
	return nil
}

func (local *localLambda) serveStaticFile(request types.Request, response io.Writer) error {
	// poor man path trimming
	// trailing slash always removed (later replaced by index.html)
	_, file, _ := strings.Cut(strings.Trim(request.Path, "/"), "/")
	return local.writeStaticFile(file, response)
}

func (local *localLambda) Remove() error {
	return os.RemoveAll(local.rootDir)
}

func (local *localLambda) reindex() error {
	err := local.reloadManifest()
	if err != nil {
		return fmt.Errorf("reload manifest: %w", err)
	}
	root, err := filepath.Abs(local.rootDir)
	if err != nil {
		return fmt.Errorf("get root dir: %w", err)
	}
	if local.manifest.Static != "" {
		staticDir, err := filepath.Abs(filepath.Join(root, local.manifest.Static))
		if err != nil {
			return fmt.Errorf("get static dir: %w", err)
		}
		local.staticDir = staticDir
	} else {
		local.staticDir = ""
	}
	local.rootDir = root
	local.uid = filepath.Base(root)
	return nil
}

func (local *localLambda) manifestFile() string {
	return filepath.Join(local.rootDir, internal.ManifestFile)
}

func (local *localLambda) reloadManifest() error {
	var mf types.Manifest
	err := mf.LoadFrom(local.manifestFile())
	if err != nil {
		return err
	}
	local.manifest = mf
	return nil
}

func (local *localLambda) readIgnore() ([]string, error) {
	content, err := os.ReadFile(filepath.Join(local.rootDir, internal.CGIIgnore))
	if err == nil {
		return strings.Split(string(content), "\n"), nil
	}
	if os.IsNotExist(err) {
		return nil, nil
	}
	return nil, fmt.Errorf("read ignore file: %w", err)
}

func (local *localLambda) writeStaticFile(path string, out io.Writer) error {
	if path == "" {
		path = "index.html"
	}
	destPath, isLocal := local.resolvePath(local.staticDir, path)
	if !isLocal {
		return fmt.Errorf("attempt to access file out of the jail")
	}
	f, err := os.Open(destPath)
	if err != nil {
		return err
	}
	dir, err := f.Readdir(1)
	if len(dir) > 0 {
		f.Close()
		return local.writeStaticFile( path + "/index.html", out )
	}
	defer f.Close()
	_, err = io.Copy(out, f)
	return err
}
