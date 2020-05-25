package application

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

const (
	// modified version from https://stackoverflow.com/a/26339924/1195316 that sounds like dark magic but working
	scriptListMakeTargets = `make -pRrq 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($1 !~ "^[#.]") {print $1}}' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@$'`
)

// List Make actions (if Makefile defined)
func (app *App) ListActions(ctx context.Context) ([]string, error) {
	makefile := filepath.Join(app.location, "Makefile")
	if _, err := os.Stat(makefile); err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	var buffer bytes.Buffer

	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", scriptListMakeTargets)
	cmd.Dir = app.location
	cmd.Stdout = &buffer
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig:  syscall.SIGINT,
		Setpgid:    true,
		Credential: app.creds,
	}

	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	var ans = make([]string, 0)
	for _, line := range strings.Split(buffer.String(), "\n") {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		ans = append(ans, line)
	}
	return ans, nil
}

// Invoke action by name (make target)
func (app *App) InvokeAction(ctx context.Context, name string) error {
	cmd := exec.CommandContext(ctx, "make", name)
	cmd.Dir = app.location
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig:  syscall.SIGINT,
		Setpgid:    true,
		Credential: app.creds,
	}
	return cmd.Run()
}
