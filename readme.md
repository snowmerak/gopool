# GoPool

gopool is simple goroutione pool

## HowTo

```bash
go get github.com/snowmerak/gopool
```

```go
package main

import "github.com/snowmerak/gopool"

func main() {
    // set logger of gopool
    gopool.SetLogger(log.New(os.Stdout, "gopool: ", log.LstdFlags))

    // Set max goroutine counts
	gopool.SetMax(100)
	wg := sync.WaitGroup{}
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		gopool.Go(func() {
			time.Sleep(time.Second)
			wg.Done()
		})
	}
	for i := 0; i < 5; i++ {
        // can recover panic
		gopool.Go(func() {
			panic("any error")
		})
	}
	wg.Wait()

    // print current goroutines
	fmt.Println(gopool.GetCurrnet())
}
```