package server_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/reddec/trusted-cgi/api/services"
	"github.com/reddec/trusted-cgi/application"
	"github.com/reddec/trusted-cgi/application/cases"
	"github.com/reddec/trusted-cgi/application/platform"
	"github.com/reddec/trusted-cgi/application/policy"
	"github.com/reddec/trusted-cgi/application/queuemanager"
	"github.com/reddec/trusted-cgi/queue"
	"github.com/reddec/trusted-cgi/queue/inmemory"
	"github.com/reddec/trusted-cgi/server"
	"github.com/reddec/trusted-cgi/stats/impl/memlog"
	"github.com/reddec/trusted-cgi/templates"
	"github.com/reddec/trusted-cgi/types"
)

type testServer struct {
	Server server.Server
	Dir    string
}

func (ts *testServer) AddDummyLambda(ctx context.Context, run string, args ...string) (string, error) {
	return ts.Server.Cases.CreateFromTemplate(ctx, templates.Template{
		Manifest: types.Manifest{
			Run: append([]string{run}, args...),
		},
	})
}

func createTestServer() (*testServer, error) {
	ctx := context.Background()

	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, err
	}
	policies, err := policy.New(policy.FileConfig(filepath.Join(tmpDir, "policies.json")))
	if err != nil {
		return nil, err
	}

	basePlatform, err := platform.New(filepath.Join(tmpDir, "project.json"))
	if err != nil {
		return nil, err
	}

	queueFactory := func(name string) (queue.Queue, error) {
		return inmemory.New(1024), nil
	}

	queueManager, err := queuemanager.New(ctx, queuemanager.FileConfig(filepath.Join(tmpDir, "queues.json")), basePlatform, queueFactory)
	if err != nil {
		return nil, err
	}

	useCases, err := cases.New(basePlatform, queueManager, policies, tmpDir, filepath.Join(tmpDir, ".templates"))
	if err != nil {
		return nil, err
	}

	tracker := memlog.New(1000)

	projectApi := services.NewProjectSrv(useCases, tracker)
	lambdaApi := services.NewLambdaSrv(useCases, tracker)
	queuesApi := services.NewQueuesSrv(queueManager)
	policiesApi := services.NewPoliciesSrv(policies)
	userApi, err := services.CreateUserSrv(filepath.Join(tmpDir, "server.json"), "admin")
	if err != nil {
		return nil, err
	}

	srv := server.Server{
		Policies:     policies,
		Platform:     basePlatform,
		Cases:        useCases,
		Queues:       queueManager,
		Dev:          true,
		Tracker:      tracker,
		TokenHandler: userApi,
		ProjectAPI:   projectApi,
		LambdaAPI:    lambdaApi,
		UserAPI:      userApi,
		QueuesAPI:    queuesApi,
		PoliciesAPI:  policiesApi,
	}
	return &testServer{
		Server: srv,
		Dir:    tmpDir,
	}, nil
}

func TestHandlerByUID(t *testing.T) {
	ctx := context.Background()
	srv, err := createTestServer()
	if !assert.NoError(t, err) {
		return
	}
	defer os.RemoveAll(srv.Dir)
	handler := srv.Server.Handler(ctx)

	uid, err := srv.AddDummyLambda(ctx, "cat", "-")
	assert.NoError(t, err)

	t.Log("UID:", uid)

	rr := httptest.NewRecorder()

	req, err := http.NewRequest(http.MethodPost, "https://example.com/a/"+uid, bytes.NewBufferString("hello"))
	assert.NoError(t, err)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "hello", rr.Body.String())
}

func TestHandlerByUID_forbidden(t *testing.T) {
	ctx := context.Background()
	srv, err := createTestServer()
	if !assert.NoError(t, err) {
		return
	}
	defer os.RemoveAll(srv.Dir)
	handler := srv.Server.Handler(ctx)

	uid, err := srv.AddDummyLambda(ctx, "cat", "-")
	assert.NoError(t, err)

	_, err = srv.Server.Policies.Create("temp", application.PolicyDefinition{
		Public: false,
	})
	assert.NoError(t, err)
	err = srv.Server.Policies.Apply(uid, "temp")
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	req, err := http.NewRequest(http.MethodPost, "https://example.com/a/"+uid, bytes.NewBufferString("hello"))
	assert.NoError(t, err)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestHandlerByAlias(t *testing.T) {
	ctx := context.Background()
	srv, err := createTestServer()
	if !assert.NoError(t, err) {
		return
	}
	defer os.RemoveAll(srv.Dir)
	handler := srv.Server.Handler(ctx)

	uid, err := srv.AddDummyLambda(ctx, "cat", "-")
	assert.NoError(t, err)
	_, err = srv.Server.Platform.Link(uid, "test-link")
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	req, err := http.NewRequest(http.MethodPost, "https://example.com/l/test-link", bytes.NewBufferString("hello"))
	assert.NoError(t, err)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "hello", rr.Body.String())
}

func TestHandlerByAlias_forbidden(t *testing.T) {
	ctx := context.Background()
	srv, err := createTestServer()
	if !assert.NoError(t, err) {
		return
	}
	defer os.RemoveAll(srv.Dir)
	handler := srv.Server.Handler(ctx)

	uid, err := srv.AddDummyLambda(ctx, "cat", "-")
	assert.NoError(t, err)

	_, err = srv.Server.Policies.Create("temp", application.PolicyDefinition{
		Public: false,
	})
	assert.NoError(t, err)
	err = srv.Server.Policies.Apply(uid, "temp")
	assert.NoError(t, err)
	_, err = srv.Server.Platform.Link(uid, "test-link")
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	req, err := http.NewRequest(http.MethodPost, "https://example.com/l/test-link", bytes.NewBufferString("hello"))
	assert.NoError(t, err)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestHandlerByQueue(t *testing.T) {
	ctx := context.Background()
	srv, err := createTestServer()
	if !assert.NoError(t, err) {
		return
	}
	defer os.RemoveAll(srv.Dir)
	handler := srv.Server.Handler(ctx)

	uid, err := srv.AddDummyLambda(ctx, "cat", "-")
	assert.NoError(t, err)
	err = srv.Server.Queues.Add(application.Queue{
		Name:           "my-queue",
		Target:         uid,
		Retry:          1,
		MaxElementSize: 1024,
	})
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	req, err := http.NewRequest(http.MethodPost, "https://example.com/q/my-queue", bytes.NewBufferString("hello"))
	assert.NoError(t, err)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestHandlerByQueue_forbidden(t *testing.T) {
	ctx := context.Background()
	srv, err := createTestServer()
	if !assert.NoError(t, err) {
		return
	}
	defer os.RemoveAll(srv.Dir)
	handler := srv.Server.Handler(ctx)

	uid, err := srv.AddDummyLambda(ctx, "cat", "-")
	assert.NoError(t, err)
	err = srv.Server.Queues.Add(application.Queue{
		Name:           "my-queue",
		Target:         uid,
		Retry:          1,
		MaxElementSize: 1024,
	})
	assert.NoError(t, err)

	_, err = srv.Server.Policies.Create("temp", application.PolicyDefinition{
		Public: false,
	})
	assert.NoError(t, err)
	err = srv.Server.Policies.Apply(uid, "temp")

	rr := httptest.NewRecorder()

	req, err := http.NewRequest(http.MethodPost, "https://example.com/q/my-queue", bytes.NewBufferString("hello"))
	assert.NoError(t, err)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusForbidden, rr.Code)
}
