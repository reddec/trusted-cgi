package main

import (
	"fmt"
	"github.com/reddec/trusted-cgi/cmd/internal"
	internal2 "github.com/reddec/trusted-cgi/internal"
	"log"
	"os"
	"path/filepath"
)

type updateManifest struct {
	remoteLink
	UID string `short:"o" long:"uid" env:"UID" description:"Lambda UID (if empty - dirname of input will be used)"`
}

func (cmd *updateManifest) Execute(args []string) error {
	ctx, closer := internal.SignalContext()
	defer closer()
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("detect work dir: %w", err)
	}
	if cmd.UID == "" {
		cmd.UID = filepath.Base(wd)
	}
	log.Println("lambda", cmd.UID)
	log.Println("login...")
	token, err := cmd.Token(ctx)
	if err != nil {
		return fmt.Errorf("login: %w", err)
	}
	log.Println("getting info...")
	info, err := cmd.Lambdas().Info(ctx, token, cmd.UID)
	if err != nil {
		return fmt.Errorf("get info: %w", err)
	}
	log.Println("saving...")
	err = info.Manifest.SaveAs(internal2.ManifestFile)
	if err != nil {
		return fmt.Errorf("update manifest file: %w", err)
	}
	log.Println("done")
	return nil
}
