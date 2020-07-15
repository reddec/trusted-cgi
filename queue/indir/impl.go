package indir

import (
	"context"
	"github.com/reddec/dfq"
	"github.com/reddec/trusted-cgi/types"
	"github.com/tinylib/msgp/msgp"
	"io"
)

func New(directory string) (*inDirQueue, error) {
	back, err := dfq.Open(directory)
	if err != nil {
		return nil, err
	}
	return &inDirQueue{backend: back}, nil
}

type inDirQueue struct {
	backend dfq.Queue
}

func (queue *inDirQueue) Put(ctx context.Context, request *types.Request) error {
	defer request.Body.Close()
	return queue.backend.Stream(func(out io.Writer) error {
		w := msgp.NewWriter(out)
		err := request.EncodeMsg(w)
		if err != nil {
			return err
		}
		_, err = io.Copy(w, request.Body)
		if err != nil {
			return err
		}
		return w.Flush()
	})
}

func (queue *inDirQueue) Peek(ctx context.Context) (*types.Request, error) {
	in, err := queue.backend.Wait(ctx)
	if err != nil {
		return nil, err
	}
	reader := msgp.NewReader(in)
	var head types.Request
	err = head.DecodeMsg(reader)
	if err != nil {
		_ = in.Close()
		return nil, err
	}
	return head.WithBody(&readCloser{reader: reader.R, closer: in}), nil
}

func (queue *inDirQueue) Commit(ctx context.Context) error {
	return queue.backend.Commit()
}

type readCloser struct {
	reader io.Reader
	closer io.Closer
}

func (rc *readCloser) Read(p []byte) (n int, err error) {
	return rc.reader.Read(p)
}

func (rc *readCloser) Close() error {
	return rc.closer.Close()
}
