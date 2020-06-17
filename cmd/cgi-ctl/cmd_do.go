package main

import (
	"fmt"
	"github.com/reddec/trusted-cgi/cmd/internal"
	"log"
	"os"
	"path/filepath"
)

type do struct {
	remoteLink
	UID  string `short:"o" long:"uid" env:"UID" description:"Lambda UID (if empty - dirname of input will be used)"`
	Args struct {
		Actions []string `positional-arg:"yes" name:"action" description:"action names"`
	} `positional-args:"yes"`
}

func (cmd *do) Execute(args []string) error {
	ctx, closer := internal.SignalContext()
	defer closer()

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("detect work dir: %w", err)
	}
	if cmd.UID == "" {
		cmd.UID = filepath.Base(wd)
	}
	log.Println("login...")
	token, err := cmd.Token(ctx)
	if err != nil {
		return fmt.Errorf("login: %w", err)
	}
	log.Println("lambda", cmd.UID)

	if len(cmd.Args.Actions) == 0 {
		list, err := cmd.Lambdas().Actions(ctx, token, cmd.UID)
		if err != nil {
			return fmt.Errorf("list actions: %w", err)
		}
		if len(list) > 0 {
			for _, name := range list {
				fmt.Println(name)
			}
		} else {
			log.Println("no available actions")
		}
		return nil
	}

	for _, action := range cmd.Args.Actions {
		log.Println("invoking", action, "...")
		out, err := cmd.Lambdas().Invoke(ctx, token, cmd.UID, action)
		log.Println("response:", out)
		if err != nil {
			return fmt.Errorf("invoke %s: %w", action, err)
		}
	}
	log.Println("done")
	return nil
}
