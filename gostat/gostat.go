// Package gostat provides Go runtime statistics.
// The metrics may be applied with something like the following:
//
//	go func() {
//		for range time.Tick(time.Minute) {
//			gostat.Capture()
//		}
//	}
//
// The bindings are the same as the standard client's implementation. See
// the prometheus.NewGoCollector function documentation for details.
package gostat

import (
	"runtime"
	"time"

	"github.com/pascaldekloe/metrics"
)

var (
	NumGoroutine = metrics.MustNewGaugeSample("go_goroutines").Help("Number of goroutines that currently exist.")
	ThreadCreate = metrics.MustNewGaugeSample("go_threads").Help("Number of OS threads created.")

	// BUG(pascaldekloe): go_gc_duration_seconds not implemented

	// GCPause = metrics.MustNewSummarySample("go_gc_duration_seconds").Help("A summary of the GC invocation durations.")

	GoInfo = metrics.Must1LabelGauge("go_info", "version").With(runtime.Version()).Help("Information about the Go environment.")
)

// Runtime MemStats Copies
var (
	Alloc         = metrics.MustNewGaugeSample("go_memstats_alloc_bytes").Help("Number of bytes allocated and still in use.")
	TotalAlloc    = metrics.MustNewCounterSample("go_memstats_alloc_bytes_total").Help("Total number of bytes allocated, even if freed.")
	Sys           = metrics.MustNewGaugeSample("go_memstats_sys_bytes").Help("Number of bytes obtained from system.")
	Lookups       = metrics.MustNewCounterSample("go_memstats_lookups_total").Help("Total number of pointer lookups.")
	Mallocs       = metrics.MustNewCounterSample("go_memstats_mallocs_total").Help("Total number of mallocs.")
	Frees         = metrics.MustNewCounterSample("go_memstats_frees_total").Help("Total number of frees.")
	HeapAlloc     = metrics.MustNewGaugeSample("go_memstats_heap_alloc_bytes").Help("Number of heap bytes allocated and still in use.")
	HeapSys       = metrics.MustNewGaugeSample("go_memstats_heap_sys_bytes").Help("Number of heap bytes obtained from system.")
	HeapIdle      = metrics.MustNewGaugeSample("go_memstats_heap_idle_bytes").Help("Number of heap bytes waiting to be used.")
	HeapInuse     = metrics.MustNewGaugeSample("go_memstats_heap_inuse_bytes").Help("Number of heap bytes that are in use.")
	HeapReleased  = metrics.MustNewGaugeSample("go_memstats_heap_released_bytes").Help("Number of heap bytes released to OS.")
	HeapObjects   = metrics.MustNewGaugeSample("go_memstats_heap_objects").Help("Number of allocated objects.")
	StackInuse    = metrics.MustNewGaugeSample("go_memstats_stack_inuse_bytes").Help("Number of bytes in use by the stack allocator.")
	StackSys      = metrics.MustNewGaugeSample("go_memstats_stack_sys_bytes").Help("Number of bytes obtained from system for stack allocator.")
	MSpanInuse    = metrics.MustNewGaugeSample("go_memstats_mspan_inuse_bytes").Help("Number of bytes in use by mspan structures.")
	MSpanSys      = metrics.MustNewGaugeSample("go_memstats_mspan_sys_bytes").Help("Number of bytes used for mspan structures obtained from system.")
	MCacheInuse   = metrics.MustNewGaugeSample("go_memstats_mcache_inuse_bytes").Help("Number of bytes in use by mcache structures.")
	MCacheSys     = metrics.MustNewGaugeSample("go_memstats_mcache_sys_bytes").Help("Number of bytes used for mcache structures obtained from system.")
	BuckHashSys   = metrics.MustNewGaugeSample("go_memstats_buck_hash_sys_bytes").Help("Number of bytes used by the profiling bucket hash table.")
	GCSys         = metrics.MustNewGaugeSample("go_memstats_gc_sys_bytes").Help("Number of bytes used for garbage collection system metadata.")
	OtherSys      = metrics.MustNewGaugeSample("go_memstats_other_sys_bytes").Help("Number of bytes used for other system allocations.")
	NextGC        = metrics.MustNewGaugeSample("go_memstats_next_gc_bytes").Help("Number of heap bytes when next garbage collection will take place.")
	LastGC        = metrics.MustNewGaugeSample("go_memstats_last_gc_time_seconds").Help("Number of seconds since 1970 of last garbage collection.")
	GCCPUFraction = metrics.MustNewGaugeSample("go_memstats_gc_cpu_fraction").Help("The fraction of this program's available CPU time used by the GC since the program started.")
)

// Capture updates the Samples.
func Capture() {
	timestamp := time.Now()
	NumGoroutine.Set(float64(runtime.NumGoroutine()), timestamp)
	recordCount, _ := runtime.ThreadCreateProfile(nil)
	ThreadCreate.Set(float64(recordCount), timestamp)

	stats := new(runtime.MemStats)
	runtime.ReadMemStats(stats)
	timestamp = time.Now()

	Alloc.Set(float64(stats.Alloc), timestamp)
	TotalAlloc.Set(float64(stats.TotalAlloc), timestamp)
	Sys.Set(float64(stats.Sys), timestamp)
	Lookups.Set(float64(stats.Lookups), timestamp)
	Mallocs.Set(float64(stats.Mallocs), timestamp)
	Frees.Set(float64(stats.Frees), timestamp)
	HeapAlloc.Set(float64(stats.HeapAlloc), timestamp)
	HeapSys.Set(float64(stats.HeapSys), timestamp)
	HeapIdle.Set(float64(stats.HeapIdle), timestamp)
	HeapInuse.Set(float64(stats.HeapInuse), timestamp)
	HeapReleased.Set(float64(stats.HeapReleased), timestamp)
	HeapObjects.Set(float64(stats.HeapObjects), timestamp)
	StackInuse.Set(float64(stats.StackInuse), timestamp)
	StackSys.Set(float64(stats.StackSys), timestamp)
	MSpanInuse.Set(float64(stats.MSpanInuse), timestamp)
	MSpanSys.Set(float64(stats.MSpanSys), timestamp)
	MCacheInuse.Set(float64(stats.MCacheInuse), timestamp)
	MCacheSys.Set(float64(stats.MCacheSys), timestamp)
	BuckHashSys.Set(float64(stats.BuckHashSys), timestamp)
	GCSys.Set(float64(stats.GCSys), timestamp)
	OtherSys.Set(float64(stats.OtherSys), timestamp)
	NextGC.Set(float64(stats.NextGC), timestamp)
	LastGC.Set(float64(stats.LastGC)/1e9, timestamp)
	GCCPUFraction.Set(stats.GCCPUFraction, timestamp)
}

func init() {
	// fixed value
	GoInfo.Set(1)
}
