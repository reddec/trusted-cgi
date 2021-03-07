package local

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/reddec/trusted-cgi/internal"
	"github.com/reddec/trusted-cgi/types"
)

// Forget known lambda definition - remove from cache. Could be re-scanned by ScanAll or Scan.
func (mgr *LambdaManager) Forget(uid string) {
	mgr.lambdas.lock.Lock()
	defer mgr.lambdas.lock.Unlock()
	delete(mgr.lambdas.data, uid)
}

// Scan single lambda by UID.
func (mgr *LambdaManager) Scan(uid string) error {
	if filepath.Base(uid) != uid {
		// should be just a dir
		return fmt.Errorf("malformed UID")
	}

	def, err := mgr.scanDefinition(filepath.Join())
	if err != nil {
		return fmt.Errorf("scan definition: %w", err)
	}

	mgr.lambdas.lock.Lock()
	defer mgr.lambdas.lock.Unlock()
	if mgr.lambdas.data == nil {
		mgr.lambdas.data = make(map[string]*lambdaDefinition)
	}
	mgr.lambdas.data[uid] = def
	return nil
}

// Scan all lambdas in root directory.
func (mgr *LambdaManager) ScanAll() error {
	list, err := ioutil.ReadDir(mgr.rootDir)
	if err != nil {
		return fmt.Errorf("list root dir %s: %w", mgr.rootDir, err)
	}
	var lambdas = make(map[string]*lambdaDefinition)

	for _, item := range list {
		if !item.IsDir() {
			continue
		}
		uid := item.Name()

		def, err := mgr.scanDefinition(uid)
		if err != nil {
			log.Println("failed inspect lambda", uid, "-", err)
			continue
		}

		lambdas[uid] = def
	}

	mgr.lambdas.lock.Lock()
	defer mgr.lambdas.lock.Unlock()
	mgr.lambdas.data = lambdas
	return nil
}

func (mgr *LambdaManager) scanDefinition(uid string) (*lambdaDefinition, error) {
	lambdaDir := filepath.Join(mgr.rootDir, uid)

	var manifest types.Manifest
	err := manifest.LoadFrom(filepath.Join(lambdaDir, internal.ManifestFile))

	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}

	return &lambdaDefinition{
		uid:      uid,
		manager:  mgr,
		manifest: manifest,
		rootDir:  lambdaDir,
	}, nil
}
