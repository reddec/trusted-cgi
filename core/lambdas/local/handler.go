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

type lambdaManager struct {
	locks     sync.Map
	creds     *types.Credential
	globalEnv map[string]string  // atomic swap, immutable
	lambdas   map[string]*lambda // atomic swap, immutable
}

func (mgr *lambdaManager) getLock(name string) *sync.Mutex {
	v, _ := mgr.locks.LoadOrStore(name, &sync.Mutex{})
	return v.(*sync.Mutex)
}

func (mgr *lambdaManager) Find(name string) (lambdas.Lambda, error) {
	lmb, ok := mgr.lambdas[name]
	if !ok {
		return nil, lambdas.ErrNotFound
	}
	return lmb, nil
}

type lambda struct {
	uid      string
	manager  *lambdaManager
	manifest types.Manifest
	rootDir  string
}

func (local *lambda) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	// TODO: log
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

	if local.manifest.Method != "" && local.manifest.Method != local.manifest.Method {
		response.WriteHeader(http.StatusMethodNotAllowed)
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
	for header, mapped := range local.manager.globalEnv {
		environments = append(environments, header+"="+mapped)
	}
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

func (local *lambda) asStatic(response http.ResponseWriter, request *http.Request) bool {
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
