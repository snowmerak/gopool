# GoPool

gopool은 간단한 고루틴 풀입니다.

## HowTo

### 설치

다음 명령어를 프로젝트 루트 폴더에서 실행해주면 프로젝트에 의존성이 추가됩니다.

```bash
go get github.com/snowmerak/gopool
```

### gopool 선언

```go
max := 100
gp := gopool.New(max)
```

`max`는 실행될 수 있는 최대 고루틴 수를 의미합니다. `max`를 패러미터로 새로운 `gopool` 객체를 생성합니다.

### 고루틴 실행

```go
gp.Go(func() interface{} {
	time.Sleep(time.Nanosecond * 100)
	return nil
})
```

`Go()` 메서드를 통해 `func() interface{}` 함수를 넘겨주어 고루틴 내에서 실행하도록 합니다.

### 반환값 처리

```go
ch := gp.Go(func() interface{} {
	time.Sleep(time.Nanosecond * 100)
	return 0
})

println(<-ch)
```

```bash
0
```

`Go()` 메서드는 `interface{}` 타입을 전달하는 채널을 반환합니다. 매개변수로 넘겨주는 함수의 반환값을 전달하여 join같은 느낌으로 사용할 수 있습니다. 이는 반환값이 `nil`일 때도 성립됩니다.

### 전역 대기

```go
gp.Wait()
```

`Wait()` 메서드는 고루틴 풀을 통해 실행된 모든 고루틴이 동작을 멈출 때까지 대기합니다. `sync.WaitGroup`과 동일한 효과를 보입니다.

### panic & recover

```go
ch := gp.Go(func() interface{} {
	pamic("any error")
	return nil
})

println(<-ch)
```

```bash
any error
```

고루틴 내에서 패닉이 발생할 경우 내부에서 `recover`를 처리하여 반환값을 전달하는 채널로 던져줍니다. 그리고 전체 풀의 고루틴 수가 하나 줄어듭니다.
