package application

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"syscall"
	"time"
)

var targetsPattern = regexp.MustCompile(`^([\d\w-/]+)\s*:\s*[\d\w-/\s]*$`)

// List Make actions (if Makefile defined)
func (app *App) ListActions() ([]string, error) {
	makefile := filepath.Join(app.location, "Makefile")
	f, err := os.Open(makefile)
	if os.IsNotExist(err) {
		return []string{}, nil
	} else if err != nil {
		return nil, err
	}
	defer f.Close()

	var ans = make([]string, 0)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		matches := targetsPattern.FindAllStringSubmatch(scanner.Text(), -1)
		if len(matches) < 1 || len(matches[0]) != 2 {
			continue
		}
		ans = append(ans, matches[0][1])
	}
	return ans, nil
}

// Invoke action by name (make target)
func (app *App) InvokeAction(ctx context.Context, name string, timeLimit time.Duration, globalEnv map[string]string) (string, error) {
	var out bytes.Buffer

	if timeLimit > 0 {
		cctx, cancel := context.WithTimeout(ctx, timeLimit)
		defer cancel()
		ctx = cctx
	}
	environments := os.Environ()
	for k, v := range globalEnv {
		environments = append(environments, k+"="+v)
	}
	for k, v := range app.Manifest.Environment {
		environments = append(environments, k+"="+v)
	}

	cmd := exec.CommandContext(ctx, "make", name)
	cmd.Dir = app.location
	cmd.Stdout = &out
	cmd.Stderr = &out
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig:  syscall.SIGINT,
		Setpgid:    true,
		Credential: app.creds,
	}
	cmd.Env = environments

	err := cmd.Run()
	return out.String(), err
}
