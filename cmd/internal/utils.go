package internal

import (
	"context"
	"os"
	"os/signal"
)

func SignalContext() (context.Context, func()) {
	gctx, closer := context.WithCancel(context.Background())
	go func() {
		c := make(chan os.Signal, 2)
		signal.Notify(c, os.Kill, os.Interrupt)
		for range c {
			closer()
			break
		}
	}()
	return gctx, closer
}
