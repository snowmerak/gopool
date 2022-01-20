package gopool

import (
	"log"
	"runtime"
	"sync"
	"sync/atomic"
)

var pool = sync.Pool{}
var count = int64(0)
var max = int64(2 << 40)

func init() {
	pool.New = func() interface{} {
		ch := make(chan func(), 1)
		atomic.AddInt64(&count, 1)
		go func() {
			defer func() {
				r := recover()
				if r != nil {
					log.Println(r)
				}
				atomic.AddInt64(&count, -1)
			}()
			for f := range ch {
				f()
				if atomic.LoadInt64(&count) > atomic.LoadInt64(&max) {
					close(ch)
					return
				}
				pool.Put(ch)
			}
		}()
		return ch
	}
}

func SetMax(n int64) {
	atomic.StoreInt64(&max, n)
}

func GetMax() int64 {
	return atomic.LoadInt64(&max)
}

func GetCurrnet() int64 {
	return atomic.LoadInt64(&count)
}

func Go(f func()) {
	for atomic.LoadInt64(&count) > atomic.LoadInt64(&max) {
		runtime.Gosched()
	}
	ch := pool.Get().(chan func())
	ch <- f
}
