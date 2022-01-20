# GoPool

gopool is simple goroutione pool

## HowTo

```bash
go get github.com/snowmerak/gopool
```

```go
package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/snowmerak/gopool"
)

func main() {
	// create new logger
	logger := log.New(os.Stderr, "gopool: ", log.LstdFlags)

	// create new gopool with max size
	gp := gopool.New(200)
	s := time.Now()
	for i := 0; i < 1000; i++ {
		// execute function through goroutine
		gp.Go(func() interface{} {
			time.Sleep(time.Millisecond * 100)
			return nil
		})
	}

	// wait gopool's all goroutines are stopped
	gp.Wait()
	e := time.Now()
	fmt.Println(e.Sub(s))

	// print current goroutine number
	fmt.Println(gp.GetCurrnet())

	// return panic value
	// can receive function's returning value
	ret := gp.Go(func() interface{} {
		panic("test")
	})
	logger.Println(<-ret)
}

```