// Package metrics provides atomic measures and Prometheus exposition.
//
// Counter, Integer, Real and Histogram are live representations of events.
// Value updates should be part of the respective implementation. Otherwise,
// use Sample for captures with a timestamp.
//
// The Must functions deal with registration. Their use is intended for setup
// during application launch only.
// All metrics are permanent-the API offers no deletion by design.
package metrics

import (
	"math"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Special Comments
const (
	// line starts
	typePrefix = "\n# TYPE "
	helpPrefix = "# HELP "

	// terminations of TYPE comment
	gaugeTypeLineEnd     = " gauge\n"
	counterTypeLineEnd   = " counter\n"
	histogramTypeLineEnd = " histogram\n"
)

// Serialisation Byte Limits
const (
	maxFloat64Text = 24
	maxUint64Text  = 20
	maxInt64Text   = 21
)

// Counter is a cumulative metric that represents a single monotonically
// increasing counter whose value can only increase or be reset to zero on
// restart. The default/initial value is zero.
// Multiple goroutines may invoke methods on a Counter simultaneously.
type Counter struct {
	// value first due atomic alignment requirement
	value uint64
	// fixed start of serial line is <name> <label-map>? ' '
	prefix string
}

// Integer gauge is a metric that represents a single numerical value that can
// arbitrarily go up and down. The default/initial value is zero.
// Multiple goroutines may invoke methods on a Integer simultaneously.
type Integer struct {
	// value first due atomic alignment requirement
	value int64
	// fixed start of serial line is <name> <label-map>? ' '
	prefix string
}

// Real gauge is a metric that represents a single numerical value that can
// arbitrarily go up and down. The default/initial value is zero.
// Multiple goroutines may invoke methods on a Real simultaneously.
type Real struct {
	// value first due atomic alignment requirement
	valueBits uint64
	// fixed start of serial line is <name> <label-map>? ' '
	prefix string
}

// Sample is a specialised metric for measurement captures, as opposed to
// holding the current value at all times. The precision is enhanced with
// a timestamp, at the cost of performance degradation. Serialisation
// omits samples with a zero timestamp. The default/initial value is zero
// with a zero timestamp.
// Multiple goroutines may invoke methods on a Sample simultaneously.
type Sample struct {
	mux       sync.Mutex
	value     float64 // current capture
	timestamp uint64  // capture moment
	// fixed start of serial line is <name> <label-map>? ' '
	prefix string
}

func parseMetricName(s string) string {
	i := strings.IndexAny(s, " {")
	if i >= 0 {
		return s[:i]
	}
	return s
}

// Name returns the metric identifier.
func (m *Counter) Name() string { return parseMetricName(m.prefix) }

// Name returns the metric identifier.
func (m *Integer) Name() string { return parseMetricName(m.prefix) }

// Name returns the metric identifier.
func (m *Real) Name() string { return parseMetricName(m.prefix) }

// Name returns the metric identifier.
func (m *Sample) Name() string { return parseMetricName(m.prefix) }

// Name returns the metric identifier.
func (m *Histogram) Name() string { return parseMetricName(m.bucketPrefixes[0]) }

// Labels returns a new map if m has labels.
func (m *Counter) Labels() map[string]string { return parseMetricLabels(m.prefix) }

// Labels returns a new map if m has labels.
func (m *Integer) Labels() map[string]string { return parseMetricLabels(m.prefix) }

// Labels returns a new map if m has labels.
func (m *Real) Labels() map[string]string { return parseMetricLabels(m.prefix) }

// Labels returns a new map if m has labels.
func (m *Sample) Labels() map[string]string { return parseMetricLabels(m.prefix) }

// Labels returns a new map if m has labels.
func (m *Histogram) Labels() map[string]string { return parseMetricLabels(m.bucketPrefixes[0]) }

// Get returns the current value.
func (m *Counter) Get() uint64 {
	return atomic.LoadUint64(&m.value)
}

// Get returns the current value.
func (m *Integer) Get() int64 {
	return atomic.LoadInt64(&m.value)
}

// Get returns the current value.
func (m *Real) Get() float64 {
	return math.Float64frombits(atomic.LoadUint64(&m.valueBits))
}

// Get returns the current value with its Unix time in milliseconds.
func (m *Sample) Get() (value float64, timestamp uint64) {
	m.mux.Lock()
	defer m.mux.Unlock()
	return m.value, m.timestamp
}

// Set defines the current value.
func (m *Integer) Set(update int64) {
	atomic.StoreInt64(&m.value, update)
}

// Set defines the current value.
func (m *Real) Set(update float64) {
	atomic.StoreUint64(&m.valueBits, math.Float64bits(update))
}

// SetSeconds defines the current value.
func (m *Real) SetSeconds(update time.Duration) {
	m.Set(float64(update) / float64(time.Second))
}

// Set defines the current value.
func (m *Sample) Set(value float64, timestamp time.Time) {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.value = value
	m.timestamp = uint64(timestamp.UnixNano()) / 1e6
}

// SetSeconds defines the current value.
func (m *Sample) SetSeconds(value time.Duration, timestamp time.Time) {
	m.Set(float64(value)/float64(time.Second), timestamp)
}

// Add increments the current value with n.
func (m *Counter) Add(n uint64) {
	atomic.AddUint64(&m.value, n)
}

// Add summs the current value with n.
// Note that n can be negative (for subtraction).
func (m *Integer) Add(n int64) {
	atomic.AddInt64(&m.value, n)
}

// Histogram samples observations and counts them in configurable buckets.
// It also provides a sum of all observed values.
// Multiple goroutines may invoke methods on a Histogram simultaneously.
type Histogram struct {
	// Counters are memory aligned for atomic access.
	// The 15 64-bit padding entries ensure isolation
	// with CPU cache lines up to 128 bytes in size.

	// The total number of observations are stored at index 0 when the hot
	// index is 0. Otherwise, the index is 16 (when the hot index is 1).
	hotAndColdCounts [2 * 16]uint64
	// The sums of all observed values are stored at index 0 when the hot
	// index is 0. Otherwise, the index is 16 (when the hot index is 1).
	hotAndColdSumBits [2 * 16]uint64

	// CountAndHotIndex enables lock-free writes with use of atomic updates.
	// The most significant bit is the hot index [0 or 1] of each hotAndCold
	// field. Writes update the hot one. All remaining bits count the number
	// of writes initiated. Write transactions start by incrementing this
	// counter, and finish by incrementing the hotAndColdCounts field, as a
	// marker for completion.
	//
	// Reads swap the hot–cold in a switchMutex lock. A cooldown is awaited
	// (in such lock) by comparing the number of writes with the initiation
	// count. Once they match, then the last write transaction on the now
	// cool one completed. All cool fields must be merged into the new hot
	// before the unlock of switchMutex.
	countAndHotIndex uint64

	// counts for each BucketBounds, +Inf omitted
	hotAndColdBuckets [2][]uint64

	// Upper value for each bucket, sorted, +Inf omitted.
	// This field is read-only.
	BucketBounds []float64

	// fixed start of each serial line is <name> '{le="' … '"} '
	bucketPrefixes []string // including +Inf
	// fixed start of serial line is <name> '_sum '
	sumPrefix string
	// fixed start of serial line is <name> '_count '
	countPrefix string

	// locked on hotAndCold switch (by reads)
	switchMutex sync.Mutex
}

// Add applies value to the countings.
func (h *Histogram) Add(value float64) {
	// define bucket index with padding
	pi := sort.SearchFloat64s(h.BucketBounds, value) * 16

	// start transaction with count increment & resolve hot index [0 or 1]
	hotIndex := atomic.AddUint64(&h.countAndHotIndex, 1) >> 63

	// update hot buckets; skips +Inf
	if buckets := h.hotAndColdBuckets[hotIndex]; pi < len(buckets) {
		atomic.AddUint64(&buckets[pi], 1)
	}

	// update hot sum
	for p := &h.hotAndColdSumBits[hotIndex*16]; ; {
		oldBits := atomic.LoadUint64(p)
		newBits := math.Float64bits(math.Float64frombits(oldBits) + value)
		if atomic.CompareAndSwapUint64(p, oldBits, newBits) {
			break
		}
		// lost race
		runtime.Gosched()
	}

	// end transaction by matching count(AndHotIndex).
	atomic.AddUint64(&h.hotAndColdCounts[hotIndex*16], 1)
}

// AddSince applies the number of seconds since start to the countings.
// The following one-liner measures the execution time of a function.
//
//	defer DurationHistogram.AddSince(time.Now())
//
func (h *Histogram) AddSince(start time.Time) {
	h.Add(float64(time.Since(start)) * 1e-9)
}

func newHistogram(name string, bucketBounds []float64) *Histogram {
	// Use copy of bucketBounds to prevent unexpected mutations,
	// in case the variadic was invoked with a collapsed slice.
	var a []float64
	for _, f := range bucketBounds {
		// skip NaN and ∞
		if f == f && f <= math.MaxFloat64 {
			a = append(a, f)
		}
	}
	if len(a) < 2 {
		bucketBounds = a
	} else {
		sort.Float64s(a)
		bucketBounds = a[:1]
		for _, f := range a[1:] {
			if f > bucketBounds[len(bucketBounds)-1] {
				bucketBounds = append(bucketBounds, f)
			}
		}
	}

	// Counters are memory aligned for atomic access.
	// The 15 64-bit padding entries ensure isolation
	// with CPU cache lines up to 128 bytes in size.
	bucketCounts := make([]uint64, 2*16*len(bucketBounds))

	h := Histogram{
		bucketPrefixes: make([]string, len(bucketBounds)+1),
		BucketBounds:   bucketBounds,
		hotAndColdBuckets: [2][]uint64{
			bucketCounts[:len(bucketCounts)/2],
			bucketCounts[len(bucketCounts)/2:],
		},
	}

	// install fixed start of serial lines
	for i, f := range h.BucketBounds {
		const suffixHead, suffixTail = `{le="`, `"} `
		var buf strings.Builder
		buf.Grow(len(name) + len(suffixHead) + maxFloat64Text + len(suffixTail))
		buf.WriteString(name)
		buf.WriteString(suffixHead)
		buf.WriteString(strconv.FormatFloat(f, 'g', -1, 64))
		buf.WriteString(suffixTail)
		h.bucketPrefixes[i] = buf.String()
	}
	h.bucketPrefixes[len(h.BucketBounds)] = name + `{le="+Inf"} `
	h.countPrefix = name + "_count "
	h.sumPrefix = name + "_sum "

	return &h
}

// Get appends the observation counts for each Histogram.BucketBounds to a and
// returns the resulting slice (as buckets). The count return has the total
// number of observations, a.k.a. the positive inifinity bucket.
func (h *Histogram) Get(a []uint64) (buckets []uint64, count uint64, sum float64) {
	buckets = a

	// see struct comments for algorithm description
	h.switchMutex.Lock()
	defer h.switchMutex.Unlock()

	// Adding 1<<63 swaps the index of hotAndCold from 0 to 1,
	// or from 1 to 0, without touching the initiation counter.
	updated := atomic.AddUint64(&h.countAndHotIndex, 1<<63)

	// write destination after switch
	hotIndex := updated >> 63
	coldIndex := hotIndex ^ 1

	// number of writes to cold
	count = updated &^ (1 << 63)

	// cooldown: await initiated writes to complete
	for count > atomic.LoadUint64(&h.hotAndColdCounts[coldIndex*16]) {
		runtime.Gosched()
	}

	// merge count into hot and reset cold to zero
	atomic.StoreUint64(&h.hotAndColdCounts[coldIndex*16], 0)
	atomic.AddUint64(&h.hotAndColdCounts[hotIndex*16], count)

	// merge buckets into hot and reset cold to zero
	hotBuckets := h.hotAndColdBuckets[hotIndex]
	coldBuckets := h.hotAndColdBuckets[coldIndex]
	for i := 0; i < len(coldBuckets); i += 16 {
		n := atomic.LoadUint64(&coldBuckets[i])
		atomic.StoreUint64(&coldBuckets[i], 0)
		atomic.AddUint64(&hotBuckets[i], n)

		buckets = append(buckets, n)
	}

	// merge sum into hot and reset cold to zero
	sum = math.Float64frombits(atomic.LoadUint64(&h.hotAndColdSumBits[coldIndex*16]))
	atomic.StoreUint64(&h.hotAndColdSumBits[coldIndex*16], 0)
	for p := &h.hotAndColdSumBits[hotIndex*16]; ; {
		oldBits := atomic.LoadUint64(p)
		newBits := math.Float64bits(math.Float64frombits(oldBits) + sum)
		if atomic.CompareAndSwapUint64(p, oldBits, newBits) {
			break
		}
		// lost race
		runtime.Gosched()
	}

	return
}
