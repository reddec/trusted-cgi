package application_test

import (
	"bytes"
	"context"
	"github.com/google/uuid"
	"github.com/reddec/trusted-cgi/application"
	"github.com/reddec/trusted-cgi/application/lambda"
	"github.com/reddec/trusted-cgi/application/platform"
	"github.com/reddec/trusted-cgi/internal"
	"github.com/reddec/trusted-cgi/stats/impl/memlog"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestHandlerByUID(t *testing.T) {
	d, err := ioutil.TempDir("", "test-adapters-")
	if !assert.NoError(t, err) {
		return
	}
	defer os.RemoveAll(d)
	base, err := platform.New(filepath.Join(d, internal.ProjectManifest))
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

	handler := application.HandlerByUID(context.Background(), memlog.New(1), base)

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
