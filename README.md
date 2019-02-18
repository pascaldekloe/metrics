[![API Documentation](https://godoc.org/github.com/pascaldekloe/metrics?status.svg)](https://godoc.org/github.com/pascaldekloe/metrics)

A Prometheus exposition library for the Go programming language.

This is free and unencumbered software released into the
[public domain](https://creativecommons.org/publicdomain/zero/1.0).


### Performance on a Mac Pro (late 2013)

```
name                             time/op
LabelWith/sequential/4-12          19.4ns ± 2%
LabelWith/sequential/4x4-12        26.6ns ± 1%
LabelWith/sequential/4x4x4-12      51.2ns ± 0%
LabelWith/parallel/4-12            89.5ns ± 0%
LabelWith/parallel/4x4-12           118ns ± 1%
LabelWith/parallel/4x4x4-12         160ns ± 0%
Get/counter/sequential-12          0.27ns ± 3%
Get/counter/2routines-12           0.19ns ± 0%
Get/gauge/sequential-12            0.52ns ± 0%
Get/gauge/2routines-12             0.20ns ± 0%
Get/sample/sequential-12           1.13ns ± 4%
Get/sample/2routines-12            0.28ns ± 0%
Set/gauge/sequential-12            5.69ns ± 0%
Set/gauge/2routines-12             18.6ns ± 0%
Set/sample/sequential-12           33.4ns ± 1%
Set/sample/2routines-12            36.4ns ± 8%
Add/counter/sequential-12          5.69ns ± 0%
Add/counter/2routines-12           26.9ns ± 1%
Add/gauge/sequential-12            10.1ns ± 0%
Add/gauge/2routines-12             17.9ns ± 1%
Add/histogram5/sequential-12       28.8ns ± 0%
Add/histogram5/2routines-12        79.2ns ± 0%
HTTPHandler/32/counter-12          2.36µs ± 0%
HTTPHandler/32/gauge-12            3.71µs ± 0%
HTTPHandler/32/histogram5-12       14.6µs ± 1%
HTTPHandler/32/label5-12           4.32µs ± 0%
HTTPHandler/32/label2x3x5-12       4.31µs ± 1%
HTTPHandler/32/sample-12           4.81µs ± 0%
HTTPHandler/1024/counter-12        49.6µs ± 2%
HTTPHandler/1024/gauge-12          87.0µs ± 1%
HTTPHandler/1024/histogram5-12      425µs ± 1%
HTTPHandler/1024/label5-12          108µs ± 2%
HTTPHandler/1024/label2x3x5-12      106µs ± 2%
HTTPHandler/1024/sample-12          121µs ± 1%
HTTPHandler/32768/counter-12       1.80ms ± 1%
HTTPHandler/32768/gauge-12         3.05ms ± 1%
HTTPHandler/32768/histogram5-12    14.9ms ± 1%
HTTPHandler/32768/label5-12        3.75ms ± 0%
HTTPHandler/32768/label2x3x5-12    3.83ms ± 1%
HTTPHandler/32768/sample-12        4.14ms ± 2%

name                             speed
HTTPHandler/32/counter-12        1.02GB/s ± 0%
HTTPHandler/32/gauge-12           577MB/s ± 0%
HTTPHandler/32/histogram5-12      808MB/s ± 1%
HTTPHandler/32/label5-12          665MB/s ± 0%
HTTPHandler/32/label2x3x5-12      837MB/s ± 1%
HTTPHandler/32/sample-12          471MB/s ± 0%
HTTPHandler/1024/counter-12      1.61GB/s ± 2%
HTTPHandler/1024/gauge-12         821MB/s ± 1%
HTTPHandler/1024/histogram5-12    910MB/s ± 1%
HTTPHandler/1024/label5-12        880MB/s ± 2%
HTTPHandler/1024/label2x3x5-12   1.12GB/s ± 2%
HTTPHandler/1024/sample-12        627MB/s ± 1%
HTTPHandler/32768/counter-12     1.51GB/s ± 1%
HTTPHandler/32768/gauge-12        805MB/s ± 1%
HTTPHandler/32768/histogram5-12   859MB/s ± 1%
HTTPHandler/32768/label5-12       857MB/s ± 0%
HTTPHandler/32768/label2x3x5-12  1.03GB/s ± 1%
HTTPHandler/32768/sample-12       625MB/s ± 2%
```
