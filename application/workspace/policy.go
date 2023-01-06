package workspace

import (
	"context"
	"net/http"
	"strings"
	"sync"

	"github.com/reddec/trusted-cgi/application/config"
	"golang.org/x/crypto/bcrypt"
)

func NewPolicy(cfg *config.Policy) *Policy {
	return &Policy{
		cfg: cfg,
	}
}

type Policy struct {
	cfg    *config.Policy
	tokens sync.Map
}

func Protect(handler http.Handler, policies ...*Policy) http.Handler {
	if len(policies) == 0 {
		return handler
	}
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		for _, p := range policies {
			if p.isAllowed(request) {
				handler.ServeHTTP(writer, request.WithContext(WithPolicy(request.Context(), p.cfg)))
				return
			}
		}
		writer.WriteHeader(http.StatusUnauthorized)
	})
}

func (p *Policy) isAllowed(request *http.Request) bool {
	if !p.isAddressAllowed(request) {
		return false
	}
	if !p.isOriginAllowed(request) {
		return false
	}
	if !p.isTokenAllowed(request) {
		return false
	}
	return true
}

func (p *Policy) isAddressAllowed(request *http.Request) bool {
	if len(p.cfg.IPs) == 0 {
		return true
	}
	return p.cfg.IPs[remoteAddress(request)]
}

func (p *Policy) isOriginAllowed(request *http.Request) bool {
	if len(p.cfg.Origins) == 0 {
		return true
	}
	return p.cfg.Origins[request.Header.Get("Origin")]
}

func (p *Policy) isTokenAllowed(request *http.Request) bool {
	if len(p.cfg.Tokens) == 0 {
		return true
	}
	token := request.Header.Get("Authorization")
	if _, ok := p.tokens.Load(token); ok {
		return true
	}
	for _, t := range p.cfg.Tokens {
		if bcrypt.CompareHashAndPassword([]byte(t.Hash), []byte(token)) == nil {
			p.tokens.Store(token, true)
			return true
		}
	}
	return false
}

func remoteAddress(r *http.Request) string {
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

type policyCtx struct{}

func WithPolicy(ctx context.Context, p *config.Policy) context.Context {
	return context.WithValue(ctx, policyCtx{}, p)
}

func PolicyFromCtx(ctx context.Context) *config.Policy {
	v := ctx.Value(policyCtx{})
	if v == nil {
		return nil
	}
	return v.(*config.Policy)
}
