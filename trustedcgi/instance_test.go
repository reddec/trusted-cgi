package trustedcgi_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/reddec/trusted-cgi/templates"
	"github.com/reddec/trusted-cgi/trustedcgi"
	"github.com/reddec/trusted-cgi/types"
)

func createTemp() (*trustedcgi.Instance, error) {
	dir, err := ioutil.TempDir("", "trusted-cgi-*")
	if err != nil {
		return nil, err
	}
	return trustedcgi.Default().Directory(dir).SSH(false).New()
}

func destroy(inst *trustedcgi.Instance) {
	inst.Stop()
	_ = os.RemoveAll(inst.Location)
}

func TestDefault_run(t *testing.T) {
	inst, err := createTemp()
	if !assert.NoError(t, err) {
		return
	}
	defer destroy(inst)

	uid, err := inst.Server().Cases.CreateFromTemplate(inst.Context(), templates.Template{
		Manifest: types.Manifest{
			Run: []string{"cat", "-"},
		},
	})
	assert.NoError(t, err)
	handler := inst.Handler()
	t.Run("200 on exists", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/a/"+uid, bytes.NewBufferString("hello world"))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "hello world", rec.Body.String())
	})
	t.Run("404 on not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/a/"+uuid.New().String(), bytes.NewBufferString("hello world"))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}
