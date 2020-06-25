package main

import (
	"context"
	"fmt"
	"github.com/reddec/trusted-cgi/api"
	"github.com/reddec/trusted-cgi/cmd/internal"
	"log"
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
	for _, alias := range cmd.Args.Aliases {
		log.Println("adding alias", alias)
		_, err := cmd.Lambdas().Link(ctx, token, cmd.UID, alias)
		if err != nil {
			return fmt.Errorf("add alias %s: %w", alias, err)
		}
	}
	return nil
}

func (cmd *alias) printAliases(ctx context.Context, token *api.Token) error {
	info, err := cmd.Lambdas().Info(ctx, token, cmd.UID)
	if err != nil {
		return fmt.Errorf("list aliases: %w", err)
	}
	if len(info.Aliases) > 0 {
		for name := range info.Aliases {
			fmt.Println(name)
		}
	} else {
		log.Println("no available aliases")
	}
	return nil
}

func (cmd *alias) removeAliases(ctx context.Context, token *api.Token) error {
	for _, alias := range cmd.Args.Aliases {
		log.Println("removing alias", alias)
		_, err := cmd.Lambdas().Unlink(ctx, token, alias)
		if err != nil {
			return fmt.Errorf("remove alias %s: %w", alias, err)
		}
	}
	return nil
}
