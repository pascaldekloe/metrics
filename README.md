[![API Documentation](https://godoc.org/github.com/pascaldekloe/metrics?status.svg)](https://godoc.org/github.com/pascaldekloe/metrics)
[![Build Status](https://travis-ci.org/pascaldekloe/metrics.svg?branch=master)](https://travis-ci.org/pascaldekloe/metrics)

## About

Atomic measures with Prometheus exposition for the Go programming language.

This is free and unencumbered software released into the
[public domain](https://creativecommons.org/publicdomain/zero/1.0).


## Use

Metric regisration on package level comes recommened. The declarations help to
document the provided funcionality too.

```go
var (
	ConnectCount = metrics.MustCounter("db_connects_total", "Number of established initiations.")
	CacheBytes   = metrics.MustInteger("db_cache_bytes", "Size of collective responses.")
	DiskUsage    = metrics.Must1LabelRealSample("db_disk_usage_ratio", "device")
)
```

Updates are error free by design, e.g., `CacheBytes.Add(-72)` or
`DiskUsage(dev.Name).Set(1 - dev.Free, time.Now())`.

Serve HTTP with just `http.HandleFunc("/metrics", metrics.ServeHTTP)`.

```
< HTTP/1.1 200 OK
< Content-Type: text/plain;version=0.0.4;charset=utf-8
< Date: Sun, 07 Mar 2021 15:22:47 GMT
< Content-Length: 351
< 
# Prometheus Samples

# TYPE db_connects_total counter
# HELP db_connects_total Number of established initiations.
db_connects_total 4 1615130567389

# TYPE db_cache_bytes gauge
# HELP db_cache_bytes Size of collective responses.
db_cache_bytes 7600 1615130567389

# TYPE db_disk_usage_ratio gauge
db_disk_usage_ratio{device="sda"} 0.19 1615130563595
```

Package `github.com/pascaldekloe/metrics/gostat` provides a defacto standard
collection of Go metrics, similar to the setup from the
[original Prometheus library](https://github.com/prometheus/client_golang).
See the
[lazy example](https://pkg.go.dev/github.com/pascaldekloe/metrics#example-Sample-Lazy)
for detail on capturing.


## Performance

The following benchmarks were measured on a 3.5 GHz Xeon from the year 2013.

```
name                           time/op
Label/sequential/4-12            20.6ns ± 0%
Label/sequential/4x4-12          25.1ns ± 0%
Label/sequential/4x4x4-12        46.2ns ± 1%
Label/parallel/4-12               100ns ± 0%
Label/parallel/4x4-12             119ns ± 0%
Label/parallel/4x4x4-12           155ns ± 0%
Get/histogram5/sequential-12      105ns ± 0%
Get/histogram5/2routines-12       140ns ± 0%
Set/real/sequential-12           5.71ns ± 0%
Set/real/2routines-12            14.5ns ±32%
Set/sample/sequential-12         33.0ns ± 0%
Set/sample/2routines-12          42.4ns ± 1%
Add/counter/sequential-12        5.71ns ± 0%
Add/counter/2routines-12         16.2ns ±22%
Add/integer/sequential-12        5.70ns ± 0%
Add/integer/2routines-12         16.7ns ±19%
Add/histogram5/sequential-12     29.7ns ± 0%
Add/histogram5/2routines-12      71.5ns ±11%
ServeHTTP/32/counter-12          2.08µs ± 0%
ServeHTTP/32/real-12             3.42µs ± 0%
ServeHTTP/32/integer-12          2.09µs ± 0%
ServeHTTP/32/histogram5-12       11.9µs ± 0%
ServeHTTP/32/label5-12           3.92µs ± 1%
ServeHTTP/32/label2x3x5-12       3.98µs ± 1%
ServeHTTP/32/sample-12           3.45µs ± 0%
ServeHTTP/1024/counter-12        41.2µs ± 1%
ServeHTTP/1024/real-12           81.1µs ± 3%
ServeHTTP/1024/integer-12        43.0µs ± 2%
ServeHTTP/1024/histogram5-12      342µs ± 1%
ServeHTTP/1024/label5-12         97.7µs ± 4%
ServeHTTP/1024/label2x3x5-12     98.2µs ± 3%
ServeHTTP/1024/sample-12         82.4µs ± 1%
ServeHTTP/32768/counter-12       1.36ms ± 3%
ServeHTTP/32768/real-12          2.63ms ± 4%
ServeHTTP/32768/integer-12       1.36ms ± 2%
ServeHTTP/32768/histogram5-12    12.1ms ± 1%
ServeHTTP/32768/label5-12        3.35ms ± 1%
ServeHTTP/32768/label2x3x5-12    3.44ms ± 1%
ServeHTTP/32768/sample-12        2.73ms ± 1%

name                           speed
ServeHTTP/32/counter-12         934MB/s ± 0%
ServeHTTP/32/real-12            494MB/s ± 0%
ServeHTTP/32/integer-12         900MB/s ± 0%
ServeHTTP/32/histogram5-12      728MB/s ± 0%
ServeHTTP/32/label5-12          619MB/s ± 1%
ServeHTTP/32/label2x3x5-12      795MB/s ± 1%
ServeHTTP/32/sample-12          489MB/s ± 0%
ServeHTTP/1024/counter-12      1.59GB/s ± 1%
ServeHTTP/1024/real-12          705MB/s ± 3%
ServeHTTP/1024/integer-12      1.47GB/s ± 2%
ServeHTTP/1024/histogram5-12    836MB/s ± 1%
ServeHTTP/1024/label5-12        826MB/s ± 4%
ServeHTTP/1024/label2x3x5-12   1.06GB/s ± 3%
ServeHTTP/1024/sample-12        693MB/s ± 1%
ServeHTTP/32768/counter-12     1.67GB/s ± 3%
ServeHTTP/32768/real-12         760MB/s ± 3%
ServeHTTP/32768/integer-12     1.61GB/s ± 2%
ServeHTTP/32768/histogram5-12   793MB/s ± 1%
ServeHTTP/32768/label5-12       821MB/s ± 1%
ServeHTTP/32768/label2x3x5-12  1.02GB/s ± 1%
ServeHTTP/32768/sample-12       732MB/s ± 1%
```
