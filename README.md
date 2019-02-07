[![API Documentation](https://godoc.org/github.com/pascaldekloe/metrics?status.svg)](https://godoc.org/github.com/pascaldekloe/metrics)

A Prometheus exposition library for the Go programming language.

This is free and unencumbered software released into the
[public domain](https://creativecommons.org/publicdomain/zero/1.0).


### Performance on a Mac Pro (late 2013)

```
name                             time/op
PlaceLabels/sequential/4-12      23.0ns ± 0%
PlaceLabels/sequential/4x4-12    39.5ns ± 1%
PlaceLabels/sequential/4x4x4-12   143ns ± 1%
PlaceLabels/parallel/4-12        91.8ns ± 0%
PlaceLabels/parallel/4x4-12       152ns ± 0%
PlaceLabels/parallel/4x4x4-12     387ns ± 0%
Get/counter/sequential-12        0.26ns ± 0%
Get/counter/2routines-12         0.20ns ± 0%
Get/gauge/sequential-12          0.26ns ± 0%
Get/gauge/2routines-12           0.19ns ± 0%
Set/gauge/sequential-12          5.71ns ± 0%
Set/gauge/2routines-12           20.2ns ± 1%
Add/counter/sequential-12        5.72ns ± 0%
Add/counter/2routines-12         22.8ns ± 1%
Add/gauge/sequential-12          10.1ns ± 0%
Add/gauge/2routines-12            128ns ± 1%
Add/histogram5/sequential-12     30.6ns ± 0%
Add/histogram5/2routines-12       156ns ± 0%
HTTPHandler/32/counter-12        1.99µs ± 0%
HTTPHandler/32/gauge-12          3.51µs ± 0%
HTTPHandler/32/histogram5-12     17.7µs ± 0%
HTTPHandler/32/label5-12         3.79µs ± 0%
HTTPHandler/32/label2x3x5-12     3.70µs ± 0%
HTTPHandler/1024/counter-12      43.7µs ± 1%
HTTPHandler/1024/gauge-12        88.4µs ± 1%
HTTPHandler/1024/histogram5-12    532µs ± 1%
HTTPHandler/1024/label5-12       96.3µs ± 1%
HTTPHandler/1024/label2x3x5-12   96.9µs ± 1%
HTTPHandler/32768/counter-12     1.41ms ± 3%
HTTPHandler/32768/gauge-12       2.87ms ± 1%
HTTPHandler/32768/histogram5-12  18.3ms ± 1%
HTTPHandler/32768/label5-12      3.21ms ± 1%
HTTPHandler/32768/label2x3x5-12  3.29ms ± 0%
```
