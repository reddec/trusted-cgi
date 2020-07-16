package queue

import (
	"context"
	"github.com/reddec/trusted-cgi/types"
)

// Thread-safe FIFO queue designed for one multiple concurrent writers and single consumer.
// Queue should store somewhere request body.
type Queue interface {
	// Put request to queue
	Put(ctx context.Context, request *types.Request) error
	// Peek oldest request or wait till new data arrived/context expiration
	Peek(ctx context.Context) (*types.Request, error)
	// Commit (remove) oldest record
	Commit(ctx context.Context) error
	// Clean all internal allocated resource
	Destroy() error
}
