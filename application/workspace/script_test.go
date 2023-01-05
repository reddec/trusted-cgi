package workspace_test

import (
	"context"
	"github.com/reddec/trusted-cgi/application/config"
	"github.com/reddec/trusted-cgi/application/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"testing"
)

func TestNew(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		cfg := config.Lambda{
			Exec:    []string{"echo", "1234"},
			Timeout: 10,
			WorkDir: "/tmp",
		}

		lm, err := workspace.NewLambda(cfg, nil, nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, lm)
	})

	t.Run("failed initialize - no exec", func(t *testing.T) {
		cfg := config.Lambda{
			Exec:    []string{},
			Timeout: 10,
			WorkDir: "/tmp",
		}

		_, err := workspace.NewLambda(cfg, nil, nil)
		assert.Error(t, err)
	})
}

func TestLambda_Invoke(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		cfg := config.Lambda{
			Exec:    []string{"echo", "1234"},
			Timeout: 10,
			WorkDir: "/tmp",
		}

		lm, err := workspace.NewLambda(cfg, nil, nil)
		require.NoError(t, err)

		out, err := lm.Invoke(context.Background(), nil, nil)
		require.NoError(t, err)
		defer out.Close()

		data, err := io.ReadAll(out)
		require.NoError(t, err)

		assert.Equal(t, "1234\n", string(data))
	})

	t.Run("Timeout", func(t *testing.T) {
		cfg := config.Lambda{
			Exec:    []string{"sleep", "1000"},
			Timeout: 1,
			WorkDir: "/tmp",
		}

		lm, err := workspace.NewLambda(cfg, nil, nil)
		require.NoError(t, err)

		out, err := lm.Invoke(context.Background(), nil, nil)
		require.NoError(t, err)
		defer out.Close()

		_, err = io.ReadAll(out)
		assert.Error(t, err)
	})
}
