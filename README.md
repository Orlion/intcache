# intcache
A Golang thread-safe lock-free uint32 cache that supports LRU for specific scenarios

一个Golang无锁并且支持LRU的线程安全的uint32缓存

# Benchmark
```
goos: darwin
goarch: arm64
pkg: github.com/Orlion/intcache
BenchmarkIntcache
BenchmarkIntcache-10                  14          78865301 ns/op
BenchmarkFastcache
BenchmarkFastcache-10                 10         113746746 ns/op
PASS
ok      github.com/Orlion/intcache      9.767s
```

# Usage
```
cache := intcache.New(21)
cache.Set(123, 123)
cache.Get(123)
cache.Set(456, 456)
cache.Get(456)
```