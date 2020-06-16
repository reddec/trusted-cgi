package templates

import (
	"context"
	"encoding/json"
	"github.com/reddec/trusted-cgi/types"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func Read(filename string) (*Template, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var t = &Template{}

	return t, json.NewDecoder(f).Decode(t)
}

type Template struct {
	Description string            `json:"description" yaml:"description"`
	Manifest    types.Manifest    `json:"manifest" yaml:"manifest"`               // manifest to copy
	PostClone   string            `json:"post_clone,omitempty" yaml:"post_clone"` // action (make target) name that should be invoked after clone
	Check       [][]string        `json:"check,omitempty" yaml:"check,omitempty"` // check availability (one line - one check)
	Files       map[string]string `json:"files" yaml:"files,omitempty"`           //only for embedded
}

func (t *Template) IsAvailable(ctx context.Context) bool {
	for _, check := range t.Check {
		cmd := exec.CommandContext(ctx, check[0], check[1:]...)
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setpgid:   true,
			Pdeathsig: syscall.SIGINT,
		}
		if cmd.Run() != nil {
			return false
		}
	}
	return true
}

// List embedded and external templates
func List(templatesDir string) (map[string]*Template, error) {
	merged := ListEmbedded()
	ext, err := ListDir(templatesDir)
	if err != nil {
		return nil, err
	}
	for name, t := range ext {
		merged[name] = t
	}
	return merged, nil
}

func ListDir(dir string) (map[string]*Template, error) {
	const suffix = ".json"
	items, err := ioutil.ReadDir(dir)
	if os.IsNotExist(err) {
		return map[string]*Template{}, nil
	} else if err != nil {
		return nil, err
	}
	var ans = make(map[string]*Template)
	for _, item := range items {
		name := item.Name()
		if item.IsDir() || !strings.HasSuffix(name, suffix) {
			continue
		}
		name = name[:len(name)-len(suffix)]
		t, err := Read(filepath.Join(dir, item.Name()))
		if err != nil {
			return nil, err
		}
		ans[name] = t
	}
	return ans, nil
}

func ListEmbedded() map[string]*Template {
	return map[string]*Template{
		"Python": {
			Description: "Python basic function",
			Check: [][]string{
				{"which", "make"},
				{"which", "python3"},
				{"python3", "-m", "venv", "--help"},
			},
			Files: map[string]string{
				"app.py":           pythonScript,
				"Makefile":         pythonMake,
				"requirements.txt": "requests",
				".cgiignore":       "venv",
			},
			Manifest: types.Manifest{
				Name: "Example Python Function",
				Description: `### Usage

    curl --data-binary '{"name": "reddec"}' -H 'Content-Type: application/json' "http://example.com/a/xyz"

Replace url to the real
`,
				Run:            []string{"./venv/bin/python3", "app.py"},
				TimeLimit:      types.JsonDuration(time.Second),
				Public:         true,
				MaximumPayload: 8192,
				OutputHeaders: map[string]string{
					"Content-Type": "application/json",
				},
			},
			PostClone: "install",
		},
		"Node JS": {
			Description: "Node JS basic function",
			Check: [][]string{
				{"which", "make"},
				{"which", "node"},
				{"which", "npm"},
			},
			Files: map[string]string{
				"app.js":       nodeJsScript,
				"package.json": nodeJsManifest,
				"Makefile":     nodeJsMake,
				".cgiignore":   "node_modules",
			},
			Manifest: types.Manifest{
				Name: "Example NodeJS Function",
				Description: `### Usage

    curl --data-binary '{"name": "reddec"}' -H 'Content-Type: application/json' "http://example.com/a/xyz"

Replace url to the real
`,
				Run:            []string{"node", "app.js"},
				TimeLimit:      types.JsonDuration(time.Second),
				Public:         true,
				MaximumPayload: 8192,
				OutputHeaders: map[string]string{
					"Content-Type": "application/json",
				},
			},
			PostClone: "install",
		},
		"PHP": {
			Description: "PHP basic function",
			Manifest: types.Manifest{
				Name: "Example PHP Function",
				Description: `### Usage

    curl --data-binary '{"name": "reddec"}' -H 'Content-Type: application/json' "http://example.com/a/xyz"

Replace url to the real
`,
				Run:            []string{"php", "app.php"},
				TimeLimit:      types.JsonDuration(time.Second),
				Public:         true,
				MaximumPayload: 8192,
				OutputHeaders: map[string]string{
					"Content-Type": "application/json",
				},
			},
			Check: [][]string{
				{"which", "php"},
			},
			Files: map[string]string{
				"app.php": phpScript,
			},
		},
		"Nim": {
			Description: "Nim lang basic function",
			Manifest: types.Manifest{
				Name: "Fast python-like function",
				Description: `### Usage

    curl --data-binary '{"name": "reddec"}' -H 'Content-Type: application/json' "http://example.com/a/xyz"

Replace url to the real
`,
				Run:            []string{"./bin/lambda"},
				TimeLimit:      types.JsonDuration(time.Second),
				Public:         true,
				MaximumPayload: 8192,
				OutputHeaders: map[string]string{
					"Content-Type": "application/json",
				},
			},
			PostClone: "build",
			Check: [][]string{
				{"which", "make"},
				{"which", "nim"},
				{"which", "nimble"},
			},
			Files: map[string]string{
				"src/lambda.nim": nimScript,
				"lambda.nimble":  nimbleManifest,
				"Makefile":       nimMake,
			},
		},
	}
}

const pythonScript = `
import sys
import json

request = json.load(sys.stdin)
response = ['hello', 'world']
json.dump(response, sys.stdout)
`

const pythonMake = `
install:
	python3 -m venv venv
	./venv/bin/pip install -r requirements.txt
`

const nodeJsScript = `
async function run(request) {
     return ["hello", "world"];
}

let input = '';
process.stdin.resume();
process.stdin.setEncoding('utf8');
process.stdin.on('data', function (chunk) {
    input += chunk;
});
process.stdin.on('end', function () {
	run(JSON.parse(input)).catch((e)=> {
		return {"error": e + ''};
	}).then((response)=> {
		process.stdout.write(JSON.stringify(response));
	})
});
`

const nodeJsMake = `
install:
	npm install .
`

const nodeJsManifest = `{
  "name": "",
  "version": "1.0.0",
  "description": "",
  "main": "index.js",
  "scripts": {
    "test": "echo \"Error: no test specified\" && exit 1"
  },
  "author": "",
  "license": "",
  "dependencies": {
    "axios": "^0.19.2"
  }
}`

const phpScript = `
<?php
$request = json_decode(stream_get_contents(STDIN));

$response = array("hello", "world");

echo json_encode($response, JSON_PRETTY_PRINT);
?>`

const nimScript = `
import json

let request = stdin.readAll().parseJson()

echo pretty(%*["hello", "world"])
`

const nimbleManifest = `
version       = "0.1.0"
author        = ""
description   = ""
license       = ""
srcDir        = "src"
bin           = @["lambda"]

# Dependencies
requires "nim >= 1.2.0"
`

const nimMake = `
build:
	nimble build
	mkdir -p bin
	mv -f lambda bin/
`
