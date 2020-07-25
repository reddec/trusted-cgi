package main

import (
	"github.com/reddec/trusted-cgi/cmd/internal"
	internal_app "github.com/reddec/trusted-cgi/internal"
	"github.com/reddec/trusted-cgi/types"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type Bare struct {
	Git         bool          `long:"git" env:"GIT" description:"Enable Git"`
	Description string        `short:"d" long:"description" env:"DESCRIPTION" description:"Description" default:"Bare project"`
	TimeLimit   time.Duration `short:"t" long:"time-limit" env:"TIME_LIMIT" description:"Time limit for execution" default:"10s"`
	MaxPayload  int64         `short:"p" long:"max-payload" env:"MAX_PAYLOAD" description:"Maximum payload" default:"8192"`
}

func (b Bare) Execute(args []string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	def := types.Manifest{
		Name:        filepath.Base(wd),
		Description: b.Description,
		Run:         []string{"/bin/echo", "[\"hello\", \"world\"]"},
		OutputHeaders: map[string]string{
			"Content-Type": "application/json",
		},
		TimeLimit:      types.JsonDuration(b.TimeLimit),
		MaximumPayload: b.MaxPayload,
	}

	err = def.SaveAs(internal_app.ManifestFile)
	if err != nil {
		return err
	}

	makefile := "# define actions here\n"
	if b.Git {
		makefile += "update:\n\tgit pull origin master\n"
	}

	err = ioutil.WriteFile("Makefile", []byte(makefile), 0755)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(internal_app.CGIIgnore, []byte(""), 0755)
	if err != nil {
		return err
	}

	if b.Git {
		gctx, closer := internal.SignalContext()
		defer closer()
		return exec.CommandContext(gctx, "git", "init").Run()
	}
	return nil
}
