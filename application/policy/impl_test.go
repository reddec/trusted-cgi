package policy

import (
	"bytes"
	"github.com/reddec/trusted-cgi/application"
	"github.com/reddec/trusted-cgi/types"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func TestNew(t *testing.T) {
	policy, err := New(Mock(application.Policy{
		ID: "foo",
		Definition: application.PolicyDefinition{
			Public: false,
			Tokens: map[string]string{
				"DEADBEAF": "Consumer 1",
				"BEAFDEAD": "Consumer 2",
			},
		},
		Lambdas: map[string]bool{
			"lambda-1": true,
			"lambda-2": true,
		},
	}))
	if err != nil {
		t.Error(err)
		return
	}
	t.Run("no applied policy", func(t *testing.T) {
		err := policy.Inspect("lambda-3", mockRequest("hello"))
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("valid token", func(t *testing.T) {
		req := mockRequest("hello")
		req.Headers["Authorization"] = "DEADBEAF"
		err := policy.Inspect("lambda-1", req)
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("invalid token", func(t *testing.T) {
		req := mockRequest("hello")
		req.Headers["Authorization"] = "1111"
		err := policy.Inspect("lambda-1", req)
		if err == nil {
			t.Error("should fail")
		}
	})
	t.Run("list policies", func(t *testing.T) {
		list := policy.List()
		assert.Len(t, list, 1)
		assert.Equal(t, "foo", list[0].ID)
		assert.Equal(t, list[0].Lambdas, types.StringSet("lambda-1", "lambda-2"))
	})
	t.Run("clear", func(t *testing.T) {
		err := policy.Clear("lambda-2")
		assert.NoError(t, err)
		list := policy.List()
		assert.Len(t, list, 1)
		assert.Equal(t, "foo", list[0].ID)
		assert.Equal(t, list[0].Lambdas, types.StringSet("lambda-1"))
	})
	t.Run("apply", func(t *testing.T) {
		err := policy.Apply("lambda-4", "foo")
		assert.NoError(t, err)
		list := policy.List()
		assert.Len(t, list, 1)
		assert.Equal(t, "foo", list[0].ID)
		assert.Contains(t, list[0].Lambdas, "lambda-4")
	})
	t.Run("update", func(t *testing.T) {
		err := policy.Update("foo", application.PolicyDefinition{
			AllowedOrigin: types.StringSet("google"),
			Public:        true,
		})
		assert.NoError(t, err)
		req := mockRequest("hello")
		req.Headers["Origin"] = "google"
		err = policy.Inspect("lambda-1", req)
		assert.NoError(t, err)
	})
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
