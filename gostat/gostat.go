// Package gostat provides Go statistics to the default registry.
//
// The bindings are equivalent to the standard client implementation.
// See the prometheus.NewGoCollector function documentation for details.
package gostat

import (
	"runtime"
	"time"

	"github.com/pascaldekloe/metrics"
)

func init() {
	metrics.MustHelp("go_info", "Information about the Go environment.")
	GoInfo.Set(1)
}

// GoInfo is set to 1.
var GoInfo = metrics.Must1LabelInteger("go_info", "version")(runtime.Version())

// Runtime Samples
var (
	NumGoroutine = metrics.MustRealSample("go_goroutines", "Number of goroutines that currently exist.")
	ThreadCreate = metrics.MustRealSample("go_threads", "Number of OS threads created.")

	// BUG(pascaldekloe): go_gc_duration_seconds not implemented

	// GCPause = metrics.MustSummarySample("go_gc_duration_seconds", "A summary of the GC invocation durations.")
)

// Memory Allocation Samples
var (
	Alloc         = metrics.MustRealSample("go_memstats_alloc_bytes", "Number of bytes allocated and still in use.")
	TotalAlloc    = metrics.MustCounterSample("go_memstats_alloc_bytes_total", "Total number of bytes allocated, even if freed.")
	Sys           = metrics.MustRealSample("go_memstats_sys_bytes", "Number of bytes obtained from system.")
	Lookups       = metrics.MustCounterSample("go_memstats_lookups_total", "Total number of pointer lookups.")
	Mallocs       = metrics.MustCounterSample("go_memstats_mallocs_total", "Total number of mallocs.")
	Frees         = metrics.MustCounterSample("go_memstats_frees_total", "Total number of frees.")
	HeapAlloc     = metrics.MustRealSample("go_memstats_heap_alloc_bytes", "Number of heap bytes allocated and still in use.")
	HeapSys       = metrics.MustRealSample("go_memstats_heap_sys_bytes", "Number of heap bytes obtained from system.")
	HeapIdle      = metrics.MustRealSample("go_memstats_heap_idle_bytes", "Number of heap bytes waiting to be used.")
	HeapInuse     = metrics.MustRealSample("go_memstats_heap_inuse_bytes", "Number of heap bytes that are in use.")
	HeapReleased  = metrics.MustRealSample("go_memstats_heap_released_bytes", "Number of heap bytes released to OS.")
	HeapObjects   = metrics.MustRealSample("go_memstats_heap_objects", "Number of allocated objects.")
	StackInuse    = metrics.MustRealSample("go_memstats_stack_inuse_bytes", "Number of bytes in use by the stack allocator.")
	StackSys      = metrics.MustRealSample("go_memstats_stack_sys_bytes", "Number of bytes obtained from system for stack allocator.")
	MSpanInuse    = metrics.MustRealSample("go_memstats_mspan_inuse_bytes", "Number of bytes in use by mspan structures.")
	MSpanSys      = metrics.MustRealSample("go_memstats_mspan_sys_bytes", "Number of bytes used for mspan structures obtained from system.")
	MCacheInuse   = metrics.MustRealSample("go_memstats_mcache_inuse_bytes", "Number of bytes in use by mcache structures.")
	MCacheSys     = metrics.MustRealSample("go_memstats_mcache_sys_bytes", "Number of bytes used for mcache structures obtained from system.")
	BuckHashSys   = metrics.MustRealSample("go_memstats_buck_hash_sys_bytes", "Number of bytes used by the profiling bucket hash table.")
	GCSys         = metrics.MustRealSample("go_memstats_gc_sys_bytes", "Number of bytes used for garbage collection system metadata.")
	OtherSys      = metrics.MustRealSample("go_memstats_other_sys_bytes", "Number of bytes used for other system allocations.")
	NextGC        = metrics.MustRealSample("go_memstats_next_gc_bytes", "Number of heap bytes when next garbage collection will take place.")
	LastGC        = metrics.MustRealSample("go_memstats_last_gc_time_seconds", "Number of seconds since 1970 of last garbage collection.")
	GCCPUFraction = metrics.MustRealSample("go_memstats_gc_cpu_fraction", "The fraction of this program's available CPU time used by the GC since the program started.")
)

// Capture updates the samples.
func Capture() {
	NumGoroutine.Set(float64(runtime.NumGoroutine()), time.Now())
	recordCount, _ := runtime.ThreadCreateProfile(nil)
	ThreadCreate.Set(float64(recordCount), time.Now())

	stats := new(runtime.MemStats)
	runtime.ReadMemStats(stats)
	timestamp := time.Now()

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

// CaptureEvery updates the samples with an interval, starting now.
// The routine terminates with a send or close on cancel.
func CaptureEvery(interval time.Duration) (cancel chan<- struct{}) {
	ch := make(chan struct{})

	go func() {
		Capture()

		ticker := time.NewTicker(interval)
		for {
			select {
			case <-ticker.C:
				Capture()

			case <-ch:
				ticker.Stop()
				return
			}
		}
	}()

	return ch
}
