[![API Documentation](https://godoc.org/github.com/pascaldekloe/metrics?status.svg)](https://godoc.org/github.com/pascaldekloe/metrics)

A Prometheus exposition library for the Go programming language.

This is free and unencumbered software released into the
[public domain](https://creativecommons.org/publicdomain/zero/1.0).


### Performance on a Mac Pro (late 2013)

```
name                           time/op
Label/sequential/4-12            23.5ns ± 2%
Label/sequential/4x4-12          27.9ns ± 3%
Label/sequential/4x4x4-12        49.7ns ± 1%
Label/parallel/4-12               110ns ± 0%
Label/parallel/4x4-12             123ns ± 0%
Label/parallel/4x4x4-12           154ns ± 0%
Get/counter/sequential-12        0.26ns ± 0%
Get/counter/2routines-12         0.18ns ± 0%
Get/real/sequential-12           0.52ns ± 0%
Get/real/2routines-12            0.20ns ± 0%
Get/sample/sequential-12         1.04ns ± 0%
Get/sample/2routines-12          0.28ns ± 3%
Set/real/sequential-12           5.69ns ± 0%
Set/real/2routines-12            21.2ns ± 1%
Set/sample/sequential-12         33.4ns ± 1%
Set/sample/2routines-12          37.2ns ±11%
Add/counter/sequential-12        5.69ns ± 0%
Add/counter/2routines-12         19.2ns ± 0%
Add/integer/sequential-12        5.69ns ± 0%
Add/integer/2routines-12         22.6ns ± 1%
Add/histogram5/sequential-12     29.4ns ± 0%
Add/histogram5/2routines-12      74.7ns ± 0%
ServeHTTP/32/counter-12          2.10µs ± 0%
ServeHTTP/32/real-12             3.36µs ± 0%
ServeHTTP/32/integer-12          2.11µs ± 0%
ServeHTTP/32/histogram5-12       12.4µs ± 1%
ServeHTTP/32/label5-12           4.01µs ± 0%
ServeHTTP/32/label2x3x5-12       4.05µs ± 0%
ServeHTTP/32/sample-12           3.44µs ± 0%
ServeHTTP/1024/counter-12        40.4µs ± 0%
ServeHTTP/1024/real-12           81.9µs ± 2%
ServeHTTP/1024/integer-12        43.9µs ± 1%
ServeHTTP/1024/histogram5-12      360µs ± 1%
ServeHTTP/1024/label5-12          101µs ± 3%
ServeHTTP/1024/label2x3x5-12      101µs ± 3%
ServeHTTP/1024/sample-12         82.3µs ± 1%
ServeHTTP/32768/counter-12       1.36ms ± 2%
ServeHTTP/32768/real-12          2.74ms ± 4%
ServeHTTP/32768/integer-12       1.46ms ± 1%
ServeHTTP/32768/histogram5-12    13.5ms ± 1%
ServeHTTP/32768/label5-12        3.49ms ± 2%
ServeHTTP/32768/label2x3x5-12    3.56ms ± 1%
ServeHTTP/32768/sample-12        2.81ms ± 0%

name                           speed
ServeHTTP/32/counter-12         927MB/s ± 0%
ServeHTTP/32/real-12            503MB/s ± 0%
ServeHTTP/32/integer-12         892MB/s ± 0%
ServeHTTP/32/histogram5-12      701MB/s ± 1%
ServeHTTP/32/label5-12          605MB/s ± 0%
ServeHTTP/32/label2x3x5-12      782MB/s ± 0%
ServeHTTP/32/sample-12          492MB/s ± 0%
ServeHTTP/1024/counter-12      1.61GB/s ± 0%
ServeHTTP/1024/real-12          698MB/s ± 2%
ServeHTTP/1024/integer-12      1.44GB/s ± 1%
ServeHTTP/1024/histogram5-12    794MB/s ± 1%
ServeHTTP/1024/label5-12        801MB/s ± 3%
ServeHTTP/1024/label2x3x5-12   1.04GB/s ± 3%
ServeHTTP/1024/sample-12        694MB/s ± 1%
ServeHTTP/32768/counter-12     1.67GB/s ± 2%
ServeHTTP/32768/real-12         731MB/s ± 4%
ServeHTTP/32768/integer-12     1.51GB/s ± 1%
ServeHTTP/32768/histogram5-12   714MB/s ± 1%
ServeHTTP/32768/label5-12       788MB/s ± 2%
ServeHTTP/32768/label2x3x5-12   985MB/s ± 1%
ServeHTTP/32768/sample-12       711MB/s ± 0%
```
