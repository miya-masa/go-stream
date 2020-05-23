package main

import (
	"context"
	"fmt"
	"time"
)

func main() {

	normal()
	withTimeout()
	delegate()
	delegateWithTimeout()
	delegateWrapper()
	delegateWrapperWithTimeout()
}

func delegateWrapper() {
	fmt.Printf("%s\n", "Start Stream Delegate Wrapper")
	defer fmt.Printf("End\n")
	st := Make()
	defer st.Close()
	for i := 0; i < 5; i++ {
		if err := st.Send(i); err != nil {
			fmt.Printf("err = %+v\n", err)
			return
		}
	}
}

func delegateWrapperWithTimeout() {
	fmt.Printf("%s\n", "Start Stream DelegateWrapperWithTimeout")
	defer fmt.Printf("End\n")

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*25)
	defer cancel()
	st := Make()
	defer st.Close()
	for i := 0; i < 5; i++ {
		if err := st.SendContext(ctx, i); err != nil {
			return
		}
	}
}

func delegateWithTimeout() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*25)
	defer cancel()
	fmt.Printf("%s\n", "Start Stream Delegate With Timeout")
	defer fmt.Printf("End\n")
	ch, doneCh := startDelegate()
	defer func() {
		// 必ずchをクローズしてからdoneを待つ
		close(ch)
		<-doneCh
	}()
	for i := 0; i < 5; i++ {
		select {
		case <-ctx.Done():
			return
		case ch <- i:
		}
	}
}

func delegate() {
	fmt.Printf("%s\n", "Start Stream Delegate")
	defer fmt.Printf("End\n")
	ch, doneCh := startDelegate()
	defer func() {
		// 必ずchをクローズしてからdoneを待つ
		close(ch)
		<-doneCh
	}()
	for i := 0; i < 5; i++ {
		ch <- i
	}
}

func withTimeout() {
	fmt.Printf("%s\n", "Start Stream Timeout")
	defer fmt.Printf("End\n")
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*25)
	defer cancel()
	intCh := start(ctx, 5)
	intCh = multiply(intCh, 3)
	for num := range intCh {
		fmt.Printf("result = %+v\n", num)
	}
}

func normal() {
	fmt.Printf("%s\n", "Start Stream")
	defer fmt.Printf("End\n")
	ctx := context.Background()
	intCh := start(ctx, 5)
	intCh = multiply(intCh, 3)
	for num := range intCh {
		fmt.Printf("result = %+v\n", num)
	}
}

func start(ctx context.Context, num int) <-chan int {
	intCh := make(chan int, 0)
	go func() {
		defer close(intCh)
		for i := 0; i < num; i++ {
			time.Sleep(time.Millisecond * 10)
			select {
			case <-ctx.Done():
				return
			case intCh <- i:
			}
		}
	}()
	return intCh
}

func multiply(inCh <-chan int, num int) <-chan int {
	intCh := make(chan int, 0)
	go func() {
		defer close(intCh)
		for v := range inCh {
			intCh <- v * num
		}
	}()
	return intCh
}

func startDelegate() (chan int, <-chan struct{}) {
	intCh := make(chan int, 0)
	doneCh := make(chan struct{}, 0)
	go func() {
		defer close(doneCh)
		for v := range intCh {
			time.Sleep(time.Millisecond * 10)
			fmt.Printf("result = %+v\n", v)
		}
	}()
	return intCh, doneCh
}

func startDelegateWrapper() Stream {
	return Make()
}

// write channelのWrapperと割り切る
// stopの後のsendはpanic
// 原則、writeとstopは同じスコープで扱うことを前提とする（= write channel と同じお作法）
type Stream struct {
	inCh   chan int
	doneCh chan struct{}
}

func (s Stream) Send(num int) error {
	s.inCh <- num
	return nil
}

func (s Stream) SendContext(ctx context.Context, num int) error {
	select {
	case <-ctx.Done():
		return context.Canceled
	case s.inCh <- num:
	}
	return nil
}

func (s Stream) Close() {
	// 必ずchをクローズしてからdoneを待つ
	close(s.inCh)
	<-s.doneCh
}

func Make() Stream {
	intCh := make(chan int, 0)
	doneCh := make(chan struct{}, 0)
	go func() {
		defer close(doneCh)
		for v := range intCh {
			time.Sleep(time.Millisecond * 10)
			fmt.Printf("result = %+v\n", v)
		}
	}()
	return Stream{
		inCh:   intCh,
		doneCh: doneCh,
	}
}
