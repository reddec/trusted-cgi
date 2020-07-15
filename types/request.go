package types

import (
	"io"
	"net/http"
)

//go:generate msgp
type Request struct {
	Method        string            `json:"method" msg:"method"`
	URL           string            `json:"url" msg:"url"`
	Path          string            `json:"path" msg:"path"`
	RemoteAddress string            `json:"remote_address" msg:"remote_address"`
	Form          map[string]string `json:"form" msg:"form"`
	Headers       map[string]string `json:"headers" msg:"headers"`
	Body          io.ReadCloser     `json:"-" msg:"-"`
}

// Create request from HTTP request
func FromHTTP(r *http.Request) *Request {
	_ = r.ParseForm()
	var vals = make(map[string]string)
	for k, v := range r.Form {
		vals[k] = v[0]
	}
	var headers = make(map[string]string)
	for k, v := range r.Header {
		headers[k] = v[0]
	}
	return &Request{
		Method:        r.Method,
		URL:           r.RequestURI,
		Path:          r.URL.Path,
		RemoteAddress: r.RemoteAddr,
		Form:          vals,
		Headers:       headers,
		Body:          r.Body,
	}
}

// Returns shallow copy of request with new body
func (z *Request) WithBody(reader io.ReadCloser) *Request {
	if z == nil {
		return nil
	}
	cp := *z
	cp.Body = reader
	return &cp
}
