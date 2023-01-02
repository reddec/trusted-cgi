package internal

import "context"

type DemonFunc func(ctx context.Context) error

func Spawn(fn DemonFunc) *Daemon {
	return SpawnContext(context.Background(), fn)
}

func SpawnContext(global context.Context, fn DemonFunc) *Daemon {
	ctx, cancel := context.WithCancel(global)

	d := &Daemon{
		done:   make(chan struct{}),
		cancel: cancel,
	}

	go func() {
		defer cancel()
		defer close(d.done)
		d.err = fn(ctx)
	}()

	return d
}

type Daemon struct {
	err    error
	done   chan struct{}
	cancel func()
}

func (d *Daemon) Stop() {
	d.cancel()
	<-d.done
}

func (d *Daemon) Error() error {
	return d.err
}
