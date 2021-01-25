package lambda

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/reddec/trusted-cgi/types"
)

func (local *localLambda) ListFiles(path string) ([]types.File, error) {
	path, isLocal := local.resolvePath(local.rootDir, path)
	if !isLocal {
		return nil, fmt.Errorf("non-local file")
	}
	list, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var ans = make([]types.File, len(list))
	for i, item := range list {
		ans[i] = types.File{
			Name: item.Name(),
			Dir:  item.IsDir(),
		}
	}
	return ans, nil
}

func (local *localLambda) ReadFile(path string, output io.Writer) error {
	path, isLocal := local.resolvePath(local.rootDir, path)
	if !isLocal {
		return fmt.Errorf("non-local file")
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(output, f)
	return err
}

func (local *localLambda) WriteFile(path string, input io.Reader) error {
	path, isLocal := local.resolvePath(local.rootDir, path)
	if !isLocal {
		return fmt.Errorf("non-local file")
	}
	if path == local.manifestFile() {
		var manifest types.Manifest
		err := json.NewDecoder(input).Decode(&manifest)
		if err != nil {
			return fmt.Errorf("parse manifest: %w", err)
		}
		return local.SetManifest(manifest)
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, input)
	if err != nil {
		return err
	}
	creds := local.creds
	if creds == nil {
		return nil
	}
	return os.Chown(path, creds.User, creds.Group)
}

func (local *localLambda) EnsureDir(path string) error {
	path, isLocal := local.resolvePath(local.rootDir, path)
	if !isLocal {
		return fmt.Errorf("non-local file")
	}
	err := os.MkdirAll(path, 0600)
	if err != nil {
		return err
	}
	return local.applyFilesOwner()
}

func (local *localLambda) RemoveFile(path string) error {
	path, isLocal := local.resolvePath(local.rootDir, path)
	if !isLocal {
		return fmt.Errorf("non-local file")
	}
	if !local.isRemovable(path) {
		return fmt.Errorf("non-removable file")
	}
	return os.RemoveAll(path)
}

func (local *localLambda) RenameFile(src, dest string) error {
	srcPath, isLocal := local.resolvePath(local.rootDir, src)
	if !isLocal {
		return fmt.Errorf("non-local source file")
	}
	destPath, isLocal := local.resolvePath(local.rootDir, dest)
	if !isLocal {
		return fmt.Errorf("non-local desination file")
	}
	if srcPath == destPath {
		return nil
	}
	if !local.isRemovable(srcPath) {
		return fmt.Errorf("non-removable file")
	}
	return os.Rename(srcPath, destPath)
}

func (local *localLambda) Content(tarball io.Writer) error {
	local.lock.RLock()
	defer local.lock.RUnlock()
	ignore, err := local.readIgnore()
	if err != nil {
		return err
	}
	gz := gzip.NewWriter(tarball)
	defer gz.Close()
	return tarFiles(local.rootDir, gz, ignore)
}

func (local *localLambda) SetContent(tarball io.Reader) error {
	local.lock.Lock()
	defer local.lock.Unlock()
	gz, err := gzip.NewReader(tarball)
	if err != nil {
		return err
	}
	defer gz.Close()
	err = untarFiles(gz, local.rootDir)
	if err != nil {
		return err
	}
	err = local.applyFilesOwner()
	if err != nil {
		return err
	}
	return local.reindex()
}

func (local *localLambda) applyFilesOwner() error {
	if local.creds == nil {
		return nil
	}
	return filepath.Walk(local.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		return os.Chown(path, local.creds.User, local.creds.Group)
	})
}

func (local *localLambda) resolvePath(rootDir string, path string) (string, bool) {
	path = filepath.Join(rootDir, path)
	abs, err := filepath.Abs(path)
	if err != nil {
		return path, false
	}
	path = abs

	if path == rootDir {
		return path, true
	}
	return path, strings.HasPrefix(path, rootDir+string(filepath.Separator))
}

func (local *localLambda) isRemovable(path string) bool {
	if path == local.manifestFile() {
		return false
	}
	return true
}
