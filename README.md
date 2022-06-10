[![Go Reference](https://pkg.go.dev/badge/github.com/pascaldekloe/metrics.svg)](https://pkg.go.dev/github.com/pascaldekloe/metrics)
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

The following benchmarks were measured on a Intel(R) Core(TM) i5-7500 CPU @ 3.40GHz.

```
name                          time/op
Label/sequential/4-4            22.2ns ± 1%
Label/sequential/4x4-4          25.9ns ± 1%
Label/sequential/4x4x4-4        42.2ns ± 1%
Label/parallel/4-4              45.9ns ± 0%
Label/parallel/4x4-4            50.8ns ± 1%
Label/parallel/4x4x4-4          71.7ns ± 1%
Get/histogram5/sequential-4     89.1ns ± 0%
Get/histogram5/2routines-4       112ns ± 2%
Set/real/sequential-4           6.74ns ± 0%
Set/real/2routines-4            14.8ns ± 4%
Set/sample/sequential-4         18.6ns ± 6%
Set/sample/2routines-4          22.2ns ± 1%
Add/counter/sequential-4        6.75ns ± 0%
Add/counter/2routines-4         21.2ns ± 1%
Add/integer/sequential-4        6.75ns ± 0%
Add/integer/2routines-4         21.2ns ± 1%
Add/histogram5/sequential-4     37.9ns ± 0%
Add/histogram5/2routines-4      85.9ns ± 1%
ServeHTTP/32/counter-4          1.11µs ± 0%
ServeHTTP/32/real-4             2.24µs ± 1%
ServeHTTP/32/integer-4          1.13µs ± 1%
ServeHTTP/32/histogram5-4       9.85µs ± 0%
ServeHTTP/32/label5-4           2.73µs ± 1%
ServeHTTP/32/label2x3x5-4       2.79µs ± 2%
ServeHTTP/32/sample-4           2.70µs ± 1%
ServeHTTP/1024/counter-4        32.1µs ± 1%
ServeHTTP/1024/real-4           62.5µs ± 2%
ServeHTTP/1024/integer-4        32.6µs ± 1%
ServeHTTP/1024/histogram5-4      303µs ± 1%
ServeHTTP/1024/label5-4         76.3µs ± 4%
ServeHTTP/1024/label2x3x5-4     75.8µs ± 3%
ServeHTTP/1024/sample-4         77.7µs ± 3%
ServeHTTP/32768/counter-4       1.18ms ± 1%
ServeHTTP/32768/real-4          2.15ms ± 1%
ServeHTTP/32768/integer-4       1.20ms ± 4%
ServeHTTP/32768/histogram5-4    13.9ms ± 2%
ServeHTTP/32768/label5-4        2.99ms ± 1%
ServeHTTP/32768/label2x3x5-4    2.90ms ± 4%
ServeHTTP/32768/sample-4        2.70ms ± 6%

name                          alloc/op
Label/sequential/4-4             0.00B     
Label/sequential/4x4-4           0.00B     
Label/sequential/4x4x4-4         0.00B     
Label/parallel/4-4               0.00B     
Label/parallel/4x4-4             0.00B     
Label/parallel/4x4x4-4           0.00B     
Get/histogram5/sequential-4      0.00B     
Get/histogram5/2routines-4       0.00B     
Set/real/sequential-4            0.00B     
Set/real/2routines-4             0.00B     
Set/sample/sequential-4          0.00B     
Set/sample/2routines-4           0.00B     
Add/counter/sequential-4         0.00B     
Add/counter/2routines-4          0.00B     
Add/integer/sequential-4         0.00B     
Add/integer/2routines-4          0.00B     
Add/histogram5/sequential-4      0.00B     
Add/histogram5/2routines-4       0.00B     
ServeHTTP/32/counter-4            560B ± 0%
ServeHTTP/32/real-4               512B ± 0%
ServeHTTP/32/integer-4            560B ± 0%
ServeHTTP/32/histogram5-4       1.19kB ± 0%
ServeHTTP/32/label5-4             560B ± 0%
ServeHTTP/32/label2x3x5-4         752B ± 0%
ServeHTTP/32/sample-4             512B ± 0%
ServeHTTP/1024/counter-4          560B ± 0%
ServeHTTP/1024/real-4             560B ± 0%
ServeHTTP/1024/integer-4          560B ± 0%
ServeHTTP/1024/histogram5-4     1.19kB ± 0%
ServeHTTP/1024/label5-4           560B ± 0%
ServeHTTP/1024/label2x3x5-4       576B ± 0%
ServeHTTP/1024/sample-4           560B ± 0%
ServeHTTP/32768/counter-4         565B ± 0%
ServeHTTP/32768/real-4            569B ± 0%
ServeHTTP/32768/integer-4         565B ± 0%
ServeHTTP/32768/histogram5-4    1.26kB ± 0%
ServeHTTP/32768/label5-4          573B ± 0%
ServeHTTP/32768/label2x3x5-4      588B ± 0%
ServeHTTP/32768/sample-4          571B ± 0%

name                          allocs/op
Label/sequential/4-4              0.00     
Label/sequential/4x4-4            0.00     
Label/sequential/4x4x4-4          0.00     
Label/parallel/4-4                0.00     
Label/parallel/4x4-4              0.00     
Label/parallel/4x4x4-4            0.00     
Get/histogram5/sequential-4       0.00     
Get/histogram5/2routines-4        0.00     
Set/real/sequential-4             0.00     
Set/real/2routines-4              0.00     
Set/sample/sequential-4           0.00     
Set/sample/2routines-4            0.00     
Add/counter/sequential-4          0.00     
Add/counter/2routines-4           0.00     
Add/integer/sequential-4          0.00     
Add/integer/2routines-4           0.00     
Add/histogram5/sequential-4       0.00     
Add/histogram5/2routines-4        0.00     
ServeHTTP/32/counter-4            5.00 ± 0%
ServeHTTP/32/real-4               5.00 ± 0%
ServeHTTP/32/integer-4            5.00 ± 0%
ServeHTTP/32/histogram5-4         10.0 ± 0%
ServeHTTP/32/label5-4             5.00 ± 0%
ServeHTTP/32/label2x3x5-4         6.00 ± 0%
ServeHTTP/32/sample-4             5.00 ± 0%
ServeHTTP/1024/counter-4          5.00 ± 0%
ServeHTTP/1024/real-4             5.00 ± 0%
ServeHTTP/1024/integer-4          5.00 ± 0%
ServeHTTP/1024/histogram5-4       10.0 ± 0%
ServeHTTP/1024/label5-4           5.00 ± 0%
ServeHTTP/1024/label2x3x5-4       5.00 ± 0%
ServeHTTP/1024/sample-4           5.00 ± 0%
ServeHTTP/32768/counter-4         5.00 ± 0%
ServeHTTP/32768/real-4            5.00 ± 0%
ServeHTTP/32768/integer-4         5.00 ± 0%
ServeHTTP/32768/histogram5-4      10.0 ± 0%
ServeHTTP/32768/label5-4          5.00 ± 0%
ServeHTTP/32768/label2x3x5-4      5.00 ± 0%
ServeHTTP/32768/sample-4          5.00 ± 0%

name                          speed
ServeHTTP/32/counter-4        1.75GB/s ± 0%
ServeHTTP/32/real-4            754MB/s ± 1%
ServeHTTP/32/integer-4        1.67GB/s ± 1%
ServeHTTP/32/histogram5-4      878MB/s ± 0%
ServeHTTP/32/label5-4          889MB/s ± 1%
ServeHTTP/32/label2x3x5-4     1.14GB/s ± 2%
ServeHTTP/32/sample-4          625MB/s ± 1%
ServeHTTP/1024/counter-4      2.03GB/s ± 1%
ServeHTTP/1024/real-4          914MB/s ± 2%
ServeHTTP/1024/integer-4      1.94GB/s ± 1%
ServeHTTP/1024/histogram5-4    946MB/s ± 1%
ServeHTTP/1024/label5-4       1.06GB/s ± 4%
ServeHTTP/1024/label2x3x5-4   1.38GB/s ± 3%
ServeHTTP/1024/sample-4        735MB/s ± 3%
ServeHTTP/32768/counter-4     1.92GB/s ± 1%
ServeHTTP/32768/real-4         929MB/s ± 1%
ServeHTTP/32768/integer-4     1.83GB/s ± 4%
ServeHTTP/32768/histogram5-4   690MB/s ± 2%
ServeHTTP/32768/label5-4       921MB/s ± 1%
ServeHTTP/32768/label2x3x5-4  1.21GB/s ± 4%
ServeHTTP/32768/sample-4       742MB/s ± 5%
```
