package main

import (
	"fmt"
	"github.com/reddec/trusted-cgi/cmd/internal"
	internal2 "github.com/reddec/trusted-cgi/internal"
	"github.com/reddec/trusted-cgi/types"
	"log"
)

type apply struct {
	remoteLink
	uidLocator
}

func (cmd *apply) Execute(args []string) error {
	ctx, closer := internal.SignalContext()
	defer closer()
	if err := cmd.parseUID(); err != nil {
		return err
	}
	log.Println("lambda", cmd.UID)

	var manifest types.Manifest
	if err := manifest.LoadFrom(internal2.ManifestFile); err != nil {
		return fmt.Errorf("load local manifest: %w", err)
	}

	log.Println("login...")
	token, err := cmd.Token(ctx)
	if err != nil {
		return fmt.Errorf("login: %w", err)
	}
	log.Println("pushing manifest...")
	_, err = cmd.Lambdas().Update(ctx, token, cmd.UID, manifest)
	if err != nil {
		return fmt.Errorf("update remote manifest: %w", err)
	}
	log.Println("done")
	return nil
}
