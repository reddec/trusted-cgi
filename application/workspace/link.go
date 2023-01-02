package workspace

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Masterminds/sprig"
	"github.com/reddec/trusted-cgi/application/config"
	"io"
	"strings"
	"text/template"
)

func NewSync(cfg config.Call, lambda *Lambda) (*Sync, error) {
	link, err := newLink(cfg.Invoke)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}
	return &Sync{
		lambda: lambda,
		link:   link,
	}, nil
}

type Sync struct {
	lambda *Lambda
	link   *Link
}

func (s *Sync) Call(ctx context.Context, payload io.Reader, dataContext any) (io.ReadCloser, error) {
	envs, payload, err := s.link.prepare(payload, dataContext)
	if err != nil {
		return nil, fmt.Errorf("prepare: %w", err)
	}

	return s.lambda.Invoke(ctx, envs, payload)
}

func NewAsync(cfg config.Enqueue, queue *Queue) (*Async, error) {
	link, err := newLink(cfg.Invoke)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}
	return &Async{
		queue: queue,
		link:  link,
	}, nil
}

type Async struct {
	queue *Queue
	link  *Link
}

func (as *Async) Push(ctx context.Context, payload io.Reader, dataContext any) error {
	envs, payload, err := as.link.prepare(payload, dataContext)
	if err != nil {
		return fmt.Errorf("prepare: %w", err)
	}

	return as.queue.Push(ctx, envs, payload)
}

func newLink(cfg config.Invoke) (*Link, error) {
	envs, err := parseEnvTemplate(cfg.Environment)
	if err != nil {
		return nil, fmt.Errorf("parse environment: %w", err)
	}
	var payload *template.Template
	if cfg.Payload != nil {
		t, err := parseTemplate(*cfg.Payload)
		if err != nil {
			return nil, fmt.Errorf("parse payload: %w", err)
		}
		payload = t
	}
	return &Link{
		envs:    envs,
		payload: payload,
	}, nil
}

type Link struct {
	envs    map[string]*template.Template
	payload *template.Template
}

func (link *Link) prepare(payload io.Reader, dataContext any) (map[string]string, io.Reader, error) {
	var envs = make(map[string]string, len(link.envs))
	for k, t := range link.envs {
		v, err := renderTemplate(t, dataContext)
		if err != nil {
			return nil, nil, fmt.Errorf("render environment %s: %w", k, err)
		}
		envs[k] = v
	}
	if link.payload != nil {
		data, err := renderTemplate(link.payload, dataContext)
		if err != nil {
			return nil, nil, fmt.Errorf("render body: %w", err)
		}
		payload = strings.NewReader(data)
	}
	return envs, payload, nil
}

func parseEnvTemplate(envTemplates map[string]string) (map[string]*template.Template, error) {
	var envs = make(map[string]*template.Template, len(envTemplates))
	for k, v := range envTemplates {
		t, err := parseTemplate(v)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", k, err)
		}
		envs[k] = t
	}
	return envs, nil
}

func parseTemplate(text string) (*template.Template, error) {
	return template.New("").Funcs(sprig.TxtFuncMap()).Parse(text)
}

func renderTemplate(t *template.Template, dataContext any) (string, error) {
	var buf bytes.Buffer
	err := t.Execute(&buf, dataContext)
	return buf.String(), err
}

func cloneEnvs(envs map[string]string) map[string]string {
	cp := make(map[string]string, len(envs))
	for k, v := range envs {
		cp[k] = v
	}
	return cp
}
