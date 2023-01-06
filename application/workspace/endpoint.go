package workspace

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"text/template"

	"github.com/go-chi/chi/v5"
	"github.com/reddec/trusted-cgi/application/config"
	"github.com/reddec/trusted-cgi/trace"
)

type Function interface {
	Call(ctx context.Context, renderCtx any, payload io.Reader) (io.ReadCloser, error)
}

func NewHandler(project *Project, cfg *config.HTTP, fn Function) (http.Handler, error) {
	headers, err := parseEnvTemplate(cfg.Headers)
	if err != nil {
		return nil, fmt.Errorf("parse header: %w", err)
	}
	vars, err := parseEnvTemplate(cfg.Vars)
	if err != nil {
		return nil, fmt.Errorf("parse vars: %w", err)
	}
	return sizeLimit(cfg.Body, &Handler{
		project: project,
		config:  cfg,
		fn:      fn,
		vars:    vars,
		headers: headers,
	}), nil
}

type Handler struct {
	project *Project
	config  *config.HTTP
	vars    map[string]*template.Template
	headers map[string]*template.Template
	fn      Function
}

func (ep *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	policy := PolicyFromCtx(request.Context())
	tracer := ep.project.Trace()
	tracer.Set("path", ep.config.Path)
	tracer.Set("method", ep.config.Method)
	// TODO: tracer.Set("address",)
	tracer.Set("headers", request.Header)
	tracer.Set("request_uri", request.RequestURI)
	if policy != nil {
		tracer.Set("policy", policy.Name)
	}

	defer request.Body.Close()

	ec := &endpointContext{
		req:  request,
		vars: map[string]string{},
	}

	// render vars
	for k, t := range ep.vars {
		v, err := renderTemplate(t, ec)
		if err != nil {
			log.Println("failed render var:", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		ec.vars[k] = v
	}

	// render env
	for k, t := range ep.vars {
		v, err := renderTemplate(t, ec)
		if err != nil {
			log.Println("failed render var:", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		ec.vars[k] = v
	}

	ctx := trace.WithTrace(request.Context(), tracer)
	out, err := ep.fn.Call(ctx, ec, request.Body)
	if err != nil {
		log.Println("failed call:", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer out.Close()

	// render response headers
	for k, t := range ep.headers {
		v, err := renderTemplate(t, ec)
		if err != nil {
			log.Println("failed render header:", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		writer.Header().Set(k, v)
	}

	writer.WriteHeader(ep.status())

	_, err = io.Copy(writer, out)
	if err != nil {
		log.Println("failed pipe response:", err)
		return
	}
}

func (ep *Handler) status() int {
	if s := ep.config.Status; s > 0 {
		return s
	}
	return http.StatusOK
}

// lazy-caching thread-unsafe context for data source.
type endpointContext struct {
	req   *http.Request
	form  url.Values
	query url.Values
	vars  map[string]string
	path  struct {
		parsed bool
		data   map[string]string
	}
	json struct {
		parsed bool
		value  interface{}
	}
	body struct {
		cached bool
		data   []byte
	}
}

func (rc *endpointContext) Var() map[string]string {
	return rc.vars
}

func (rc *endpointContext) Path() map[string]string {
	if rc.path.parsed {
		return rc.path.data
	}
	ctx := chi.RouteContext(rc.req.Context())
	var params = make(map[string]string, len(ctx.URLParams.Keys))
	for _, k := range ctx.URLParams.Keys {
		params[k] = ctx.URLParam(k)
	}
	rc.path.data = params
	rc.path.parsed = true
	return params
}

func (rc *endpointContext) Header() http.Header {
	return rc.req.Header
}

func (rc *endpointContext) JSON() (interface{}, error) {
	if rc.json.parsed {
		return rc.json.value, nil
	}

	var value interface{}
	if err := json.NewDecoder(rc.req.Body).Decode(&value); err != nil {
		return nil, err
	}
	rc.json.parsed = true
	rc.json.value = value
	return rc.json, nil
}

func (rc *endpointContext) Query() url.Values {
	if rc.query != nil {
		return rc.query
	}
	rc.query, _ = url.ParseQuery(rc.req.URL.RawQuery)
	return rc.query
}

func (rc *endpointContext) Body() ([]byte, error) {
	if rc.body.cached {
		return rc.body.data, nil
	}
	data, err := io.ReadAll(rc.req.Body)
	if err != nil {
		return nil, fmt.Errorf("read Body: %w", err)
	}
	rc.body.cached = true
	rc.body.data = data
	return data, nil
}

func (rc *endpointContext) Form() (url.Values, error) {
	if rc.form != nil {
		return rc.form, nil
	}
	if rc.req.Method == http.MethodGet {
		return rc.Query(), nil
	}
	if err := rc.req.ParseForm(); err != nil {
		return nil, err
	}
	rc.form = rc.req.PostForm
	return rc.form, nil
}

func sizeLimit(maxSize int64, handler http.Handler) http.Handler {
	if maxSize <= 0 {
		return handler
	}
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.ContentLength > 0 && request.ContentLength > maxSize {
			writer.WriteHeader(http.StatusRequestEntityTooLarge)
			return
		}
		request.Body = &limitedReader{
			wrap: request.Body,
			left: maxSize,
		}
		handler.ServeHTTP(writer, request)
	})
}

var ErrTooLarge = fmt.Errorf("entity is too large")

type limitedReader struct {
	wrap io.ReadCloser
	left int64
}

func (l *limitedReader) Read(p []byte) (n int, err error) {
	if l.left < 0 {
		err = ErrTooLarge
		return
	}
	if int64(len(p)) > l.left {
		p = p[:l.left+1]
	}
	n, err = l.wrap.Read(p)
	l.left -= int64(n)
	if err != nil {
		return
	}
	if l.left < 0 {
		err = ErrTooLarge
	}
	return
}

func (l *limitedReader) Close() error {
	return l.wrap.Close()
}
