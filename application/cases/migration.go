package cases

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/reddec/trusted-cgi/application"
	"github.com/reddec/trusted-cgi/internal"
	"github.com/reddec/trusted-cgi/types"
)

type legacyManifestPart struct {
	Aliases       types.JsonStringSet `json:"aliases"`
	AllowedIP     types.JsonStringSet `json:"allowed_ip,omitempty"`     // limit incoming connections from list of IP
	AllowedOrigin types.JsonStringSet `json:"allowed_origin,omitempty"` // limit incoming connections by origin header
	Public        bool                `json:"public"`                   // if public, tokens are ignores
	Tokens        map[string]string   `json:"tokens,omitempty"`
}

func (lmr *legacyManifestPart) hasPolicy() bool {
	return len(lmr.AllowedIP) > 0 || len(lmr.AllowedOrigin) > 0 || len(lmr.Tokens) > 0
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
	if !m.hasPolicy() {
		return fn.SetManifest(fn.Manifest())
	}

	policy := application.PolicyDefinition{
		AllowedIP:     m.AllowedIP,
		AllowedOrigin: m.AllowedOrigin,
		Public:        m.Public,
		Tokens:        m.Tokens,
	}
	p, err := impl.policies.Create(uid+"-"+fn.Manifest().Name, policy)
	if err != nil {
		return err
	}
	err = impl.policies.Apply(uid, p.ID)
	if err != nil {
		return err
	}
	return fn.SetManifest(fn.Manifest())
}
