package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/gops/agent"
)

func main() {
	agent.Listen(agent.Options{})
	normal()
	withTimeout()
	for i := 0; i < 1000; i++ {
		takeWhileDo()
	}
}

func takeWhileDo() {
	fmt.Printf("Start takeWhile\n")
	defer fmt.Printf("End takeWhile\n")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r := generater(ctx)
	r = takeWhile(r, 5)
	for {
		v, err := r.Read()
		if err != nil {
			return
		}
		fmt.Printf("v = %+v\n", v)
	}
}

func withTimeout() {
	fmt.Printf("Start withTimeout\n")
	defer fmt.Printf("End withTimeout\n")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r := generater(ctx)
	r = multiply(r, 3)
	for {
		v, err := r.Read()
		if err != nil {
			return
		}
		fmt.Printf("v = %+v\n", v)
	}
}

func normal() {
	fmt.Printf("Start normal\n")
	defer fmt.Printf("End normal\n")
	r := counter()
	r = multiply(r, 3)
	for {
		v, err := r.Read()
		if err != nil {
			return
		}
		fmt.Printf("v = %+v\n", v)
	}
}

func generater(ctx context.Context) ReadStream {
	w, r := Pipe()
	go func() {
		defer w.Close()
		var count int
		for {
			time.Sleep(time.Millisecond * 100)
			if err := w.WriteContext(ctx, count); err != nil {
				return
			}
			count++
		}
	}()
	return r
}

func takeWhile(in ReadStream, until int) ReadStream {
	w, r := Pipe()
	go func() {
		defer w.Close()
		for {
			v, err := in.Read()
			if err != nil {
				return
			}
			i := v.(int)
			if i > until {
				return
			}
			if err := w.Write(i); err != nil {
				return
			}
		}
	}()
	return r
}

func multiply(in ReadStream, num int) ReadStream {
	w, r := Pipe()
	go func() {
		defer w.Close()
		for {
			v, err := in.Read()
			if err != nil {
				return
			}
			i := v.(int) * num
			if err := w.Write(i); err != nil {
				return
			}
		}
	}()
	return r
}

func counter() ReadStream {
	w, r := Pipe()
	go func() {
		defer w.Close()
		for i := 0; i < 5; i++ {
			time.Sleep(time.Millisecond * 100)
			if err := w.Write(i); err != nil {
				return
			}
		}
	}()
	return r
}

func Pipe() (WriteStream, ReadStream) {
	pipeCh := make(chan interface{}, 0)
	doneCh := make(chan struct{}, 0)
	return WriteStream{
			inCh:   pipeCh,
			doneCh: doneCh,
		}, ReadStream{
			inCh:   pipeCh,
			doneCh: doneCh,
		}
}
