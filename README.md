# What is it

It's a library with connection pool implementation and all connections in pool work simultaneously.
 It's important if you want send more requests than one connection may serve. For example, to send
 multiple email- or push-notifications fast.

On other hand it may be useful for site scrapping or to ddos someone ;)

## How to

There are some [examples](./examples/http-get-parallel/main.go) and here [too](./pkg/pool/implementation_test.go).

So to start you need implement your own:

- [connection](./pkg/connection/interfaces.go)
- [dialer](./pkg/connection/interfaces.go)
- [message](./pkg/message/interfaces.go)

and you are ready to download all internet:
```go
go func() {
    _ = p.Serve(inChan, outChan)
    close(outChan)
}()
```

### Benchmarks

It's too hard to make simple benchmarks, because project aims to
 improve performance in `IO`-bound tasks so, there are some metrics
 from my mac:

```bash
$ time go run examples/http-get-sequential/main.go -req-num=100
...
2019/08/31 21:04:58 Completed

real    0m6.245s
user    0m0.768s
sys     0m0.556s
```

```bash
$ time go run examples/http-get-parallel/main.go -conn-num=10 -req-num=100
...
2019/08/31 21:10:37 Completed

real    0m1.232s
user    0m0.762s
sys     0m0.469s
```

So in example on localhost pool wins in ~5 times.

But if we try to use as target, for example, google.com, than the difference will be greater:

```bash
time go run examples/http-get-sequential/main.go -req-num=100 -url="https://google.com"
...
2019/08/31 21:48:32 Completed

real    0m58.127s
user    0m1.064s
sys     0m0.653s
```

```bash
time go run examples/http-get-parallel/main.go -conn-num=20 -req-num=100 -url="https://google.com"
...
2019/08/31 21:46:39 Completed

real    0m5.300s
user    0m1.808s
sys     0m0.610s
```

The difference is obvious.
