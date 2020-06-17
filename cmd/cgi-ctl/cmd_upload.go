package main

import (
	"bytes"
	"fmt"
	"github.com/alecthomas/units"
	"github.com/reddec/trusted-cgi/cmd/internal"
	internal_app "github.com/reddec/trusted-cgi/internal"
	"log"
	"os"
	"os/exec"
)

type upload struct {
	remoteLink
	uidLocator
	Input string `long:"input" env:"INPUT" description:"Directory" default:"."`
}

func (cmd *upload) Execute([]string) error {
	ctx, closer := internal.SignalContext()
	defer closer()
	err := os.Chdir(cmd.Input)
	if err != nil {
		return fmt.Errorf("change dir: %w", err)
	}
	if err := cmd.parseUID(); err != nil {
		return err
	}
	var buffer = &bytes.Buffer{}
	log.SetOutput(os.Stderr)
	log.Println("archiving...")
	var args = []string{"zcf", "-"}
	if _, err := os.Stat(internal_app.CGIIgnore); err == nil {
		args = append(args, "--exclude-from", internal_app.CGIIgnore)
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
	log.Println("upload", cmd.UID, units.Base2Bytes(buffer.Len()), "...")
	_, err = cmd.Lambdas().Upload(ctx, token, cmd.UID, buffer.Bytes())
	if err != nil {
		return fmt.Errorf("upload: %w", err)
	}
	log.Println("done")
	return nil
}
