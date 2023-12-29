package lambda

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/reddec/trusted-cgi/application"
	"github.com/reddec/trusted-cgi/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_tar(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if !assert.NoError(t, err) {
		return
	}
	defer os.RemoveAll(dir)
	t.Log(dir)

	err = os.MkdirAll(filepath.Join(dir, "test"), 0755)
	if !assert.NoError(t, err) {
		return
	}

	err = ioutil.WriteFile(filepath.Join(dir, "test", "test.txt"), []byte("hello"), 0755)
	if !assert.NoError(t, err) {
		return
	}

	var buffer bytes.Buffer

	err = tarFiles(dir, &buffer, []string{})
	if !assert.NoError(t, err) {
		return
	}

	reader := tar.NewReader(&buffer)
	a_dir, err := reader.Next()
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, tar.TypeDir, int32(a_dir.Typeflag))
	assert.Equal(t, "test", a_dir.Name)
	a_file, err := reader.Next()
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, tar.TypeReg, int32(a_file.Typeflag))
	assert.Equal(t, "test/test.txt", a_file.Name)
	content, err := ioutil.ReadAll(reader)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "hello", string(content))
}

func TestLocalLambda_Content(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if !assert.NoError(t, err) {
		return
	}
	defer os.RemoveAll(dir)
	t.Log(dir)

	err = os.MkdirAll(filepath.Join(dir, "test"), 0755)
	if !assert.NoError(t, err) {
		return
	}

	err = ioutil.WriteFile(filepath.Join(dir, "test", "test.txt"), []byte("hello"), 0755)
	if !assert.NoError(t, err) {
		return
	}

	var buffer bytes.Buffer
	ll := localLambda{rootDir: dir}
	err = ll.SetManifest(types.Manifest{Name: "xxx"})
	if !assert.NoError(t, err) {
		return
	}
	err = ll.Content(&buffer)
	if !assert.NoError(t, err) {
		return
	}

	dir2, err := ioutil.TempDir("", "")
	if !assert.NoError(t, err) {
		return
	}
	defer os.RemoveAll(dir2)
	t.Log(dir2)

	ll2 := localLambda{rootDir: dir2}
	err = ll2.SetContent(&buffer)
	if !assert.NoError(t, err) {
		return
	}

	assert.FileExists(t, filepath.Join(dir2, "test", "test.txt"))
	text, err := ioutil.ReadFile(filepath.Join(dir2, "test", "test.txt"))
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "hello", string(text))
	assert.Equal(t, "xxx", ll2.manifest.Name)
}

func TestLocalLambda_Invoke(t *testing.T) {
	d, err := ioutil.TempDir("", "test-lambda-")
	if !assert.NoError(t, err) {
		return
	}
	defer os.RemoveAll(d)

	fn, err := DummyPublic(d, "cat", "-")
	if !assert.NoError(t, err) {
		return
	}

	timeout, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	var out bytes.Buffer
	err = fn.Invoke(timeout, types.Request{
		Body: ioutil.NopCloser(bytes.NewBufferString("hello world")),
	}, &out, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "hello world", out.String())
}

func TestStaticFile(t *testing.T) {
	d, err := os.MkdirTemp("", "test-lambda-*")
	require.NoError(t, err)
	defer os.RemoveAll(d)

	require.NoError(t, os.Mkdir(filepath.Join(d, "static"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(d, "static", "index.html"), []byte("index page"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(d, "static", "root"), []byte("root page"), 0755))
	require.NoError(t, os.Mkdir(filepath.Join(d, "static", "foo"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(d, "static", "foo", "foo"), []byte("foo page"), 0755))
	require.NoError(t, os.Mkdir(filepath.Join(d, "static", "foo", "bar"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(d, "static", "foo", "bar", "bar"), []byte("bar page"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(d, "static", "foo","index.html"), []byte("sub index page"), 0755))

	fn, err := DummyPublic(d, "cat", "-")
	require.NoError(t, err)

	manifest := fn.Manifest()
	manifest.Static = "static"
	require.NoError(t, fn.SetManifest(manifest))
	require.NoError(t, fn.reindex()) // fixme: usage of private API

	t.Run("index served", func(t *testing.T) {
		content, err := testRequest(fn, http.MethodGet, "/f/", nil)
		require.NoError(t, err)
		assert.Equal(t, "index page", string(content))
	})

	t.Run("index served even without slash", func(t *testing.T) {
		content, err := testRequest(fn, http.MethodGet, "/f", nil)
		require.NoError(t, err)
		assert.Equal(t, "index page", string(content))
	})

	t.Run("root path served", func(t *testing.T) {
		content, err := testRequest(fn, http.MethodGet, "/f/root", nil)
		require.NoError(t, err)
		assert.Equal(t, "root page", string(content))
	})

	t.Run("sub path served", func(t *testing.T) {
		content, err := testRequest(fn, http.MethodGet, "/f/foo/foo", nil)
		require.NoError(t, err)
		assert.Equal(t, "foo page", string(content))
	})
	t.Run("sub sub path served", func(t *testing.T) {
		content, err := testRequest(fn, http.MethodGet, "/f/foo/bar/bar", nil)
		require.NoError(t, err)
		assert.Equal(t, "bar page", string(content))
	})
	t.Run("sub path with index.html served (no trailing slash)", func(t *testing.T) {
		content, err := testRequest(fn, http.MethodGet, "/f/foo", nil)
		require.NoError(t, err)
		assert.Equal(t, "sub index page", string(content))
	})
	t.Run("sub path with index.html served (with trailing slash)", func(t *testing.T) {
		content, err := testRequest(fn, http.MethodGet, "/f/foo/", nil)
		require.NoError(t, err)
		assert.Equal(t, "sub index page", string(content))
	})
	t.Run("sub path with index.html served (with trailing slash + index.html)", func(t *testing.T) {
		content, err := testRequest(fn, http.MethodGet, "/f/foo/index.html", nil)
		require.NoError(t, err)
		assert.Equal(t, "sub index page", string(content))
	})
}

func testRequest(fn application.Invokable, method string, path string, payload []byte) ([]byte, error) {
	timeout, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	var out bytes.Buffer
	err := fn.Invoke(timeout, types.Request{
		Method: method,
		Path:   path,
		Body:   io.NopCloser(bytes.NewReader(payload)),
	}, &out, nil)
	return out.Bytes(), err
}
