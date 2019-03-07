[![API Documentation](https://godoc.org/github.com/pascaldekloe/metrics?status.svg)](https://godoc.org/github.com/pascaldekloe/metrics)

A Prometheus exposition library for the Go programming language.

Setup as `http.HandleFunc("/metrics", metrics.ServeHTTP)` and define metrics like `var RequestCount = metrics.MustCounter("rpc_requests_total", "Number of service invocations.")`.

This is free and unencumbered software released into the
[public domain](https://creativecommons.org/publicdomain/zero/1.0).


### Performance on a Mac Pro (late 2013)

```
name                           time/op
Label/sequential/4-12            22.9ns ± 1%
Label/sequential/4x4-12          28.2ns ± 0%
Label/sequential/4x4x4-12        49.3ns ± 1%
Label/parallel/4-12               100ns ± 0%
Label/parallel/4x4-12             125ns ± 0%
Label/parallel/4x4x4-12           154ns ± 0%
Get/histogram5/sequential-12      108ns ± 2%
Get/histogram5/2routines-12       389ns ± 0%
Set/real/sequential-12           5.69ns ± 0%
Set/real/2routines-12            16.3ns ± 0%
Set/sample/sequential-12         33.4ns ± 0%
Set/sample/2routines-12          35.7ns ± 9%
Add/counter/sequential-12        5.69ns ± 0%
Add/counter/2routines-12         19.9ns ± 1%
Add/integer/sequential-12        5.69ns ± 0%
Add/integer/2routines-12         17.2ns ± 0%
Add/histogram5/sequential-12     29.4ns ± 0%
Add/histogram5/2routines-12      73.6ns ± 0%
ServeHTTP/32/counter-12          2.10µs ± 0%
ServeHTTP/32/real-12             3.31µs ± 0%
ServeHTTP/32/integer-12          2.11µs ± 0%
ServeHTTP/32/histogram5-12       12.3µs ± 0%
ServeHTTP/32/label5-12           4.00µs ± 0%
ServeHTTP/32/label2x3x5-12       4.01µs ± 1%
ServeHTTP/32/sample-12           3.38µs ± 0%
ServeHTTP/1024/counter-12        40.4µs ± 0%
ServeHTTP/1024/real-12           78.8µs ± 1%
ServeHTTP/1024/integer-12        42.6µs ± 1%
ServeHTTP/1024/histogram5-12      358µs ± 1%
ServeHTTP/1024/label5-12          101µs ± 1%
ServeHTTP/1024/label2x3x5-12      103µs ± 3%
ServeHTTP/1024/sample-12         80.5µs ± 1%
ServeHTTP/32768/counter-12       1.35ms ± 3%
ServeHTTP/32768/real-12          2.60ms ± 1%
ServeHTTP/32768/integer-12       1.39ms ± 1%
ServeHTTP/32768/histogram5-12    12.7ms ± 1%
ServeHTTP/32768/label5-12        3.54ms ± 2%
ServeHTTP/32768/label2x3x5-12    3.62ms ± 1%
ServeHTTP/32768/sample-12        2.75ms ± 2%

name                           speed
ServeHTTP/32/counter-12         928MB/s ± 0%
ServeHTTP/32/real-12            510MB/s ± 0%
ServeHTTP/32/integer-12         893MB/s ± 0%
ServeHTTP/32/histogram5-12      704MB/s ± 0%
ServeHTTP/32/label5-12          607MB/s ± 0%
ServeHTTP/32/label2x3x5-12      789MB/s ± 1%
ServeHTTP/32/sample-12          500MB/s ± 0%
ServeHTTP/1024/counter-12      1.62GB/s ± 0%
ServeHTTP/1024/real-12          725MB/s ± 1%
ServeHTTP/1024/integer-12      1.48GB/s ± 1%
ServeHTTP/1024/histogram5-12    800MB/s ± 1%
ServeHTTP/1024/label5-12        799MB/s ± 1%
ServeHTTP/1024/label2x3x5-12   1.02GB/s ± 3%
ServeHTTP/1024/sample-12        709MB/s ± 1%
ServeHTTP/32768/counter-12     1.67GB/s ± 2%
ServeHTTP/32768/real-12         769MB/s ± 1%
ServeHTTP/32768/integer-12     1.58GB/s ± 1%
ServeHTTP/32768/histogram5-12   755MB/s ± 1%
ServeHTTP/32768/label5-12       778MB/s ± 2%
ServeHTTP/32768/label2x3x5-12   967MB/s ± 1%
ServeHTTP/32768/sample-12       727MB/s ± 2%
```
