package main

import (
	"bytes"
	"fmt"
	"github.com/reddec/trusted-cgi/application"
	"github.com/reddec/trusted-cgi/cmd/internal"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

type upload struct {
	remoteLink
	UID   string `short:"o" long:"uid" env:"UID" description:"Lambda UID (if empty - dirname of input will be used)"`
	Input string `long:"input" env:"INPUT" description:"Directory" default:"."`
}

func (cmd *upload) Execute([]string) error {
	ctx, closer := internal.SignalContext()
	defer closer()

	wd, err := filepath.Abs(cmd.Input)
	if err != nil {
		return fmt.Errorf("detect abs path: %w", err)
	}
	if cmd.UID == "" {
		cmd.UID = filepath.Base(wd)
	}

	err = os.Chdir(wd)
	if err != nil {
		return fmt.Errorf("change dir: %w", err)
	}

	var buffer = &bytes.Buffer{}
	log.SetOutput(os.Stderr)
	log.Println("archiving...")
	var args = []string{"zcf", "-"}
	if _, err := os.Stat(application.CGIIgnore); err == nil {
		args = append(args, "--exclude-from", application.CGIIgnore)
	}
	args = append(args, ".")
	run := exec.CommandContext(ctx, "tar", args...)
	run.Stdout = buffer
	run.Stderr = os.Stderr
	err = run.Run()
	if err != nil {
		return fmt.Errorf("archive: %w", err)
	}
	log.Println("login...")
	token, err := cmd.Token(ctx)
	if err != nil {
		return fmt.Errorf("login: %w", err)
	}
	log.Println("upload", cmd.UID, buffer.Len()/1024, "KB ...")
	_, err = cmd.Lambdas().Upload(ctx, token, cmd.UID, buffer.Bytes())
	if err != nil {
		return fmt.Errorf("upload: %w", err)
	}
	log.Println("done")
	return nil
}
