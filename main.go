package gopool

import (
	"runtime"
	"sync"
	"sync/atomic"
)

type GoPool struct {
	pool    sync.Pool
	max     int64
	count   int64
	running int64
	sync.Mutex
}

type parameter struct {
	f  func() interface{}
	ch chan<- interface{}
}

func New(max int64) *GoPool {
	gp := &GoPool{}
	gp.pool.New = func() interface{} {
		ch := make(chan parameter, 1)
		atomic.AddInt64(&gp.count, 1)
		go func() {
			defer close(ch)
			param := parameter{}
			defer func() {
				r := recover()
				param.ch <- r
				atomic.AddInt64(&gp.count, -1)
				atomic.AddInt64(&gp.running, -1)
			}()
			for p := range ch {
				param = p
				rs := p.f()
				gp.Lock()
				if gp.count > gp.max {
					gp.Unlock()
					return
				}
				gp.Unlock()
				p.ch <- rs
				close(p.ch)
				gp.pool.Put(ch)
				atomic.AddInt64(&gp.running, -1)
			}
		}()
		return ch
	}
	gp.max = max
	gp.count = 0
	return gp
}

func (gp *GoPool) SetMax(n int64) {
	atomic.StoreInt64(&gp.max, n)
}

func (gp *GoPool) GetMax() int64 {
	return atomic.LoadInt64(&gp.max)
}

func (gp *GoPool) GetCurrnet() int64 {
	return atomic.LoadInt64(&gp.count)
}

func (gp *GoPool) Go(f func() interface{}) <-chan interface{} {
	for {
		gp.Lock()
		if gp.running < gp.max {
			gp.Unlock()
			break
		}
		gp.Unlock()
		runtime.Gosched()
	}
	atomic.AddInt64(&gp.running, 1)
	ch := gp.pool.Get().(chan parameter)
	rs := make(chan interface{}, 1)
	ch <- parameter{
		f:  f,
		ch: rs,
	}
	return rs
}

func (gp *GoPool) Wait() {
	for atomic.LoadInt64(&gp.running) > 0 {
		runtime.Gosched()
	}
}
