[![API Documentation](https://godoc.org/github.com/pascaldekloe/metrics?status.svg)](https://godoc.org/github.com/pascaldekloe/metrics)

A Prometheus exposition library for the Go programming language.

This is free and unencumbered software released into the
[public domain](https://creativecommons.org/publicdomain/zero/1.0).


### Performance on a Mac Pro (late 2013)

```
name                             time/op
MustPlace-12                     30.0ns ± 1%
Help-12                           155ns ± 0%
ParallelSet/real-12              17.7ns ± 0%
ParallelSet/label2x5-12           149ns ± 1%
ParallelAdd/integer-12           19.3ns ± 0%
ParallelAdd/real-12               141ns ± 0%
ParallelAdd/label2x5-12           154ns ± 0%
HTTPHandler/32/integer-12        1.89µs ± 0%
HTTPHandler/32/real-12           3.31µs ± 0%
HTTPHandler/32/label5-12         3.53µs ± 0%
HTTPHandler/32/label2x3x5-12     3.52µs ± 0%
HTTPHandler/1024/integer-12      42.9µs ± 2%
HTTPHandler/1024/real-12         84.5µs ± 1%
HTTPHandler/1024/label5-12       91.3µs ± 2%
HTTPHandler/1024/label2x3x5-12   94.5µs ± 2%
HTTPHandler/32768/integer-12     1.40ms ± 2%
HTTPHandler/32768/real-12        2.78ms ± 0%
HTTPHandler/32768/label5-12      3.00ms ± 1%
HTTPHandler/32768/label2x3x5-12  3.14ms ± 0%

name                             alloc/op
MustPlace-12                      0.00B     
Help-12                           96.0B ± 0%
ParallelSet/real-12               0.00B     
ParallelSet/label2x5-12           0.00B     
ParallelAdd/integer-12            0.00B     
ParallelAdd/real-12               0.00B     
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
MustPlace-12                       0.00     
Help-12                            2.00 ± 0%
ParallelSet/real-12                0.00     
ParallelSet/label2x5-12            0.00     
ParallelAdd/integer-12             0.00     
ParallelAdd/real-12                0.00     
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
