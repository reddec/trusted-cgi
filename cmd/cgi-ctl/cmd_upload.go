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
	UID   string `short:"o" long:"uid" env:"UID" description:"Lambda UID" required:"yes"`
	Input string `long:"input" env:"INPUT" description:"Directory" default:"."`
}

func (cmd *upload) Execute([]string) error {
	ctx, closer := internal.SignalContext()
	defer closer()

	var buffer = &bytes.Buffer{}
	log.SetOutput(os.Stderr)
	log.Println("archiving...")
	var args = []string{"zcf", "-"}
	ignoreFile := filepath.Join(cmd.Input, application.CGIIgnore)
	if _, err := os.Stat(ignoreFile); err == nil {
		args = append(args, "--exclude-from", ignoreFile)
	}
	args = append(args, cmd.Input)
	run := exec.CommandContext(ctx, "tar", args...)
	run.Stdout = buffer
	run.Stderr = os.Stderr
	err := run.Run()
	if err != nil {
		return fmt.Errorf("archive: %w", err)
	}
	log.Println("login...")
	token, err := cmd.Token(ctx)
	if err != nil {
		return fmt.Errorf("login: %w", err)
	}
	log.Println("upload...")
	_, err = cmd.Lambdas().Upload(ctx, token, cmd.UID, buffer.Bytes())
	if err != nil {
		return fmt.Errorf("upload: %w", err)
	}
	log.Println("done")
	return nil
}
