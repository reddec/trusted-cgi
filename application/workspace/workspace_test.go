package workspace_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/reddec/trusted-cgi/application/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const projA = `
static: "done"
endpoints:
  - method: GET
    exec: echo -n today is a great day

  - method: GET
    path: /out.txt
    exec: echo -n it is everywhere

  - method: POST
    path: "calc"
    exec: "bc"

queues:
  - method: POST
    path: "calc"
    vars:
      request_id: "{{uuidv4}}"
    exec: "mkdir -p done/out && bc > done/out/$REQUEST_ID"
    environment:
      REQUEST_ID: "{{.Var.request_id}}"
    headers:
      "X-Correlation-Id": "{{.Var.request_id}}"

`

func TestNew_Workspace(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	workspaceDir := filepath.Join(tmpDir, "workspace")
	queueDir := filepath.Join(tmpDir, "queues")
	err = os.MkdirAll(workspaceDir, 0755)
	require.NoError(t, err)

	addProject(workspaceDir, "project-a", map[string]string{
		workspace.ProjectFiles[0]: projA,
	})

	wrk, err := workspace.New(workspace.Config{
		QueueDir: queueDir,
		Shell:    "/bin/bash",
	}, workspaceDir)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg multierror.Group

	wg.Go(func() error {
		return wrk.Run(ctx)
	})

	t.Run("endpoints should work", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/a/project-a", nil)
		rec := httptest.NewRecorder()
		wrk.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "today is a great day", rec.Body.String())
	})

	t.Run("payload should work", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/a/project-a/calc", bytes.NewBufferString("1+2+3\n3*3*2"))
		rec := httptest.NewRecorder()
		wrk.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "6\n18\n", rec.Body.String())
	})

	t.Run("async call", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/q/project-a/calc", bytes.NewBufferString("1+2+3\n3*3*2"))
		rec := httptest.NewRecorder()
		wrk.ServeHTTP(rec, req)
		require.Equal(t, http.StatusAccepted, rec.Code)
		requestID := rec.Header().Get("X-Correlation-ID")
		assert.NotEmpty(t, requestID)

		time.Sleep(time.Second)

		req = httptest.NewRequest(http.MethodGet, "/s/project-a/out/"+requestID, nil)
		rec = httptest.NewRecorder()
		wrk.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "6\n18\n", rec.Body.String())
	})

	cancel()
	require.NoError(t, wg.Wait().ErrorOrNil())
}

func TestNewReloadable(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	workspaceDir := filepath.Join(tmpDir, "workspace")
	queueDir := filepath.Join(tmpDir, "queues")
	err = os.MkdirAll(workspaceDir, 0755)
	require.NoError(t, err)

	addProject(workspaceDir, "project-a", map[string]string{
		workspace.ProjectFiles[0]: projA,
	})

	wrk, err := workspace.NewReloadable(workspace.Config{
		QueueDir: queueDir,
		Shell:    "/bin/bash",
	}, workspaceDir)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg multierror.Group

	wg.Go(func() error {
		return wrk.Run(ctx)
	})

	t.Run("reload should point to the new endpoint", func(t *testing.T) {
		// first should not work
		req := httptest.NewRequest(http.MethodGet, "/a/project-b", nil)
		rec := httptest.NewRecorder()
		wrk.ServeHTTP(rec, req)
		require.Equal(t, http.StatusNotFound, rec.Code)

		addProject(workspaceDir, "project-b", map[string]string{
			workspace.ProjectFiles[0]: `
endpoints:
  - method: GET # no path is default path
    exec: echo -n hell in world

`,
		})

		err = wrk.Reload()
		require.NoError(t, err)

		// after the reload this endpoint should work (we added project)
		req = httptest.NewRequest(http.MethodGet, "/a/project-b", nil)
		rec = httptest.NewRecorder()
		wrk.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "hell in world", rec.Body.String())

	})

	cancel()
	require.NoError(t, wg.Wait().ErrorOrNil())
}

func TestWorkspace_LegacyStatic(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	workspaceDir := filepath.Join(tmpDir, "workspace")
	err = os.MkdirAll(workspaceDir, 0755)
	require.NoError(t, err)

	addProject(workspaceDir, "project-a", map[string]string{
		workspace.ProjectFiles[0]: projA,
		"done/test.txt":           "hello world",
		"out.txt":                 "hell in world",
	})
	assert.FileExists(t, filepath.Join(workspaceDir, "project-a", "done/test.txt"))

	t.Run("non-legacy version should work by-default", func(t *testing.T) {
		wrk, err := workspace.New(workspace.Config{}, workspaceDir)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/s/project-a/test.txt", nil)
		rec := httptest.NewRecorder()
		wrk.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "hello world", rec.Body.String())
	})

	t.Run("legacy version should work over lambda prefix", func(t *testing.T) {
		wrk, err := workspace.New(workspace.Config{
			LegacyStatic: true,
		}, workspaceDir)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/a/project-a/test.txt", nil)
		rec := httptest.NewRecorder()
		wrk.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "hello world", rec.Body.String())
	})

	t.Run("in legacy, lambda has higher priority than static file", func(t *testing.T) {
		wrk, err := workspace.New(workspace.Config{
			LegacyStatic: true,
			Shell:        "/bin/bash",
		}, workspaceDir)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/a/project-a/out.txt", nil)
		rec := httptest.NewRecorder()
		wrk.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "it is everywhere", rec.Body.String())
	})
}

func addProject(rootDir, name string, files map[string]string) {
	p := filepath.Join(rootDir, name)
	err := os.MkdirAll(p, 0755)
	if err != nil {
		panic(err)
	}

	for fname, content := range files {
		path := filepath.Join(p, fname)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			panic(err)
		}
		err = os.WriteFile(path, []byte(content), 0755)
		if err != nil {
			panic(err)
		}
	}
}
