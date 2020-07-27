package application_test

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"github.com/reddec/trusted-cgi/application"
	"github.com/reddec/trusted-cgi/application/lambda"
	"github.com/reddec/trusted-cgi/application/platform"
	"github.com/reddec/trusted-cgi/internal"
	"github.com/reddec/trusted-cgi/stats/impl/memlog"
	"github.com/reddec/trusted-cgi/types"

	"github.com/stretchr/testify/assert"
)

type mockValidator struct {
	forbidden map[string]error
}

func (mv *mockValidator) Inspect(lambda string, request *types.Request) error {
	return mv.forbidden[lambda]
}

func TestHandlerByUID(t *testing.T) {
	mc := &mockValidator{}
	d, err := ioutil.TempDir("", "test-adapters-")
	if !assert.NoError(t, err) {
		return
	}
	defer os.RemoveAll(d)
	base, err := platform.New(filepath.Join(d, internal.ProjectManifest), mc)
	if !assert.NoError(t, err) {
		return
	}
	uid := uuid.New().String()
	lpath := filepath.Join(d, uid)
	err = os.MkdirAll(lpath, 0755)
	if !assert.NoError(t, err) {
		return
	}
	t.Log("UID:", uid)

	fn, err := lambda.DummyPublic(lpath, "/usr/bin/cat", "-")
	if !assert.NoError(t, err) {
		return
	}
	err = base.Add(uid, fn)
	if !assert.NoError(t, err) {
		return
	}

	handler := application.HandlerByUID(context.Background(), mc, memlog.New(1), base)

	testServer := httptest.NewServer(handler)
	url := testServer.URL + "/" + uid
	t.Log("URL:", url)
	res, err := testServer.Client().Post(url, "", bytes.NewBufferString("hello world"))
	if !assert.NoError(t, err) {
		return
	}
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)
	data, err := ioutil.ReadAll(res.Body)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "hello world", string(data))
}

func TestHandlerByUID_forbidden(t *testing.T) {
	mc := &mockValidator{
		forbidden: map[string]error{},
	}
	d, err := ioutil.TempDir("", "test-adapters-")
	if !assert.NoError(t, err) {
		return
	}
	defer os.RemoveAll(d)
	base, err := platform.New(filepath.Join(d, internal.ProjectManifest), mc)
	if !assert.NoError(t, err) {
		return
	}
	uid := uuid.New().String()
	mc.forbidden[uid] = fmt.Errorf("not allowed")
	lpath := filepath.Join(d, uid)
	err = os.MkdirAll(lpath, 0755)
	if !assert.NoError(t, err) {
		return
	}
	t.Log("UID:", uid)

	fn, err := lambda.DummyPublic(lpath, "/usr/bin/cat", "-")
	if !assert.NoError(t, err) {
		return
	}
	err = base.Add(uid, fn)
	if !assert.NoError(t, err) {
		return
	}

	handler := application.HandlerByUID(context.Background(), mc, memlog.New(1), base)

	testServer := httptest.NewServer(handler)
	url := testServer.URL + "/" + uid
	t.Log("URL:", url)
	res, err := testServer.Client().Post(url, "", bytes.NewBufferString("hello world"))
	if !assert.NoError(t, err) {
		return
	}
	defer res.Body.Close()
	assert.Equal(t, http.StatusForbidden, res.StatusCode)
}
