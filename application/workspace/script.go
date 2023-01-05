package workspace

import (
	"bytes"
	"context"
	"fmt"
	"github.com/reddec/trusted-cgi/application/config"
	"github.com/reddec/trusted-cgi/internal"
	"github.com/reddec/trusted-cgi/trace"
	"github.com/reddec/trusted-cgi/types"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

func NewScript(project *Project, script *config.Script) (*Script, error) {
	envs, err := parseEnvTemplate(script.Environment)
	if err != nil {
		return nil, fmt.Errorf("parse env: %w", err)
	}
	var payload *template.Template
	if script.Payload != nil {
		payload, err = parseTemplate(*script.Payload)
		if err != nil {
			return nil, fmt.Errorf("parse pyload: %w", err)
		}
	}
	return &Script{
		project: project,
		config:  script,
		env:     envs,
		payload: payload,
	}, nil
}

type Script struct {
	project *Project
	config  *config.Script
	env     map[string]*template.Template
	payload *template.Template
}

func (s *Script) WorkDir() string {
	if s.project != nil {
		return filepath.Join(s.project.dir, filepath.Clean(s.config.WorkDir))
	}
	return s.config.WorkDir
}

func (s *Script) Credentials() *types.Credential {
	if s.project != nil {
		return s.project.Credentials()
	}
	return nil
}

func (s *Script) Render(renderCtx any, payload io.Reader) (map[string]string, io.Reader, error) {
	if s.payload != nil {
		data, err := renderTemplate(s.payload, renderCtx)
		if err != nil {
			return nil, nil, fmt.Errorf("render payload: %w", err)
		}
		payload = bytes.NewReader([]byte(data))
	}

	var kv = make(map[string]string, len(s.env))
	for k, t := range s.env {
		v, err := renderTemplate(t, renderCtx)
		if err != nil {
			return nil, nil, fmt.Errorf("render env '%s': %w", k, err)
		}
		kv[k] = v
	}
	return kv, payload, nil
}

// Call lambda and render environment and payload before execution.
func (s *Script) Call(ctx context.Context, renderCtx any, payload io.Reader) (io.ReadCloser, error) {
	env, payload, err := s.Render(renderCtx, payload)
	if err != nil {
		return nil, err
	}
	return s.Invoke(ctx, env, payload)
}

// Invoke lambda. Close MUST be called.
func (s *Script) Invoke(global context.Context, environment map[string]string, payload io.Reader) (io.ReadCloser, error) {
	tracer := trace.NewTraceFromContext(global)
	tracer.Set("command", s.config.Command)
	tracer.Set("args", s.config.Args)
	ctx, cancel := s.createContext(global)
	rd, wr := io.Pipe()

	inputSniffer := trace.NewSniffer(payload, s.project.workspace.SniffSize())
	outputSniffer := trace.NewSniffer(rd, s.project.workspace.SniffSize())

	cmd := exec.CommandContext(ctx, s.config.Command, s.config.Args...)
	cmd.Dir = s.WorkDir()
	cmd.Stdin = payload
	cmd.Stdout = wr
	cmd.Stderr = os.Stderr
	if creds := s.Credentials(); creds != nil {
		internal.SetCreds(cmd, creds)
	}
	internal.SetFlags(cmd)
	var env = os.Environ()
	for k, v := range environment {
		env = append(env, k+"="+v)
	}
	cmd.Env = env
	tracer.Set("environment", env)
	if err := cmd.Start(); err != nil {
		cancel()
		_ = wr.Close()
		_ = rd.Close()
		return nil, fmt.Errorf("start lambda: %w", err)
	}

	done := make(chan error, 1)
	go func() {
		defer tracer.Close()
		defer close(done)
		defer inputSniffer.Report(tracer, "input")
		defer outputSniffer.Report(tracer, "output")
		err := cmd.Wait()
		_ = wr.CloseWithError(err)
		done <- err
	}()

	return &cancelCloseReader{
		ReadCloser: rd,
		cancel:     cancel,
		done:       done,
	}, nil
}

func (s *Script) createContext(global context.Context) (context.Context, context.CancelFunc) {
	if timeout := s.config.Timeout; timeout > 0 {
		return context.WithTimeout(global, timeout.Duration())
	}
	return context.WithCancel(global)
}

type cancelCloseReader struct {
	io.ReadCloser
	done   chan error
	cancel func()
}

func (c *cancelCloseReader) Close() error {
	c.cancel()
	err := c.ReadCloser.Close()
	<-c.done
	return err
}
