package local

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/reddec/trusted-cgi/core/lambdas"
	"github.com/reddec/trusted-cgi/internal"
	"github.com/reddec/trusted-cgi/types"
)

// New lambda manager with backed by local filesystem.
// To put all lambdas to index cache call ScanAll.
func New(rootDir string) *LambdaManager {
	return &LambdaManager{
		rootDir: rootDir,
	}
}

type LambdaManager struct {
	rootDir   string
	locks     sync.Map
	creds     *types.Credential
	globalEnv struct {
		lock sync.RWMutex
		data map[string]string
	}
	lambdas struct {
		lock sync.RWMutex
		data map[string]*lambdaDefinition
	}
}

// Set environment variable.
func (mgr *LambdaManager) SetEnv(key, value string) *LambdaManager {
	mgr.globalEnv.lock.Lock()
	defer mgr.globalEnv.lock.Unlock()
	if mgr.globalEnv.data == nil {
		mgr.globalEnv.data = make(map[string]string)
	}
	mgr.globalEnv.data[key] = value
	return mgr
}

// Find lambda handler by name (UID) in cache. ScanAll or Scan should be called first to put lambdas into the cache.
func (mgr *LambdaManager) Find(name string) (http.Handler, error) {
	mgr.lambdas.lock.RLock()
	defer mgr.lambdas.lock.RUnlock()

	lmb, ok := mgr.lambdas.data[name]
	if !ok {
		return nil, lambdas.ErrNotFound
	}
	return lmb, nil
}

func (mgr *LambdaManager) env() []string {
	cp := make([]string, 0, len(mgr.globalEnv.data))
	mgr.globalEnv.lock.RLock()
	defer mgr.globalEnv.lock.RUnlock()
	for k, v := range mgr.globalEnv.data {
		cp = append(cp, k+"="+v)
	}
	return cp
}

func (mgr *LambdaManager) getLock(name string) *sync.Mutex {
	v, _ := mgr.locks.LoadOrStore(name, &sync.Mutex{})
	return v.(*sync.Mutex)
}

type lambdaDefinition struct {
	uid      string
	manager  *LambdaManager
	manifest types.Manifest
	rootDir  string
}

func (local *lambdaDefinition) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	if local.asStatic(response, request) {
		return
	}

	if len(local.manifest.Run) == 0 {
		response.WriteHeader(http.StatusBadGateway)
		log.Println("run is not defined in manifest for lambda", local.uid)
		return
	}

	if local.manifest.Serial {
		lock := local.manager.getLock(local.uid)
		lock.Lock()
		defer lock.Unlock()
	}

	if local.manifest.Method != "" && request.Method != local.manifest.Method {
		response.WriteHeader(http.StatusMethodNotAllowed)
		log.Println("method", request.Method, "not allowed for lambda", local.uid)
		return
	}

	if local.manifest.TimeLimit > 0 {
		cctx, cancel := context.WithTimeout(request.Context(), time.Duration(local.manifest.TimeLimit))
		defer cancel()
		request = request.WithContext(cctx)
	}

	var input io.Reader = request.Body

	if local.manifest.MaximumPayload > 0 {
		input = io.LimitReader(input, local.manifest.MaximumPayload)
	}

	peekableResponse := wrapWriter(response, local.manifest.OutputHeaders)

	cmd := exec.CommandContext(request.Context(), local.manifest.Run[0], local.manifest.Run[1:]...)
	cmd.Dir = local.rootDir
	cmd.Stdin = input
	cmd.Stdout = peekableResponse
	cmd.Stderr = os.Stderr
	internal.SetCreds(cmd, local.manager.creds)
	internal.SetFlags(cmd)
	var environments = os.Environ()
	environments = append(environments, local.manager.env()...)
	for header, mapped := range local.manifest.InputHeaders {
		environments = append(environments, mapped+"="+request.Header.Get(header))
	}
	for query, mapped := range local.manifest.Query {
		environments = append(environments, mapped+"="+request.FormValue(query))
	}
	if local.manifest.MethodEnv != "" {
		environments = append(environments, local.manifest.MethodEnv+"="+request.Method)
	}
	if local.manifest.PathEnv != "" {
		environments = append(environments, local.manifest.PathEnv+"="+request.URL.Path)
	}
	for k, v := range local.manifest.Environment {
		environments = append(environments, k+"="+v)
	}
	cmd.Env = environments
	err := cmd.Run()

	if err == nil {
		_, _ = peekableResponse.flush()
		log.Println("lambda", local.uid, "invoked successfully")
		return
	}

	log.Println("run failed:", err)

	if peekableResponse.flushed {
		return
	}
	if errors.Is(err, context.DeadlineExceeded) {
		response.WriteHeader(http.StatusGatewayTimeout)
		return
	}

	response.WriteHeader(http.StatusInternalServerError)
}

func (local *lambdaDefinition) asStatic(response http.ResponseWriter, request *http.Request) bool {
	if local.manifest.Static == "" {
		return false
	}
	if !(request.Method == http.MethodGet || request.Method == http.MethodHead) {
		return false
	}

	rootDir := filepath.Join(local.rootDir, local.manifest.Static)

	http.FileServer(http.Dir(rootDir)).ServeHTTP(response, request)

	return true
}
