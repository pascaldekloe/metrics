[![API Documentation](https://godoc.org/github.com/pascaldekloe/metrics?status.svg)](https://godoc.org/github.com/pascaldekloe/metrics)

A Prometheus exposition library for the Go programming language.

```go
var ConnCount = metrics.MustCounter("db_connects_total", "Number of established initiations.")

func main() {
	// include default metrics
	gostat.CaptureEvery(time.Minute)
	// mount exposition point
	http.HandleFunc("/metrics", metrics.ServeHTTP)

	// …
}
```

This is free and unencumbered software released into the
[public domain](https://creativecommons.org/publicdomain/zero/1.0).


### Performance

The following results were measured on a 3.5 GHz Xeon E5-1650 v2 (Ivy Bridge-EP).

```
name                          time/op
Label/sequential/4-12           20.6ns ± 0%
Label/sequential/4x4-12         25.1ns ± 0%
Label/sequential/4x4x4-12       46.4ns ± 1%
Label/parallel/4-12              100ns ± 0%
Label/parallel/4x4-12            120ns ± 1%
Label/parallel/4x4x4-12          158ns ± 0%
Get/histogram5/sequential-12     105ns ± 1%
Get/histogram5/2routines-12      139ns ± 1%
Set/real/sequential-12          5.72ns ± 0%
Set/real/2routines-12           19.2ns ± 9%
Set/sample/sequential-12        33.1ns ± 1%
Set/sample/2routines-12         42.8ns ± 1%
Add/counter/sequential-12       5.73ns ± 1%
Add/counter/2routines-12        19.9ns ±19%
Add/integer/sequential-12       5.72ns ± 0%
Add/integer/2routines-12        21.6ns ± 9%
Add/histogram5/sequential-12    29.8ns ± 1%
Add/histogram5/2routines-12     58.9ns ±14%
ServeHTTP/32/counter-12         2.10µs ± 0%
ServeHTTP/32/real-12            3.44µs ± 0%
ServeHTTP/32/integer-12         2.11µs ± 1%
ServeHTTP/32/histogram5-12      11.9µs ± 0%
ServeHTTP/32/label5-12          3.96µs ± 1%
ServeHTTP/32/label2x3x5-12      3.94µs ± 1%
ServeHTTP/32/sample-12          3.47µs ± 0%
ServeHTTP/1024/counter-12       41.3µs ± 1%
ServeHTTP/1024/real-12          81.8µs ± 2%
ServeHTTP/1024/integer-12       42.7µs ± 0%

name                          speed
ServeHTTP/32/counter-12        927MB/s ± 0%
ServeHTTP/32/real-12           492MB/s ± 0%
ServeHTTP/32/integer-12        891MB/s ± 1%
ServeHTTP/32/histogram5-12     725MB/s ± 0%
ServeHTTP/32/label5-12         613MB/s ± 1%
ServeHTTP/32/label2x3x5-12     802MB/s ± 1%
ServeHTTP/32/sample-12         487MB/s ± 1%
ServeHTTP/1024/counter-12     1.58GB/s ± 1%
ServeHTTP/1024/real-12         698MB/s ± 2%
ServeHTTP/1024/integer-12     1.48GB/s ± 0%
```
