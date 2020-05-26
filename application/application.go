package application

import (
	"fmt"
	"github.com/reddec/trusted-cgi/types"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

const (
	ManifestFile = "manifest.json"
)

func OpenApp(location string, creds *syscall.Credential) (*App, error) {
	var app = &App{
		UID:      filepath.Base(location),
		creds:    creds,
		location: location,
	}
	return app, app.Manifest.LoadFrom(app.ManifestFile())
}

func CreateApp(location string, creds *syscall.Credential, manifest types.Manifest) (*App, error) {
	var app = &App{
		UID:      filepath.Base(location),
		Manifest: manifest,
		creds:    creds,
		location: location,
	}
	err := os.MkdirAll(location, 0755)
	if err != nil {
		return nil, err
	}
	err = manifest.SaveAs(app.ManifestFile())
	if err != nil {
		return nil, err
	}
	return app, app.ApplyOwner()
}

type App struct {
	UID      string              `json:"uid"`
	Manifest types.Manifest      `json:"manifest"`
	creds    *syscall.Credential `json:"-"`
	location string              `json:"-"`
}

func (app *App) ManifestFile() string {
	return filepath.Join(app.location, ManifestFile)
}

func (app *App) ApplyOwner() error {
	if app.creds == nil {
		return nil
	}
	uid := int(app.creds.Uid)
	gid := int(app.creds.Gid)
	return filepath.Walk(app.location, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		return os.Chown(path, uid, gid)
	})
}

func (app *App) File(filename string) (string, error) {
	fp, err := filepath.Abs(filepath.Join(app.location, filename))
	if err != nil {
		return "", err
	}
	root, err := filepath.Abs(app.location)
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(app.location+string(filepath.Separator), root) {
		return "", fmt.Errorf("path is not belongs to application")
	}
	return fp, nil
}

func (app *App) WriteFile(filename string, content []byte) error {
	f, err := app.File(filename)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(f, content, 0755)
	if err != nil {
		return err
	}
	if app.creds == nil {
		return nil
	}
	return os.Chown(f, int(app.creds.Uid), int(app.creds.Gid))
}

func (app *App) ReadFile(filename string) ([]byte, error) {
	f, err := app.File(filename)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadFile(f)
}

func (app *App) Touch(filename string, dir bool) error {
	if !dir {
		return app.WriteFile(filename, []byte{})
	}
	f, err := app.File(filename)
	if err != nil {
		return err
	}
	err = os.Mkdir(f, 0755)
	if err != nil {
		return err
	}
	if app.creds == nil {
		return nil
	}
	return os.Chown(f, int(app.creds.Uid), int(app.creds.Gid))
}
