package main

import (
	"fmt"
	"github.com/reddec/trusted-cgi/cmd/internal"
	internal2 "github.com/reddec/trusted-cgi/internal"
	"log"
)

type updateManifest struct {
	remoteLink
	uidLocator
}

func (cmd *updateManifest) Execute(args []string) error {
	ctx, closer := internal.SignalContext()
	defer closer()
	if err := cmd.parseUID(); err != nil {
		return err
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
