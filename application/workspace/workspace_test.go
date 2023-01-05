package workspace_test

import (
	"bytes"
	"context"
	"github.com/hashicorp/go-multierror"
	"github.com/reddec/trusted-cgi/application/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const projA = `
static = "public"
lambda "date" {
	exec = ["echo", "-n", "today is a great day"]
}

lambda "calc" {
	exec = ["bc"]
}

lambda "lazy-calc" {
	exec = ["sh", "-c", "echo ${"$"}CORRELATION_ID 1>&2; mkdir -p public/out && bc > public/out/${"$"}CORRELATION_ID"]
}

queue "lazy-calc" {
	call "lazy-calc" {}
}

get "" {
	call "date" {}
}

post "calc" {
	call "calc" {}
}

post "async-calc" {
	vars = {
		request_id = "{{uuidv4}}"
	}
	headers = {
		"X-Correlation-Id" = "{{.Var.request_id}}"
	}
	enqueue "lazy-calc" {
		Environment = {
			"CORRELATION_ID" = "{{.Var.request_id}}"
		}
	}
}
`

func TestNew_Workspace(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	workspaceDir := filepath.Join(tmpDir, "workspace")
	queueDir := filepath.Join(tmpDir, "queues")
	cacheDir := filepath.Join(tmpDir, "cache")
	err = os.MkdirAll(workspaceDir, 0755)
	require.NoError(t, err)

	addProject(workspaceDir, "project-a", map[string]string{
		workspace.ProjectFile: projA,
	})

	wrk, err := workspace.New(workspace.Config{
		QueueDir: queueDir,
		CacheDir: cacheDir,
	}, workspaceDir)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg multierror.Group

	wg.Go(func() error {
		return wrk.Run(ctx)
	})

	t.Run("endpoints should work", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/project-a", nil)
		rec := httptest.NewRecorder()
		wrk.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "today is a great day", rec.Body.String())
	})

	t.Run("payload should work", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/project-a/calc", bytes.NewBufferString("1+2+3\n3*3*2"))
		rec := httptest.NewRecorder()
		wrk.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "6\n18\n", rec.Body.String())
	})

	t.Run("async call", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/project-a/async-calc", bytes.NewBufferString("1+2+3\n3*3*2"))
		rec := httptest.NewRecorder()
		wrk.ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)
		requestID := rec.Header().Get("X-Correlation-ID")
		assert.NotEmpty(t, requestID)

		time.Sleep(time.Second)

		req = httptest.NewRequest(http.MethodGet, "/project-a/out/"+requestID, nil)
		rec = httptest.NewRecorder()
		wrk.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "6\n18\n", rec.Body.String())
	})

	t.Run("check that there are no left-over cache", func(t *testing.T) {
		list, err := os.ReadDir(cacheDir)
		require.NoError(t, err)
		assert.Empty(t, list)
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
	cacheDir := filepath.Join(tmpDir, "cache")
	err = os.MkdirAll(workspaceDir, 0755)
	require.NoError(t, err)

	addProject(workspaceDir, "project-a", map[string]string{
		workspace.ProjectFile: projA,
	})

	wrk, err := workspace.NewReloadable(workspace.Config{
		QueueDir: queueDir,
		CacheDir: cacheDir,
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
		req := httptest.NewRequest(http.MethodGet, "/project-b", nil)
		rec := httptest.NewRecorder()
		wrk.ServeHTTP(rec, req)
		require.Equal(t, http.StatusNotFound, rec.Code)

		addProject(workspaceDir, "project-b", map[string]string{
			workspace.ProjectFile: `
lambda "hell" {
	exec = ["echo", "-n", "hell in world"]
}

get "" {
	call "hell" {}
}
`,
		})

		err = wrk.Reload()
		require.NoError(t, err)

		// after the reload this endpoint should work (we added project)
		req = httptest.NewRequest(http.MethodGet, "/project-b", nil)
		rec = httptest.NewRecorder()
		wrk.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "hell in world", rec.Body.String())

	})

	cancel()
	require.NoError(t, wg.Wait().ErrorOrNil())
}

func addProject(rootDir, name string, files map[string]string) {
	p := filepath.Join(rootDir, name)
	err := os.MkdirAll(p, 0755)
	if err != nil {
		panic(err)
	}

	for fname, content := range files {
		err = os.WriteFile(filepath.Join(p, fname), []byte(content), 0755)
		if err != nil {
			panic(err)
		}
	}
}
