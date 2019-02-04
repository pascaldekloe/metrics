[![API Documentation](https://godoc.org/github.com/pascaldekloe/metrics?status.svg)](https://godoc.org/github.com/pascaldekloe/metrics)

A Prometheus exposition library for the Go programming language.

This is free and unencumbered software released into the
[public domain](https://creativecommons.org/publicdomain/zero/1.0).


### Performance on a Mac Pro (late 2013)

```
name                             time/op
ParallelAdd/integer-12           23.6ns ± 0%
ParallelAdd/label2x5-12           158ns ± 3%
HTTPHandler/32/integer-12        1.67µs ± 0%
HTTPHandler/32/real-12           3.13µs ± 0%
HTTPHandler/32/label5-12         3.96µs ± 0%
HTTPHandler/32/label2x3x5-12     3.98µs ± 0%
HTTPHandler/1024/integer-12      32.4µs ± 2%
HTTPHandler/1024/real-12         73.3µs ± 1%
HTTPHandler/1024/label5-12       97.6µs ± 1%
HTTPHandler/1024/label2x3x5-12    100µs ± 1%
HTTPHandler/32768/integer-12     1.07ms ± 2%
HTTPHandler/32768/real-12        2.41ms ± 1%
HTTPHandler/32768/label5-12      3.24ms ± 1%
HTTPHandler/32768/label2x3x5-12  3.40ms ± 0%

name                             alloc/op
ParallelAdd/integer-12            0.00B     
ParallelAdd/label2x5-12           0.00B     
HTTPHandler/32/integer-12        4.11kB ± 0%
HTTPHandler/32/real-12           4.11kB ± 0%
HTTPHandler/32/label5-12         4.11kB ± 0%
HTTPHandler/32/label2x3x5-12     4.11kB ± 0%
HTTPHandler/1024/integer-12      4.11kB ± 0%
HTTPHandler/1024/real-12         4.11kB ± 0%
HTTPHandler/1024/label5-12       4.11kB ± 0%
HTTPHandler/1024/label2x3x5-12   4.11kB ± 0%
HTTPHandler/32768/integer-12     4.12kB ± 0%
HTTPHandler/32768/real-12        4.12kB ± 0%
HTTPHandler/32768/label5-12      4.12kB ± 0%
HTTPHandler/32768/label2x3x5-12  4.12kB ± 0%

name                             allocs/op
ParallelAdd/integer-12             0.00     
ParallelAdd/label2x5-12            0.00     
HTTPHandler/32/integer-12          2.00 ± 0%
HTTPHandler/32/real-12             2.00 ± 0%
HTTPHandler/32/label5-12           2.00 ± 0%
HTTPHandler/32/label2x3x5-12       2.00 ± 0%
HTTPHandler/1024/integer-12        2.00 ± 0%
HTTPHandler/1024/real-12           2.00 ± 0%
HTTPHandler/1024/label5-12         2.00 ± 0%
HTTPHandler/1024/label2x3x5-12     2.00 ± 0%
HTTPHandler/32768/integer-12       2.00 ± 0%
HTTPHandler/32768/real-12          2.00 ± 0%
HTTPHandler/32768/label5-12        2.00 ± 0%
HTTPHandler/32768/label2x3x5-12    2.00 ± 0%
```
