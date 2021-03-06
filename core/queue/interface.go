package queue

import "net/http"

type Queue interface {
	Enqueue(req *http.Request) (string, error)
}
