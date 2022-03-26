package gopool

import (
	"runtime"
	"sync/atomic"
)

type lock int64

func (l *lock) Lock() {
	for {
		if atomic.CompareAndSwapInt64((*int64)(l), 0, 1) {
			return
		}
		runtime.Gosched()
	}
}

func (l *lock) Unlock() {
	atomic.StoreInt64((*int64)(l), 0)
}

type GoPool[T any] struct {
	list      []chan parameter[T]
	isRunning []bool
	running   int64
	lock
}

type Result[T any] struct {
	Result T
	Err    error
}

type parameter[T any] struct {
	f  func() Result[T]
	ch chan<- Result[T]
}

func do[T any](gp *GoPool[T], ch chan parameter[T], index int) {
	var resultChan chan<- Result[T]
	defer func() {
		if err := recover(); err != nil {
			if resultChan != nil {
				var t T
				resultChan <- Result[T]{t, err.(error)}
			}
		}
		gp.Lock()
		gp.running--
		gp.isRunning[index] = false
		gp.Unlock()
		go do(gp, ch, index)
	}()
	for p := range ch {
		resultChan = p.ch
		rs := p.f()
		p.ch <- rs
		gp.Lock()
		gp.running--
		gp.isRunning[index] = false
		gp.Unlock()
	}
}

func New[T any](max int) *GoPool[T] {
	gp := new(GoPool[T])
	gp.list = make([]chan parameter[T], max)
	for i := 0; i < max; i++ {
		gp.list[i] = make(chan parameter[T])
		go do(gp, gp.list[i], i)
	}
	gp.isRunning = make([]bool, max)
	gp.running = 0
	gp.lock = 0
	return gp
}

func (gp *GoPool[T]) Go(f func() Result[T]) (<-chan Result[T], error) {
	for {
		gp.Lock()
		for i := 0; i < len(gp.list); i++ {
			if !gp.isRunning[i] {
				gp.isRunning[i] = true
				resultChan := make(chan Result[T], 1)
				gp.list[i] <- parameter[T]{f, resultChan}
				gp.running++
				gp.Unlock()
				return resultChan, nil
			}
		}
		gp.Unlock()
		runtime.Gosched()
	}
}

func (gp *GoPool[T]) Wait() {
	for atomic.LoadInt64(&gp.running) > 0 {
		runtime.Gosched()
	}
}

func (gp *GoPool[T]) Close() {
	gp.Lock()
	defer gp.Unlock()
	for i := 0; i < len(gp.list); i++ {
		close(gp.list[i])
	}
}
