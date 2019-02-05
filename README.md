[![API Documentation](https://godoc.org/github.com/pascaldekloe/metrics?status.svg)](https://godoc.org/github.com/pascaldekloe/metrics)

A Prometheus exposition library for the Go programming language.

This is free and unencumbered software released into the
[public domain](https://creativecommons.org/publicdomain/zero/1.0).


### Performance on a Mac Pro (late 2013)

```
name                             time/op
ForLabel/sequential/4-12         23.0ns ± 0%
ForLabel/sequential/4x4-12       40.0ns ± 1%
ForLabel/sequential/4x4x4-12      142ns ± 1%
ForLabel/parallel/4-12           91.2ns ± 0%
ForLabel/parallel/4x4-12          155ns ± 0%
ForLabel/parallel/4x4x4-12        390ns ± 1%
Get/counter/sequential-12        0.26ns ± 0%
Get/counter/2routines-12         0.20ns ± 0%
Get/gauge/sequential-12          0.26ns ± 0%
Get/gauge/2routines-12           0.19ns ± 0%
Set/gauge/sequential-12          5.71ns ± 0%
Set/gauge/2routines-12           22.2ns ± 0%
Add/counter/sequential-12        5.71ns ± 0%
Add/counter/2routines-12         26.3ns ± 1%
Add/gauge/sequential-12          10.1ns ± 0%
Add/gauge/2routines-12            129ns ± 0%
HTTPHandler/32/counter-12        1.98µs ± 1%
HTTPHandler/32/gauge-12          3.56µs ± 0%
HTTPHandler/32/label5-12         3.71µs ± 1%
HTTPHandler/32/label2x3x5-12     3.79µs ± 0%
HTTPHandler/1024/counter-12      43.0µs ± 2%
HTTPHandler/1024/gauge-12        88.1µs ± 2%
HTTPHandler/1024/label5-12       94.7µs ± 1%
HTTPHandler/1024/label2x3x5-12   98.3µs ± 2%
HTTPHandler/32768/counter-12     1.40ms ± 1%
HTTPHandler/32768/gauge-12       2.90ms ± 2%
HTTPHandler/32768/label5-12      3.09ms ± 1%
HTTPHandler/32768/label2x3x5-12  3.27ms ± 0%
```
