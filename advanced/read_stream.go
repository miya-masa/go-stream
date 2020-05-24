package main

import (
	"context"
	"io"
)

type ReadStream struct {
	inCh   chan interface{}
	doneCh chan struct{}
}

func (r ReadStream) Read() (interface{}, error) {
	return r.ReadContext(context.Background())
}

func (r ReadStream) ReadContext(ctx context.Context) (interface{}, error) {
	select {
	case <-ctx.Done():
		return nil, io.EOF
	case <-r.doneCh:
		return nil, io.EOF
	case v, ok := <-r.inCh:
		if !ok {
			return nil, io.EOF
		}
		return v, nil
	}
}
