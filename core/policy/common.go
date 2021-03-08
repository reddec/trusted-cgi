package policy

import (
	"net"
	"net/http"
	"strings"
)

// Allow requests from with predefined values in Authorization header (tokens).
func ByToken(tokens ...string) Policy {
	rule := &byToken{tokens: map[string]bool{}}
	for _, token := range tokens {
		rule.tokens[token] = true
	}
	return rule
}

type byToken struct {
	tokens map[string]bool
}

func (bt *byToken) String() string {
	return "by token"
}

func (bt *byToken) IsAllowed(req *http.Request) bool {
	const header = "Authorization"
	return bt.tokens[req.Header.Get(header)]
}

// Allow requests only from predefined origins (Origin header). Case-insensitive.
func ByOrigins(origins ...string) Policy {
	rule := &byOrigin{origins: map[string]bool{}}
	for _, origin := range origins {
		rule.origins[strings.ToUpper(strings.TrimSpace(origin))] = true
	}
	return rule
}

type byOrigin struct {
	origins map[string]bool
}

func (bo *byOrigin) String() string {
	return "by origin"
}

func (bo *byOrigin) IsAllowed(req *http.Request) bool {
	return bo.origins[strings.ToUpper(req.Header.Get("Origin"))]
}

// Allow request only from predefined networks/IPs. If no mask defined in CIDR it means single address (/32 or /128).
// Invalid CIDRs and addresses will be silently ignored.
func ByCIDR(networksOrIP ...string) Policy {
	var nets []*net.IPNet
	for _, def := range networksOrIP {
		if !strings.Contains(def, "/") {
			if strings.Contains(def, ":") {
				def += "/128"
			} else {
				def += "/32"
			}
		}

		_, network, err := net.ParseCIDR(def)
		if err != nil {
			continue
		}
		nets = append(nets, network)
	}
	return &byIP{allowedCIDR: nets}
}

type byIP struct {
	allowedCIDR []*net.IPNet
}

func (bip *byIP) String() string {
	return "by IP"
}

func (bip *byIP) IsAllowed(req *http.Request) bool {
	var headers = []string{
		"X-Forwarded-For",
		"X-Host",
	}

	var remoteIP net.IP
	for _, header := range headers {
		value := strings.Split(req.Header.Get(header), ",")[0] // for multiple x-forwarded-for
		ip := net.ParseIP(value)
		if ip == nil {
			continue
		}
		remoteIP = ip
		break
	}

	if remoteIP == nil {
		host, _, _ := net.SplitHostPort(req.RemoteAddr)
		remoteIP = net.ParseIP(host)
	}

	if remoteIP == nil {
		// something very strange, just let it go
		return true
	}

	for _, network := range bip.allowedCIDR {
		if network.Contains(remoteIP) {
			return true
		}
	}
	return false
}
