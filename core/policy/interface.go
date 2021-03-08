package policy

import (
	"errors"
	"net/http"
)

var ErrNotFound = errors.New("policy not found")

type Storage interface {
	FindByLambda(lambdaID string) ([]Policy, error)
}

type Policy interface {
	IsAllowed(req *http.Request) bool
}
