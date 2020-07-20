package policy

import (
	"fmt"
	"github.com/reddec/trusted-cgi/application"
	"github.com/reddec/trusted-cgi/types"
	"net"
)

func checkPolicy(policy application.PolicyDefinition, req *types.Request) error {
	host, _, _ := net.SplitHostPort(req.RemoteAddress)
	if len(policy.AllowedIP) > 0 && !policy.AllowedIP.Has(host) {
		return fmt.Errorf("IP restricted")
	}
	if len(policy.AllowedOrigin) > 0 && !policy.AllowedOrigin.Has(req.Headers["Origin"]) {
		return fmt.Errorf("origin restricted")
	}

	if !policy.Public {
		_, ok := policy.Tokens[req.Headers["Authorization"]]
		if !ok {
			return fmt.Errorf("token restricted")
		}
	}
	return nil
}
