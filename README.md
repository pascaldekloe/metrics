[![API Documentation](https://godoc.org/github.com/pascaldekloe/metrics?status.svg)](https://godoc.org/github.com/pascaldekloe/metrics)

A Prometheus exposition library for the Go programming language.

This is free and unencumbered software released into the
[public domain](https://creativecommons.org/publicdomain/zero/1.0).


### Performance on a Mac Pro (late 2013)

```
name                             time/op
WithLabels/sequential/4-12         23.1ns ± 2%
WithLabels/sequential/4x4-12       39.5ns ± 1%
WithLabels/sequential/4x4x4-12     96.2ns ± 1%
WithLabels/parallel/4-12           92.8ns ± 0%
WithLabels/parallel/4x4-12          157ns ± 1%
WithLabels/parallel/4x4x4-12        290ns ± 0%
Get/counter/sequential-12          0.26ns ± 3%
Get/counter/2routines-12           0.55ns ± 3%
Get/gauge/sequential-12            0.52ns ± 0%
Get/gauge/2routines-12             0.19ns ± 0%
Get/sample/sequential-12           1.04ns ± 0%
Get/sample/2routines-12            1.54ns ± 2%
Set/gauge/sequential-12            5.70ns ± 0%
Set/gauge/2routines-12             19.4ns ± 1%
Set/sample/sequential-12           33.7ns ± 2%
Set/sample/2routines-12            35.9ns ± 8%
Add/counter/sequential-12          5.69ns ± 0%
Add/counter/2routines-12           20.1ns ± 0%
Add/gauge/sequential-12            10.1ns ± 0%
Add/gauge/2routines-12             17.9ns ± 0%
Add/histogram5/sequential-12       28.8ns ± 0%
Add/histogram5/2routines-12        81.3ns ± 3%
HTTPHandler/32/counter-12          2.18µs ± 1%
HTTPHandler/32/gauge-12            3.63µs ± 0%
HTTPHandler/32/histogram5-12       14.4µs ± 1%
HTTPHandler/32/label5-12           3.83µs ± 0%
HTTPHandler/32/label2x3x5-12       5.46µs ± 1%
HTTPHandler/32/sample-12           4.75µs ± 0%
HTTPHandler/1024/counter-12        44.8µs ± 2%
HTTPHandler/1024/gauge-12          86.3µs ± 1%
HTTPHandler/1024/histogram5-12      420µs ± 2%
HTTPHandler/1024/label5-12         93.7µs ± 0%
HTTPHandler/1024/label2x3x5-12      144µs ± 1%
HTTPHandler/1024/sample-12          119µs ± 2%
HTTPHandler/32768/counter-12       1.41ms ± 1%
HTTPHandler/32768/gauge-12         2.86ms ± 2%
HTTPHandler/32768/histogram5-12    14.7ms ± 2%
HTTPHandler/32768/label5-12        3.13ms ± 2%
HTTPHandler/32768/label2x3x5-12    5.83ms ± 0%
HTTPHandler/32768/sample-12        3.88ms ± 0%

name                             speed
HTTPHandler/32/counter-12        1.10GB/s ± 1%
HTTPHandler/32/gauge-12           589MB/s ± 0%
HTTPHandler/32/histogram5-12      816MB/s ± 1%
HTTPHandler/32/label5-12          750MB/s ± 0%
HTTPHandler/32/label2x3x5-12      661MB/s ± 1%
HTTPHandler/32/sample-12          477MB/s ± 0%
HTTPHandler/1024/counter-12      1.78GB/s ± 2%
HTTPHandler/1024/gauge-12         828MB/s ± 1%
HTTPHandler/1024/histogram5-12    919MB/s ± 2%
HTTPHandler/1024/label5-12       1.01GB/s ± 0%
HTTPHandler/1024/label2x3x5-12    824MB/s ± 1%
HTTPHandler/1024/sample-12        633MB/s ± 2%
HTTPHandler/32768/counter-12     1.92GB/s ± 1%
HTTPHandler/32768/gauge-12        859MB/s ± 1%
HTTPHandler/32768/histogram5-12   872MB/s ± 2%
HTTPHandler/32768/label5-12      1.02GB/s ± 2%
HTTPHandler/32768/label2x3x5-12   680MB/s ± 0%
HTTPHandler/32768/sample-12       667MB/s ± 0%
```
