package workspace

import (
	"github.com/reddec/trusted-cgi/application/config"
	"net/http"
)

func NewEndpoint(cfg config.Endpoint, calls []*Lambda, queues []*Queue) http.Handler {
	return nil //TODO
}
