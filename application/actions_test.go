package application

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestApp_ListActions(t *testing.T) {
	app := &App{
		location: "../",
	}
	list, err := app.ListActions()
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, []string{"clean", "bindata", "ui/src", "ui/dist", "embed_ui"}, list)
}
