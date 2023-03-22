[![API](https://pkg.go.dev/badge/github.com/pascaldekloe/metrics.svg)](https://pkg.go.dev/github.com/pascaldekloe/metrics)
[![Build](https://github.com/pascaldekloe/metrics/actions/workflows/go.yml/badge.svg)](https://github.com/pascaldekloe/metrics/actions/workflows/go.yml)

## About

Metrics are measures of quantitative assessment commonly used for comparing, and
tracking performance or production. This library offers atomic counters, gauges
and historgrams for the Go programming language. Users have the option to expose
snapshots in the Prometheus text-format.

This is free and unencumbered software released into the
[public domain](https://creativecommons.org/publicdomain/zero/1.0).


## Use

Static regisration on package level comes recommened. The declarations also help
to document the funcionality that is covered in the code.

```go
// Package Metrics
var (
	ConnectCount = metrics.MustCounter("db_connects_total", "Number of established initiations.")
	CacheBytes   = metrics.MustInteger("db_cache_bytes", "Size of collective responses.")
	DiskUsage    = metrics.Must1LabelRealSample("db_disk_usage_ratio", "device")
)
```

Update methods operate error free by design, e.g., `CacheBytes.Add(-72)` or
`DiskUsage(dev.Name).Set(1 - dev.Free, time.Now())`.

Serve HTTP with just `http.HandleFunc("/metrics", metrics.ServeHTTP)`.

```
< HTTP/1.1 200 OK
< Content-Type: text/plain;version=0.0.4
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

Package `github.com/pascaldekloe/metrics/gostat` provides a standard collection
of Go metrics which is similar to the setup as provided by the
[original Prometheus library](https://github.com/prometheus/client_golang).

Samples may be fetched in a lazy manner, like how the
[lazy example](https://pkg.go.dev/github.com/pascaldekloe/metrics#example-Sample-Lazy)
does.


## Performance

The following results were measured on an Apple M1 with Go 1.20.

```
name                          time/op
Label/sequential/4-8            14.4ns ± 0%
Label/sequential/4x4-8          17.1ns ± 0%
Label/sequential/4x4x4-8        29.6ns ± 0%
Label/parallel/4-8              85.3ns ± 2%
Label/parallel/4x4-8            89.2ns ± 1%
Label/parallel/4x4x4-8           103ns ± 0%
Get/histogram5/sequential-8     45.0ns ± 0%
Get/histogram5/2routines-8      85.1ns ± 0%
Set/real/sequential-8           5.64ns ± 0%
Set/real/2routines-8            16.5ns ± 2%
Set/sample/sequential-8         13.6ns ± 1%
Set/sample/2routines-8          38.7ns ± 7%
Add/counter/sequential-8        6.88ns ± 0%
Add/counter/2routines-8         16.1ns ± 2%
Add/integer/sequential-8        6.88ns ± 0%
Add/integer/2routines-8         16.1ns ± 1%
Add/histogram5/sequential-8     16.1ns ± 1%
Add/histogram5/2routines-8      69.5ns ± 1%
ServeHTTP/32/counter-8           687ns ± 0%
ServeHTTP/32/real-8             1.87µs ± 0%
ServeHTTP/32/integer-8           694ns ± 0%
ServeHTTP/32/histogram5-8       6.05µs ± 0%
ServeHTTP/32/label5-8           1.97µs ± 0%
ServeHTTP/32/label2x3x5-8       1.98µs ± 0%
ServeHTTP/32/sample-8           2.06µs ± 0%
ServeHTTP/1024/counter-8        18.5µs ± 0%
ServeHTTP/1024/real-8           50.9µs ± 0%
ServeHTTP/1024/integer-8        18.8µs ± 0%
ServeHTTP/1024/histogram5-8      192µs ± 0%
ServeHTTP/1024/label5-8         54.4µs ± 0%
ServeHTTP/1024/label2x3x5-8     54.4µs ± 0%
ServeHTTP/1024/sample-8         57.6µs ± 0%
```
