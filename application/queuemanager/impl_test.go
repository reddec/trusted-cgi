package queuemanager_test

import (
	"bytes"
	"context"
	"github.com/reddec/trusted-cgi/application"
	"github.com/reddec/trusted-cgi/application/queuemanager"
	"github.com/reddec/trusted-cgi/queue"
	"github.com/reddec/trusted-cgi/queue/inmemory"
	"github.com/reddec/trusted-cgi/types"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"testing"
)

type hf func(request types.Request, out io.Writer) error
type mockPlatform struct {
	handlers map[string]hf
}

func (mini *mockPlatform) InvokeByUID(ctx context.Context, uid string, request types.Request, out io.Writer) error {
	handler, ok := mini.handlers[uid]
	if !ok {
		return os.ErrNotExist
	}
	return handler(request, out)
}

type bypass struct {
}

func (b bypass) Inspect(lambda string, request *types.Request) error {
	return nil
}

func TestNew(t *testing.T) {
	var echoText string
	var echoCh = make(chan struct{})
	platform := &mockPlatform{
		handlers: map[string]hf{
			"echo": func(request types.Request, out io.Writer) error {
				defer request.Body.Close()
				defer close(echoCh)
				data, err := ioutil.ReadAll(request.Body)
				if err != nil {
					t.Error(err)
					return err
				}
				echoText = string(data)
				return nil
			},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	qm, err := queuemanager.New(ctx,
		queuemanager.Mock(application.Queue{
			Name:   "queue-1",
			Target: "echo",
		}, application.Queue{
			Name:   "queue-2",
			Target: "greeter",
		}), platform, func(name string) (queue.Queue, error) {
			return inmemory.New(10), nil
		}, &bypass{})
	if err != nil {
		t.Error(err)
		return
	}

	err = qm.Put("queue-1", mockRequest("hello world"))
	if err != nil {
		t.Error(err)
		return
	}
	<-echoCh

	if echoText != "hello world" {
		t.Error("corrupted message")
	}

	err = qm.Put("queue-not-exists", mockRequest("test"))
	if err == nil {
		t.Error("should fail")
		return
	}

	err = qm.Remove("queue-1")
	if err != nil {
		t.Error(err)
		return
	}

	list := qm.List()
	if len(list) != 1 {
		t.Error("should be list of 1")
		return
	}
	if list[0].Name != "queue-2" {
		t.Error("should be queue-2 but " + list[0].Name)
	}
	if list[0].Target != "greeter" {
		t.Error("should be echo")
	}

	err = qm.Assign("queue-2", "echo")
	if err != nil {
		t.Error(err)
	}

	err = qm.Add(application.Queue{
		Name:   "queue-3",
		Target: "echo",
	})
	if err != nil {
		t.Error(err)
	}

	list = qm.Find("echo")
	if len(list) != 2 {
		t.Error("should be list of 2")
		return
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Name < list[j].Name
	})
	if list[0].Name != "queue-2" {
		t.Error("should be queue-2 but " + list[0].Name)
	}
	if list[1].Name != "queue-3" {
		t.Error("should be queue-3 but " + list[1].Name)
	}

	cancel()
	qm.Wait()
}

func mockRequest(payload string) *types.Request {
	return &types.Request{
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
}
