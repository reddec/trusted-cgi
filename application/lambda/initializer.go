package lambda

import (
	"bytes"
	"context"
	"fmt"
	"github.com/reddec/trusted-cgi/internal"
	"github.com/reddec/trusted-cgi/templates"
	"github.com/reddec/trusted-cgi/types"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

// Create dummy public lambda in defined path with manifest based on execution specified binary with args
func DummyPublic(path string, bin string, args ...string) (*localLambda, error) {
	var manifest = types.Manifest{
		Run: append([]string{bin}, args...),
	}
	err := manifest.SaveAs(filepath.Join(path, internal.ManifestFile))
	if err != nil {
		return nil, fmt.Errorf("create manifest: %w", err)
	}
	return FromDir(path)
}

// Load lambda definition from directory
func FromDir(path string) (*localLambda, error) {
	ll := &localLambda{rootDir: path}
	return ll, ll.reindex()
}

// Clone lambda definition to directory and load
func FromGit(ctx context.Context, privateKey, repo string, path string) (*localLambda, error) {
	cmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", repo, path)
	internal.SetFlags(cmd)
	var buffer bytes.Buffer
	cmd.Stderr = &buffer
	cmd.Stdout = os.Stdout
	cmd.Env = append(os.Environ(), "GIT_SSH_COMMAND=ssh -i "+privateKey)
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", buffer.String(), err)
	}
	return FromDir(path)
}

func FromTemplate(ctx context.Context, template templates.Template, path string) (*localLambda, error) {
	err := template.Manifest.SaveAs(filepath.Join(path, internal.ManifestFile))
	if err != nil {
		return nil, fmt.Errorf("write manifest: %w", err)
	}
	for fileName, content := range template.Files {
		destFile := filepath.Join(path, fileName)
		err := ioutil.WriteFile(destFile, []byte(content), 0755)
		if err != nil {
			return nil, fmt.Errorf("write file %s content: %w", fileName, err)
		}
	}

	lambda, err := FromDir(path)
	if err != nil {
		return nil, err
	}
	if template.PostClone != "" {
		err := lambda.Do(ctx, template.PostClone, 0, nil, nil)
		if err != nil {
			return nil, fmt.Errorf("invoke post-clone %s: %w", template.PostClone, err)
		}
	}
	return lambda, nil

}
