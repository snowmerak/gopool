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

type GoPool struct {
	list      []chan parameter
	isRunning []bool
	running   int64
	lock
}

type parameter struct {
	f  func() interface{}
	ch chan<- interface{}
}

func do(gp *GoPool, ch chan parameter, index int) {
	var resultChan chan<- interface{}
	defer func() {
		if err := recover(); err != nil {
			if resultChan != nil {
				resultChan <- err
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

func New(max int) *GoPool {
	gp := new(GoPool)
	gp.list = make([]chan parameter, max)
	for i := 0; i < max; i++ {
		gp.list[i] = make(chan parameter)
		go do(gp, gp.list[i], i)
	}
	gp.isRunning = make([]bool, max)
	gp.running = 0
	gp.lock = 0
	return gp
}

func (gp *GoPool) Go(f func() interface{}) (<-chan interface{}, error) {
	for {
		gp.Lock()
		for i := 0; i < len(gp.list); i++ {
			if !gp.isRunning[i] {
				gp.isRunning[i] = true
				resultChan := make(chan interface{}, 1)
				gp.list[i] <- parameter{f, resultChan}
				gp.running++
				gp.Unlock()
				return resultChan, nil
			}
		}
		gp.Unlock()
		runtime.Gosched()
	}
}

func (gp *GoPool) Wait() {
	for atomic.LoadInt64(&gp.running) > 0 {
		runtime.Gosched()
	}
}

func (gp *GoPool) Close() {
	gp.Lock()
	defer gp.Unlock()
	for i := 0; i < len(gp.list); i++ {
		close(gp.list[i])
	}
}
