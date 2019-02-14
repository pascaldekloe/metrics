[![API Documentation](https://godoc.org/github.com/pascaldekloe/metrics?status.svg)](https://godoc.org/github.com/pascaldekloe/metrics)

A Prometheus exposition library for the Go programming language.

This is free and unencumbered software released into the
[public domain](https://creativecommons.org/publicdomain/zero/1.0).


### Performance on a Mac Pro (late 2013)

```
name                             time/op
LabelWith/sequential/4-12          19.4ns ± 1%
LabelWith/sequential/4x4-12        26.5ns ± 1%
LabelWith/sequential/4x4x4-12      51.3ns ± 0%
LabelWith/parallel/4-12            87.5ns ± 0%
LabelWith/parallel/4x4-12           118ns ± 0%
LabelWith/parallel/4x4x4-12         160ns ± 0%
Get/counter/sequential-12          0.26ns ± 0%
Get/counter/2routines-12           0.18ns ± 0%
Get/gauge/sequential-12            0.52ns ± 0%
Get/gauge/2routines-12             0.19ns ± 0%
Get/sample/sequential-12           1.04ns ± 0%
Get/sample/2routines-12            0.28ns ± 0%
Set/gauge/sequential-12            5.71ns ± 0%
Set/gauge/2routines-12             17.7ns ± 0%
Set/sample/sequential-12           33.2ns ± 0%
Set/sample/2routines-12            36.1ns ± 6%
Add/counter/sequential-12          5.71ns ± 0%
Add/counter/2routines-12           19.2ns ± 0%
Add/gauge/sequential-12            10.1ns ± 1%
Add/gauge/2routines-12             17.7ns ± 0%
Add/histogram5/sequential-12       29.0ns ± 0%
Add/histogram5/2routines-12        80.1ns ± 0%
HTTPHandler/32/counter-12          2.17µs ± 0%
HTTPHandler/32/gauge-12            3.66µs ± 0%
HTTPHandler/32/histogram5-12       14.3µs ± 1%
HTTPHandler/32/label5-12           3.87µs ± 0%
HTTPHandler/32/label2x3x5-12       3.84µs ± 0%
HTTPHandler/32/sample-12           4.84µs ± 0%
HTTPHandler/1024/counter-12        44.2µs ± 2%
HTTPHandler/1024/gauge-12          87.8µs ± 1%
HTTPHandler/1024/histogram5-12      424µs ± 2%
HTTPHandler/1024/label5-12         95.6µs ± 1%
HTTPHandler/1024/label2x3x5-12     95.5µs ± 1%
HTTPHandler/1024/sample-12          123µs ± 2%
HTTPHandler/32768/counter-12       1.42ms ± 3%
HTTPHandler/32768/gauge-12         2.92ms ± 2%
HTTPHandler/32768/histogram5-12    14.9ms ± 1%
HTTPHandler/32768/label5-12        3.16ms ± 0%
HTTPHandler/32768/label2x3x5-12    3.25ms ± 2%
HTTPHandler/32768/sample-12        3.99ms ± 0%

name                             speed
HTTPHandler/32/counter-12        1.10GB/s ± 0%
HTTPHandler/32/gauge-12           585MB/s ± 0%
HTTPHandler/32/histogram5-12      822MB/s ± 0%
HTTPHandler/32/label5-12          742MB/s ± 0%
HTTPHandler/32/label2x3x5-12      940MB/s ± 0%
HTTPHandler/32/sample-12          468MB/s ± 0%
HTTPHandler/1024/counter-12      1.80GB/s ± 2%
HTTPHandler/1024/gauge-12         813MB/s ± 1%
HTTPHandler/1024/histogram5-12    911MB/s ± 2%
HTTPHandler/1024/label5-12        994MB/s ± 1%
HTTPHandler/1024/label2x3x5-12   1.24GB/s ± 1%
HTTPHandler/1024/sample-12        617MB/s ± 2%
HTTPHandler/32768/counter-12     1.92GB/s ± 3%
HTTPHandler/32768/gauge-12        841MB/s ± 2%
HTTPHandler/32768/histogram5-12   861MB/s ± 1%
HTTPHandler/32768/label5-12      1.02GB/s ± 0%
HTTPHandler/32768/label2x3x5-12  1.22GB/s ± 2%
HTTPHandler/32768/sample-12       648MB/s ± 0%
```
