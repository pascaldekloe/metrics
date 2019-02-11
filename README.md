[![API Documentation](https://godoc.org/github.com/pascaldekloe/metrics?status.svg)](https://godoc.org/github.com/pascaldekloe/metrics)

A Prometheus exposition library for the Go programming language.

This is free and unencumbered software released into the
[public domain](https://creativecommons.org/publicdomain/zero/1.0).


### Performance on a Mac Pro (late 2013)

```
name                             time/op
WithLabels/sequential/4-12       23.2ns ± 3%
WithLabels/sequential/4x4-12     39.3ns ± 1%
WithLabels/sequential/4x4x4-12    117ns ± 1%
WithLabels/parallel/4-12         93.0ns ± 0%
WithLabels/parallel/4x4-12        152ns ± 0%
WithLabels/parallel/4x4x4-12      360ns ± 0%
Get/counter/sequential-12        0.26ns ± 0%
Get/counter/2routines-12         0.19ns ± 0%
Get/gauge/sequential-12          0.52ns ± 0%
Get/gauge/2routines-12           0.19ns ± 0%
Get/sample/sequential-12         1.03ns ± 0%
Get/sample/2routines-12          0.28ns ± 0%
Set/gauge/sequential-12          5.69ns ± 0%
Set/gauge/2routines-12           22.9ns ± 0%
Set/sample/sequential-12         34.3ns ± 0%
Set/sample/2routines-12          36.4ns ± 6%
Add/counter/sequential-12        5.70ns ± 0%
Add/counter/2routines-12         24.0ns ± 3%
Add/gauge/sequential-12          10.1ns ± 0%
Add/gauge/2routines-12            153ns ± 0%
Add/histogram5/sequential-12     29.0ns ± 1%
Add/histogram5/2routines-12       159ns ± 0%
HTTPHandler/32/counter-12        1.96µs ± 0%
HTTPHandler/32/gauge-12          3.47µs ± 0%
HTTPHandler/32/histogram5-12     14.1µs ± 1%
HTTPHandler/32/label5-12         3.69µs ± 1%
HTTPHandler/32/label2x3x5-12     3.68µs ± 1%
HTTPHandler/32/sample-12         4.64µs ± 1%
HTTPHandler/1024/counter-12      43.8µs ± 4%
HTTPHandler/1024/gauge-12        88.1µs ± 1%
HTTPHandler/1024/histogram5-12    419µs ± 2%
HTTPHandler/1024/label5-12       97.7µs ± 3%
HTTPHandler/1024/label2x3x5-12   96.5µs ± 1%
HTTPHandler/1024/sample-12        121µs ± 1%
HTTPHandler/32768/counter-12     1.42ms ± 4%
HTTPHandler/32768/gauge-12       2.88ms ± 0%
HTTPHandler/32768/histogram5-12  15.6ms ± 2%
HTTPHandler/32768/label5-12      3.19ms ± 1%
HTTPHandler/32768/label2x3x5-12  3.28ms ± 2%
HTTPHandler/32768/sample-12      4.01ms ± 1%
```
