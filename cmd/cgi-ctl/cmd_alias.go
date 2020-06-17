package main

import (
	"context"
	"fmt"
	"github.com/reddec/trusted-cgi/api"
	"github.com/reddec/trusted-cgi/cmd/internal"
	internal2 "github.com/reddec/trusted-cgi/internal"
	"github.com/reddec/trusted-cgi/types"
	"log"
	"os"
)

type alias struct {
	remoteLink
	uidLocator
	Delete bool `short:"d" long:"delete" env:"DELETE" description:"delete links, otherwise add"`
	Keep   bool `long:"keep" env:"KEEP" description:"do not update (if it exists) local manifest file"`
	Args   struct {
		Aliases []string `name:"aliases" positional-arg:"alias" description:"links/aliases names"`
	} `positional-args:"yes"`
}

func (cmd *alias) Execute(args []string) error {
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
	if len(cmd.Args.Aliases) == 0 {
		return cmd.printAliases(ctx, token)
	}
	if cmd.Delete {
		return cmd.removeAliases(ctx, token)
	}
	var app *types.App
	for _, alias := range cmd.Args.Aliases {
		log.Println("adding alias", alias)
		a, err := cmd.Lambdas().Link(ctx, token, cmd.UID, alias)
		if err != nil {
			return fmt.Errorf("add alias %s: %w", alias, err)
		}
		app = a
	}
	return cmd.hintUpdate(app)
}

func (cmd *alias) printAliases(ctx context.Context, token *api.Token) error {
	info, err := cmd.Lambdas().Info(ctx, token, cmd.UID)
	if err != nil {
		return fmt.Errorf("list aliases: %w", err)
	}
	if len(info.Manifest.Aliases) > 0 {
		for name := range info.Manifest.Aliases {
			fmt.Println(name)
		}
	} else {
		log.Println("no available aliases")
	}
	return nil
}

func (cmd *alias) removeAliases(ctx context.Context, token *api.Token) error {
	var app *types.App
	for _, alias := range cmd.Args.Aliases {
		log.Println("removing alias", alias)
		a, err := cmd.Lambdas().Unlink(ctx, token, alias)
		if err != nil {
			return fmt.Errorf("remove alias %s: %w", alias, err)
		}
		app = a
	}
	return cmd.hintUpdate(app)
}

func (cmd *alias) hintUpdate(app *types.App) error {
	if cmd.Keep || app == nil {
		log.Println("done. do not forget update manifest: cgi-ctl update manifest")
		return nil
	}
	if _, err := os.Stat(internal2.ManifestFile); err != nil {
		return nil
	}
	err := app.Manifest.SaveAs(internal2.ManifestFile)
	if err != nil {
		return fmt.Errorf("update manifest: %w", err)
	}
	log.Println("done")
	return nil
}
