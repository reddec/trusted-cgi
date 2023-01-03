package internal

import (
	"context"
	"github.com/hashicorp/go-multierror"
)

type DaemonFunc func(ctx context.Context) error

type DaemonHandler interface {
	Run(ctx context.Context) error
}

func (df DaemonFunc) Run(ctx context.Context) error {
	return df(ctx)
}

func Spawn(fn DaemonHandler) *Daemon {
	return SpawnContext(context.Background(), fn)
}

func SpawnContext(global context.Context, fn DaemonHandler) *Daemon {
	ctx, cancel := context.WithCancel(global)

	d := &Daemon{
		done:   make(chan struct{}),
		cancel: cancel,
	}

	go func() {
		defer cancel()
		defer close(d.done)
		d.err = fn.Run(ctx)
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

func NewDaemonSet(fragile bool) *DaemonSet {
	return &DaemonSet{fragile: fragile}
}

type DaemonSet struct {
	funcs   []DaemonHandler
	fragile bool
}

func (ds *DaemonSet) Start() *Daemon {
	return Spawn(ds)
}

func (ds *DaemonSet) Jobs() []DaemonHandler {
	return ds.funcs
}

func (ds *DaemonSet) Add(fn ...DaemonHandler) {
	ds.funcs = append(ds.funcs, fn...)
}

func (ds *DaemonSet) Run(ctx context.Context) error {
	var wg multierror.Group

	var cancel = func() {}

	if ds.fragile { // abort all if one process stopped
		ctx, cancel = context.WithCancel(ctx)
	}
	defer cancel()

	for _, fn := range ds.funcs {
		ref := fn
		wg.Go(func() error {
			defer cancel()
			return ref.Run(ctx)
		})
	}
	return wg.Wait().ErrorOrNil()
}
