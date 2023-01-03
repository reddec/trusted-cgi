package workspace

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
)

func NewFileCache(rootDir string) (*FileCache, error) {
	if err := os.MkdirAll(rootDir, 0755); err != nil {
		return nil, err
	}
	return &FileCache{rootDir: rootDir}, nil
}

type FileCache struct {
	rootDir string
}

func (fc *FileCache) Write(data io.Reader) (string, error) {
	f, err := os.CreateTemp(fc.rootDir, "")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	_, err = io.Copy(f, data)
	if err != nil {
		_ = f.Close()
		_ = os.RemoveAll(f.Name())
		return "", fmt.Errorf("write temp file: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = os.RemoveAll(f.Name())
		return "", fmt.Errorf("close temp file: %w", err)
	}
	return filepath.Base(f.Name()), nil
}

func (fc *FileCache) Open(s string) (io.ReadCloser, error) {
	return os.Open(filepath.Join(fc.rootDir, path.Clean(s)))
}

func (fc *FileCache) Remove(s string) error {
	return os.RemoveAll(filepath.Join(fc.rootDir, path.Clean(s)))
}
