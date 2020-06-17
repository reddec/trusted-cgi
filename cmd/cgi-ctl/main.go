package main

import (
	"github.com/jessevdk/go-flags"
	"log"
	"os"
)

const version = "dev"

type Config struct {
	Init struct {
		Bare Bare `command:"bare" description:"create bare template"`
	} `command:"init" description:"initialize function in a current directory"`
	Download download `command:"download" description:"download lambda content to the local tarball or stdout"`
	Upload   upload   `command:"upload" description:"upload content to lambda to the remote platform"`
	Clone    clone    `command:"clone" description:"clone lambda to local FS and keep URL for future tracking"`
	Do       do       `command:"do" description:"invoke actions (without actions it will print all available actions)"`
	Create   create   `command:"create" description:"create new lambda on the remote platform and initialize local environment"`
	Alias    alias    `command:"alias" description:"list, created or remove alias for the lambda"`
	Update   struct {
		Manifest updateManifest `command:"manifest" description:"pull and save remote manifest file"`
	} `command:"update" description:"update parts of the lambda"`
}

func main() {
	var config Config
	log.SetOutput(os.Stderr)
	parser := flags.NewParser(&config, flags.Default)
	parser.LongDescription = "Easy CGI-like server for development (helper tool)\nAuthor: Baryshnikov Aleksandr <dev@baryshnikov.net>\nVersion: " + version
	_, err := parser.Parse()
	if err != nil {
		os.Exit(1)
	}
}
