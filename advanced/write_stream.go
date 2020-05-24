package main

import (
	"context"
	"io"
	"sync"
)

type WriteStream struct {
	sync.Mutex

	inCh   chan interface{}
	once   sync.Once
	doneCh chan struct{}
}

func (s *WriteStream) Write(num interface{}) error {
	return s.WriteContext(context.Background(), num)
}

func (s *WriteStream) WriteContext(ctx context.Context, num interface{}) error {
	select {
	case <-ctx.Done():
		return io.EOF
	case s.inCh <- num:
		return nil
	case <-s.doneCh:
		return io.EOF
	}
}

func (s *WriteStream) Close() {
	s.once.Do(func() {
		close(s.doneCh)
	})
}

