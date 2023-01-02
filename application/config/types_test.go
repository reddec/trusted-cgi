package config_test

import (
	"github.com/reddec/trusted-cgi/application/config"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

const demo = `
static = "dist/static" // where to store static files

//  endpoints definitions

get "time" { // get/post/put/patch/delete
    body = 65535 // max body size
    // response headers
    headers = {
        X-Request-Id = "headers.X-Forwarded-For"
    }
    status = 200 // by default, for enqueue only it's 201, for call 200
    
    // trusted CGI will first save request to storage or in memory (depends on size)

    enqueue "fetch-video" {} // multiple pushes to the queue allowed
    enqueue "log" {}
    call "time" {
        payload = "asdasdasd" // payload can be overriden
    }
    call "format" {
        environment  = {
            FORWARDED_FOR  = "{{Header.X-Forwarded-For}}"
            LIMIT          = "{{Query.limit}}"
        }
    } // multiple calls allowed, response will be concatenated from the all calls results
}

get "" {
    call "time" {}
}

lambda "time" {
    exec    = ["time"]
    timeout = 30
    workDir = "/" // relative to the project dir
}

queue "fetch-video" {
    size = 1024
    interval = 10
    retry = -1
    lambda = "time"
}

cron "* * * * *" {
    enqueue "time" {
        payload = "another"
    }
    call "time" {
        payload = "hello world" // custom payload to trigger 
        environment = {
            X = "1"
        }
    }
}
`

func TestParse(t *testing.T) {
	f, err := os.CreateTemp("", "*.hcl")
	require.NoError(t, err)
	defer os.RemoveAll(f.Name())

	_, err = f.WriteString(demo)
	require.NoError(t, err)

	err = f.Close()
	require.NoError(t, err)

	p, err := config.ParseFile(f.Name())
	require.NoError(t, err)

	t.Logf("%+v", p)
}
