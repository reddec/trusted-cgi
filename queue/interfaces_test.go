package queue_test

import (
	"bytes"
	"context"
	"github.com/google/uuid"
	"github.com/reddec/trusted-cgi/queue"
	"github.com/reddec/trusted-cgi/queue/indir"
	"github.com/reddec/trusted-cgi/queue/inmemory"
	"github.com/reddec/trusted-cgi/types"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func testPutPeek(ctx context.Context, t *testing.T, queue queue.Queue) *types.Request {
	payload := uuid.New().String()

	req := &types.Request{
		Method:        "POST",
		URL:           "http://example.com:8889/sample/" + payload,
		Path:          "/sample/" + payload,
		RemoteAddress: "127.0.0.2:9992",
		Form: map[string]string{
			"USER": "user1",
		},
		Headers: map[string]string{
			"Content-Type": "text/plain",
		},
		Body: ioutil.NopCloser(bytes.NewBufferString(payload)),
	}
	err := queue.Put(ctx, req)
	if !assert.NoError(t, err) {
		return req
	}

	v, err := queue.Peek(ctx)
	if !assert.NoError(t, err) {
		return req
	}
	data, err := ioutil.ReadAll(v.Body)
	if !assert.NoError(t, err) {
		return req
	}
	assert.Equal(t, string(data), payload)
	assert.Equal(t, v.WithBody(nil), req.WithBody(nil))
	return req
}

func TestInMemory(t *testing.T) {
	ctx := context.Background()
	q := inmemory.New(1)
	defer q.Close()

	req := testPutPeek(ctx, t, q)
	// try again - should be same result
	v2, err := q.Peek(ctx)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, v2.WithBody(nil), req.WithBody(nil))

	err = q.Commit(ctx)
	if !assert.NoError(t, err) {
		return
	}
	// put again
	testPutPeek(ctx, t, q)

	q.Close()

	_, err = q.Peek(ctx)
	assert.Error(t, err)

	var closed bool
	select {
	case <-q.Done():
		closed = true
	default:

	}
	assert.True(t, closed, "should be closed")

}

func TestInDir(t *testing.T) {
	ctx := context.Background()
	q, err := indir.New("test/queue")
	if !assert.NoError(t, err) {
		return
	}

	req := testPutPeek(ctx, t, q)
	// try again - should be same result
	v2, err := q.Peek(ctx)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, v2.WithBody(nil), req.WithBody(nil))

	err = q.Commit(ctx)
	if !assert.NoError(t, err) {
		return
	}
	// put again
	testPutPeek(ctx, t, q)
}
