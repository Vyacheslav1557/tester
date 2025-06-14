package tester

import (
	"context"
	"fmt"
	"sync"
)

type Pool[T interface{}] struct {
	ch chan T
	wg sync.WaitGroup
	f  func(T)
}

func NewPool[T interface{}](n int, f func(T)) *Pool[T] {
	r := &Pool[T]{
		ch: make(chan T, n),
		f:  f,
	}

	r.wg.Add(n)
	for i := 0; i < n; i++ {
		go r.newWorker()
	}

	return r
}

func (p *Pool[T]) newWorker() {
	defer p.wg.Done()

	for task := range p.ch {
		p.f(task)
	}
}

func (p *Pool[T]) Close() {
	close(p.ch)
	p.wg.Wait()
}

func (p *Pool[T]) Do(ctx context.Context, task T) (err error) {
	defer func() {
		s := recover()
		if s != nil {
			err = fmt.Errorf("%v", s)
		}
	}()

	select {
	case p.ch <- task:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
