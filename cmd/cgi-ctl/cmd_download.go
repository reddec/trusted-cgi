package main

import (
	"context"
	"fmt"
	"github.com/reddec/trusted-cgi/api"
	"github.com/reddec/trusted-cgi/api/client"
	"github.com/reddec/trusted-cgi/cmd/internal"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"log"
	"os"
	"strings"
	"syscall"
)

type remoteLink struct {
	Login    string `short:"l" long:"login" env:"LOGIN" description:"Login name" default:"admin"`
	Password string `short:"p" long:"password" env:"PASSWORD" description:"Password" default:"admin"`
	AskPass  bool   `short:"P" long:"ask-pass" env:"ASK_PASS" description:"Get password from stdin"`
	URL      string `short:"u" long:"url" env:"URL" description:"Trusted-CGI endpoint" default:"http://127.0.0.1:3434/"`
}

func (rl *remoteLink) Users() *client.UserAPIClient {
	return &client.UserAPIClient{BaseURL: rl.URL + "u/"}
}

func (rl *remoteLink) Lambdas() *client.LambdaAPIClient {
	return &client.LambdaAPIClient{BaseURL: rl.URL + "u/"}
}

func (rl *remoteLink) Token(ctx context.Context) (*api.Token, error) {
	if rl.AskPass {
		_, _ = fmt.Fprintf(os.Stderr, "Enter Password: ")
		bytePassword, err := terminal.ReadPassword(syscall.Stdin)
		if err != nil {
			return nil, err
		}
		rl.Password = strings.TrimSpace(string(bytePassword))
	}
	return rl.Users().Login(ctx, rl.Login, rl.Password)
}

type download struct {
	remoteLink
	UID    string `short:"i" long:"uid" env:"UID" description:"Lambda UID" required:"yes"`
	Output string `short:"o" long:"output" env:"OUTPUT" description:"Output data (- means stdout, empty means as UID)" default:""`
}

func (cmd *download) Execute(args []string) error {
	ctx, closer := internal.SignalContext()
	defer closer()
	log.SetOutput(os.Stderr)
	log.Println("login...")
	token, err := cmd.Token(ctx)
	if err != nil {
		return fmt.Errorf("login: %w", err)
	}
	log.Println("download...")
	tarball, err := cmd.Lambdas().Download(ctx, token, cmd.UID)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}

	var out io.Writer
	if cmd.Output == "" {
		cmd.Output = cmd.UID + ".tar.gz"
	}
	if cmd.Output == "-" {
		out = os.Stdout
	} else {
		log.Println("saving to", cmd.Output, "...")
		f, err := os.Create(cmd.Output)
		if err != nil {
			return fmt.Errorf("create destination file: %w", err)
		}
		defer f.Close()
		out = f
	}
	var w int
	for w < len(tarball) {
		n, err := out.Write(tarball[w:])
		if err != nil {
			return err
		}
		if n <= 0 {
			break
		}
		w += n
	}
	log.Println("done")
	return nil
}
