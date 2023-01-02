package workspace

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/ohler55/ojg/jp"
	"github.com/reddec/trusted-cgi/application/config"
	"io"
	"log"
	"net/http"
	"net/url"
	"text/template"
)

type CacheStorage interface {
	Write(data io.Reader) (string, error)
	Open(string) (io.ReadCloser, error)
	Remove(string) error
}

func NewEndpoint(cfg config.Endpoint, cache CacheStorage, sync []*Sync, async []*Async) (http.Handler, error) {
	headers, err := parseEnvTemplate(cfg.Headers)
	if err != nil {
		return nil, fmt.Errorf("parse header: %w", err)
	}
	if cfg.Status <= 0 {
		if len(async) > 0 && len(sync) == 0 {
			cfg.Status = http.StatusCreated
		} else if len(async) == 0 && len(sync) == 0 {
			cfg.Status = http.StatusNoContent
		} else {
			cfg.Status = http.StatusOK
		}
	}
	return &Endpoint{
		status:  cfg.Status,
		sync:    sync,
		async:   async,
		headers: headers,
		cache:   cache,
	}, nil
}

type Endpoint struct {
	status  int
	sync    []*Sync
	async   []*Async
	headers map[string]*template.Template
	cache   CacheStorage
}

func (ep *Endpoint) ServeHTTP(original http.ResponseWriter, request *http.Request) {
	writer := &writeStatusOnce{wrapped: original}

	defer request.Body.Close()
	cacheID, err := ep.cache.Write(request.Body)
	if err != nil {
		log.Println("failed write request data to cache:", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer ep.cache.Remove(cacheID)
	ec := &endpointContext{
		req:     request,
		cacheID: cacheID,
		cache:   ep.cache,
	}

	ctx := request.Context()

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

	// push to queue
	for _, async := range ep.async {
		data, err := ep.cache.Open(cacheID)
		if err != nil {
			log.Println("failed read cache:", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = async.Push(ctx, data, ec)
		_ = data.Close()
		if err != nil {
			log.Println("failed push to queue:", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	if len(ep.sync) == 0 {
		writer.WriteHeader(ep.status)
	}

	// call and pipe results
	for _, sync := range ep.sync {
		if err := ep.callSync(ctx, sync, ec, writer); err != nil {
			log.Println("failed read cache:", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func (ep *Endpoint) callSync(ctx context.Context, sync *Sync, ec *endpointContext, writer http.ResponseWriter) error {
	data, err := ep.cache.Open(ec.cacheID)
	if err != nil {
		return fmt.Errorf("open cache: %w", err)
	}
	defer data.Close()

	out, err := sync.Call(ctx, data, ec)
	if err != nil {
		return fmt.Errorf("call lambda: %w", err)
	}

	writer.WriteHeader(ep.status)

	// pipe result
	if _, err := io.Copy(writer, out); err != nil {
		return fmt.Errorf("pipe lambda result: %w", err)
	}
	return nil
}

// lazy-caching thread-unsafe context for data source.
type endpointContext struct {
	cache   CacheStorage
	cacheID string
	req     *http.Request
	form    url.Values
	query   url.Values
	path    struct {
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

	if err := rc.resetBody(); err != nil {
		return nil, fmt.Errorf("reset body: %w", err)
	}

	var value interface{}
	if err := json.NewDecoder(rc.req.Body).Decode(&value); err != nil {
		return nil, err
	}
	rc.json.parsed = true
	rc.json.value = value
	return rc.json, nil
}

func (rc *endpointContext) JSONPath(path string) (interface{}, error) {
	expr, err := jp.ParseString(path)
	if err != nil {
		return nil, err
	}
	obj, err := rc.JSON()
	if err != nil {
		return nil, err
	}
	items := expr.Get(obj)
	if len(items) == 0 {
		return nil, nil
	}
	return items[0], nil
}

func (rc *endpointContext) Query() url.Values {
	if rc.query != nil {
		return rc.query
	}
	rc.query = rc.req.URL.Query()
	return rc.query
}

func (rc *endpointContext) Body() ([]byte, error) {
	if rc.body.cached {
		return rc.body.data, nil
	}
	if err := rc.resetBody(); err != nil {
		return nil, fmt.Errorf("reset body: %w", err)
	}
	data, err := io.ReadAll(rc.req.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
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
	if err := rc.resetBody(); err != nil {
		return nil, fmt.Errorf("reset body: %w", err)
	}
	if err := rc.req.ParseForm(); err != nil {
		return nil, err
	}
	rc.form = rc.req.PostForm
	return rc.form, nil
}

func (rc *endpointContext) resetBody() error {
	_ = rc.req.Body.Close()
	in, err := rc.cache.Open(rc.cacheID)
	if err != nil {
		return fmt.Errorf("get from cache: %w", err)
	}
	rc.req.Body = in
	return nil
}

type writeStatusOnce struct {
	status  bool
	wrapped http.ResponseWriter
}

func (ws *writeStatusOnce) Header() http.Header {
	return ws.wrapped.Header()
}

func (ws *writeStatusOnce) Write(bytes []byte) (int, error) {
	ws.status = true
	return ws.wrapped.Write(bytes)
}

func (ws *writeStatusOnce) WriteHeader(statusCode int) {
	if ws.status {
		return
	}
	ws.status = true
	ws.wrapped.WriteHeader(statusCode)
}
