package lambdas

import (
	"errors"
	"net/http"
)

var ErrNotFound = errors.New("lambda not found")

type Storage interface {
	Find(name string) (http.Handler, error)
}
