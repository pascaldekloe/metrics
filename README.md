[![API Documentation](https://godoc.org/github.com/pascaldekloe/metrics?status.svg)](https://godoc.org/github.com/pascaldekloe/metrics)

A Prometheus exposition library for the Go programming language.

This is free and unencumbered software released into the
[public domain](https://creativecommons.org/publicdomain/zero/1.0).


### Performance on a Mac Pro (late 2013)

```
name                          time/op
HTTPHandler/0-metrics-12       139ns ± 1%
HTTPHandler/32-metrics-12     4.00µs ± 0%
HTTPHandler/1024-metrics-12    132µs ± 4%
HTTPHandler/32768-metrics-12  4.28ms ± 3%

name                          alloc/op
HTTPHandler/0-metrics-12       48.0B ± 0%
HTTPHandler/32-metrics-12      48.0B ± 0%
HTTPHandler/1024-metrics-12    48.0B ± 0%
HTTPHandler/32768-metrics-12   48.0B ± 0%

name                          allocs/op
HTTPHandler/0-metrics-12        3.00 ± 0%
HTTPHandler/32-metrics-12       3.00 ± 0%
HTTPHandler/1024-metrics-12     3.00 ± 0%
HTTPHandler/32768-metrics-12    3.00 ± 0%
```
