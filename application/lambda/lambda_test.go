package lambda

import (
	"archive/tar"
	"bytes"
	"context"
	"github.com/reddec/trusted-cgi/types"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
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

	fn, err := DummyPublic(d, "/usr/bin/cat", "-")
	if !assert.NoError(t, err) {
		return
	}

	timeout, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	var out bytes.Buffer
	err = fn.Invoke(timeout, types.Request{
		Body: bytes.NewBufferString("hello world"),
	}, &out, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "hello world", out.String())
}
