package cases

import (
	"encoding/json"
	"github.com/reddec/trusted-cgi/application"
	"github.com/reddec/trusted-cgi/internal"
	"github.com/reddec/trusted-cgi/types"
	"os"
	"path/filepath"
)

type legacyManifestPart struct {
	Aliases types.JsonStringSet `json:"aliases"`
}

func (lmr *legacyManifestPart) Read(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(lmr)
}

func (impl *casesImpl) applyMigration(uid, path string, fn application.Lambda) error {
	var m legacyManifestPart
	err := m.Read(filepath.Join(path, internal.ManifestFile))
	if err != nil {
		return err
	}
	for alias := range m.Aliases {
		_, err = impl.platform.Link(uid, alias)
		if err != nil {
			return err
		}
	}
	return fn.SetManifest(fn.Manifest())
}
