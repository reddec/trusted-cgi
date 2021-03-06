package lambdas

import (
	"errors"
	"net/http"
)

type Lambda interface {
	http.Handler
}

var ErrNotFound = errors.New("lambda not found")

type Storage interface {
	Find(name string) (Lambda, error)
}
