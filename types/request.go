package types

import (
	"io"
	"net/http"
	"strings"
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
func FromHTTP(r *http.Request, behindProxy bool) *Request {
	_ = r.ParseForm()
	var vals = make(map[string]string)
	for k, v := range r.Form {
		vals[k] = v[0]
	}
	var headers = make(map[string]string)
	for k, v := range r.Header {
		headers[k] = v[0]
	}
	var address string
	if behindProxy {
		address = getRequestAddress(r)
	} else {
		address = r.RemoteAddr
	}
	return &Request{
		Method:        r.Method,
		URL:           r.RequestURI,
		Path:          r.URL.Path,
		RemoteAddress: address,
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

func getRequestAddress(r *http.Request) string {
	address := getFirstHeader(r, r.RemoteAddr, "X-Real-Ip", "X-Forwarded-For")
	address = strings.TrimSpace(strings.SplitN(address, ",", 2)[0])
	return address
}

func getFirstHeader(r *http.Request, defaultValue string, headers ...string) string {
	for _, h := range headers {
		if v := r.Header.Get(h); v != "" {
			return v
		}
	}
	return defaultValue
}
