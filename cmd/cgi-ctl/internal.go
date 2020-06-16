package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/reddec/trusted-cgi/api"
	"github.com/reddec/trusted-cgi/api/client"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

const (
	configSection   = "trusted-cgi-ctl"
	controlFilename = ".cgictl.json"
)

type remoteLink struct {
	Login       string `short:"l" long:"login" env:"LOGIN" description:"Login name" default:"admin"`
	Password    string `short:"p" long:"password" env:"PASSWORD" description:"Password" default:"admin"`
	AskPass     bool   `short:"P" long:"ask-pass" env:"ASK_PASS" description:"Get password from stdin"`
	URL         string `short:"u" long:"url" env:"URL" description:"Trusted-CGI endpoint" default:"http://127.0.0.1:3434/"`
	Ghost       bool   `long:"ghost" env:"GHOST" description:"Disable save credentials to user config dir"`
	Independent bool   `long:"independent" env:"INDEPENDENT" description:"Disable read credentials from user config dir"`
}

func (rl *remoteLink) Users() *client.UserAPIClient {
	return &client.UserAPIClient{BaseURL: rl.URL + "u/"}
}

func (rl *remoteLink) Lambdas() *client.LambdaAPIClient {
	return &client.LambdaAPIClient{BaseURL: rl.URL + "u/"}
}

func (rl *remoteLink) Token(ctx context.Context) (*api.Token, error) {
	if !rl.Independent {
		var cf controlFile
		// check local control file for URL
		if err := cf.Read(controlFilename); err == nil {
			rl.URL = cf.URL
		}

		cfg, err := rl.readConfig()
		if err != nil && !os.IsNotExist(err) {
			log.Println("failed read config:", err)
			return nil, err
		} else if err == nil {
			rl.Password = string(cfg.Password)
			rl.Login = cfg.Login
			rl.AskPass = false
		}
	}
	if rl.AskPass {
		_, _ = fmt.Fprintf(os.Stderr, "Enter Password: ")
		bytePassword, err := terminal.ReadPassword(syscall.Stdin)
		if err != nil {
			return nil, err
		}
		rl.Password = strings.TrimSpace(string(bytePassword))
		_, _ = fmt.Fprintln(os.Stderr)
	}
	token, err := rl.Users().Login(ctx, rl.Login, rl.Password)
	if err != nil {
		return nil, err
	}
	if !rl.Ghost {
		err = rl.writeConfig()
	}
	return token, err
}

func (rl *remoteLink) readConfig() (*domainConfig, error) {
	cfg, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	info, err := url.Parse(rl.URL)
	if err != nil {
		return nil, err
	}
	filename := strings.ReplaceAll(info.Host, ":", "_")
	var dc domainConfig
	return &dc, dc.Read(filepath.Join(cfg, configSection, filename))
}

func (rl *remoteLink) writeConfig() error {
	cfg, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	info, err := url.Parse(rl.URL)
	if err != nil {
		return err
	}
	filename := strings.ReplaceAll(info.Host, ":", "_")
	tp := filepath.Join(cfg, configSection, filename)
	dc := &domainConfig{
		Login:    rl.Login,
		Password: []byte(rl.Password),
	}
	err = os.MkdirAll(filepath.Dir(tp), 0755)
	if err != nil {
		return err
	}
	return dc.Save(tp)
}

type domainConfig struct {
	Login    string `json:"login"`
	Password []byte `json:"password"`
}

func (dc *domainConfig) Save(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(dc)
}

func (dc *domainConfig) Read(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(dc)
}

type controlFile struct {
	URL string `json:"url"`
}

func (dc *controlFile) Save(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(dc)
}

func (dc *controlFile) Read(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(dc)
}

func appendIfNoLine(writer io.ReadWriter, line string) error {
	scanner := bufio.NewScanner(writer)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == line {
			return nil
		}
	}
	_, err := writer.Write([]byte(line + "\n"))
	return err
}

func appendIfNoLineFile(filename string, line string) error {
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer f.Close()
	return appendIfNoLine(f, line)
}
