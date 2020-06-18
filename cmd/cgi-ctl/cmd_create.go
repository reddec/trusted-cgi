package main

import (
	"fmt"
	"github.com/reddec/trusted-cgi/cmd/internal"
	internal_app "github.com/reddec/trusted-cgi/internal"
	"log"
	"os"
	"path/filepath"
)

type create struct {
	remoteLink
	Public      bool   `long:"public" env:"PUBLIC" description:"make public lambda"`
	Description string `short:"d" long:"description" env:"DESCRIPTION" description:"lambda description"`
	Args        struct {
		Dir string `name:"dir" description:"project directory" required:"yes"`
	} `positional-args:"yes"`
}

func (cmd *create) Execute(args []string) error {
	ctx, closer := internal.SignalContext()
	defer closer()

	err := os.MkdirAll(cmd.Args.Dir, 0755)
	if err != nil {
		return fmt.Errorf("prepare directory: %w", err)
	}

	err = os.Chdir(cmd.Args.Dir)
	if err != nil {
		return fmt.Errorf("change dir: %w", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get work dir: %w", err)
	}

	log.Println("login...")
	token, err := cmd.Token(ctx)
	if err != nil {
		return fmt.Errorf("login: %w", err)
	}

	log.Println("creating...")
	info, err := cmd.Project().Create(ctx, token)
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}
	log.Println("created", info.UID)
	log.Println("saving info....")

	var cf controlFile
	cf.URL = cmd.URL
	cf.UID = info.UID
	err = cf.Save(controlFilename)
	if err != nil {
		return fmt.Errorf("save control file: %w", err)
	}
	err = appendIfNoLineFile(internal_app.CGIIgnore, controlFilename)
	if err != nil {
		return fmt.Errorf("update cgiignore file: %w", err)
	}

	info.Manifest.Name = filepath.Base(wd)
	info.Manifest.Description = cmd.Description
	info.Manifest.Public = cmd.Public

	log.Println("updating manifest...")
	info, err = cmd.Lambdas().Update(ctx, token, info.UID, info.Manifest)
	if err != nil {
		return fmt.Errorf("update manifest: %w", err)
	}

	err = info.Manifest.SaveAs(internal_app.ManifestFile)
	if err != nil {
		return fmt.Errorf("save manifest: %w", err)
	}
	log.Println("done")
	return nil
}
