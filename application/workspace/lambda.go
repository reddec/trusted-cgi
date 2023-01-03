package workspace

import (
	"context"
	"fmt"
	"github.com/reddec/trusted-cgi/application/config"
	"github.com/reddec/trusted-cgi/application/stats"
	"github.com/reddec/trusted-cgi/internal"
	"github.com/reddec/trusted-cgi/types"
	"io"
	"os"
	"os/exec"
	"time"
)

// NewLambda creates lambda from config. Credentials are used for running lambda and can be nil.
func NewLambda(config config.Lambda, creds *types.Credential, monitor *stats.LambdaMonitor) (*Lambda, error) {
	if len(config.Exec) == 0 {
		return nil, fmt.Errorf("at least one argument should be in exec")
	}
	return &Lambda{
		monitor: monitor,
		binary:  config.Exec[0],
		args:    config.Exec[1:],
		workDir: config.WorkDir,
		timeout: config.Timeout.Duration(),
		creds:   creds,
	}, nil
}

type Lambda struct {
	monitor *stats.LambdaMonitor
	binary  string
	args    []string
	workDir string
	timeout time.Duration
	creds   *types.Credential
}

// Invoke lambda. Close MUST be called.
func (pl *Lambda) Invoke(global context.Context, environment map[string]string, body io.Reader) (io.ReadCloser, error) {
	ctx, cancel := pl.createContext(global)
	rd, wr := io.Pipe()

	cmd := exec.CommandContext(ctx, pl.binary, pl.args...)
	cmd.Dir = pl.workDir
	cmd.Stdin = body
	cmd.Stdout = wr
	cmd.Stderr = os.Stderr
	if pl.creds != nil {
		internal.SetCreds(cmd, pl.creds)
	}
	internal.SetFlags(cmd)
	var env = os.Environ()
	for k, v := range environment {
		env = append(env, k+"="+v)
	}
	cmd.Env = env
	stat := pl.monitor.Started(env)
	if err := cmd.Start(); err != nil {
		cancel()
		_ = wr.Close()
		_ = rd.Close()
		_ = stat.Finished(err)
		return nil, fmt.Errorf("start lambda: %w", err)
	}

	done := make(chan error, 1)
	go func() {
		defer close(done)
		err := cmd.Wait()
		_ = stat.Finished(err)
		_ = wr.CloseWithError(err)
		done <- err
	}()

	return &cancelCloseReader{
		ReadCloser: rd,
		cancel:     cancel,
		done:       done,
	}, nil
}

func (pl *Lambda) createContext(global context.Context) (context.Context, context.CancelFunc) {
	if pl.timeout > 0 {
		return context.WithTimeout(global, pl.timeout)
	}
	return context.WithCancel(global)
}

type cancelCloseReader struct {
	io.ReadCloser
	done   chan error
	cancel func()
}

func (c *cancelCloseReader) Close() error {
	c.cancel()
	err := c.ReadCloser.Close()
	<-c.done
	return err
}
