package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/alecthomas/units"
	"github.com/reddec/trusted-cgi/cmd/internal"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type invoke struct {
	uidLocator
	Input       string            `short:"i" long:"input" env:"INPUT" description:"input file that will be used as body (- or empty is stdin)" default:"-"`
	Output      string            `short:"o" long:"output" env:"OUTPUT" description:"output file for response (- or empty is stdout)" default:"-"`
	Get         bool              `short:"g" long:"get" env:"GET" description:"use GET method instead of POST (body will be ignored)"`
	Token       string            `short:"t" long:"token" env:"TOKEN" description:"add authorization token"`
	Origin      string            `short:"O" long:"origin" env:"ORIGIN" description:"add origin header"`
	ContentType string            `short:"C" long:"content-type" env:"CONTENT_TYPE" description:"set content-type header" default:"application/json"`
	Header      map[string]string `short:"H" long:"header" env:"HEADER" description:"custom headers"`
	Field       map[string]string `short:"f" long:"field" env:"FIELD" description:"set JSON field (input will be ignored)"`
	Verbose     bool              `short:"v" long:"verbose" env:"VERBOSE" description:"show logs"`
}

func (cmd *invoke) Execute(args []string) error {
	if cmd.Verbose {
		log.SetOutput(os.Stderr)
	} else {
		log.SetOutput(ioutil.Discard)
	}
	if err := cmd.parseUID(); err != nil {
		return err
	}
	var cf controlFile
	if err := cf.Read(controlFilename); err != nil {
		return fmt.Errorf("read control file: %w", err)
	}
	ctx, closer := internal.SignalContext()
	defer closer()

	body, err := cmd.getBody(ctx)
	if err != nil {
		return fmt.Errorf("get body: %w", err)
	}
	log.Println("request body size:", units.Base2Bytes(len(body)))
	url := urlJoin(cf.URL, "a", cmd.UID)
	req, err := http.NewRequestWithContext(ctx, cmd.getMethod(), url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("prepare request: %w", err)
	}
	cmd.setHeaders(req)
	log.Println(req.Method, "request to", url)
	for k, v := range req.Header {
		log.Println(k, "=", v)
	}
	log.Println(string(body))
	log.Println("invoking...")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer res.Body.Close()
	log.Println("status code:", res.StatusCode)
	for k, v := range res.Header {
		log.Println(k, "=", v)
	}
	exitCode := 0
	if res.StatusCode/100 != 2 {
		exitCode = res.StatusCode / 100
	}

	_, _ = io.Copy(os.Stdout, res.Body)

	os.Exit(exitCode)
	return nil
}

func (cmd *invoke) getBody(ctx context.Context) ([]byte, error) {
	if len(cmd.Field) > 0 {
		return json.MarshalIndent(cmd.Field, "", "  ")
	}
	if cmd.Input == "" || cmd.Input == "-" {
		return ioutil.ReadAll(os.Stdin)
	}
	return ioutil.ReadFile(cmd.Input)
}

func (cmd *invoke) getMethod() string {
	if cmd.Get {
		return http.MethodGet
	}
	return http.MethodPost
}

func (cmd *invoke) setHeaders(req *http.Request) {
	for k, v := range cmd.Header {
		req.Header.Set(k, v)
	}
	req.Header.Set("Content-Type", cmd.ContentType)
	if cmd.Origin != "" {
		req.Header.Set("Origin", cmd.Origin)
	}
	if cmd.Token != "" {
		req.Header.Set("Authorization", cmd.Token)
	}
}
