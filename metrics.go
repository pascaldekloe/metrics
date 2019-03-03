// Package metrics provides atomic measures and Prometheus exposition.
// The Must functions deal with registration. Their use is intended for setup
// during application launch only.
// All metrics are permanent-the API offers no deletion.
//
// Gauges [Integer or Real], Counter and Histogram are live representations.
// Value updates should be part of the respective components that alter state.
// In all other cases, use Sample to capture with timestamps.
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
// restart.
// Multiple goroutines may invoke methods on a Counter simultaneously.
type Counter struct {
	// value first due atomic alignment requirement
	value uint64
	// sample line start as in <name> <label-map>? ' '
	prefix string
}

// Integer gauge is a metric that represents a single numerical value that can
// arbitrarily go up and down.
// Multiple goroutines may invoke methods on a Integer simultaneously.
type Integer struct {
	// value first due atomic alignment requirement
	value int64
	// sample line start as in <name> <label-map>? ' '
	prefix string
}

// Real gauge is a metric that represents a single numerical value that can
// arbitrarily go up and down.
// Multiple goroutines may invoke methods on a Real simultaneously.
type Real struct {
	// value first due atomic alignment requirement
	valueBits uint64
	// sample line start as in <name> <label-map>? ' '
	prefix string
}

// Sample is a specialised metric for measurement captures, as opposed to
// holding the current value at all times. The precision is enhanced with
// a timestamp, at the cost of performance degradation.
// Multiple goroutines may invoke methods on a Sample simultaneously.
type Sample struct {
	// value holds the last measurement
	value atomic.Value
	// sample line start as in <name> <label-map>? ' '
	prefix string
}

// Measurement is a Sample capture.
type measurement struct {
	value     float64
	timestamp uint64
}

// Get returns the current value.
func (c *Counter) Get() uint64 {
	return atomic.LoadUint64(&c.value)
}

// Get returns the current value.
func (z *Integer) Get() int64 {
	return atomic.LoadInt64(&z.value)
}

// Get returns the current value.
func (r *Real) Get() float64 {
	return math.Float64frombits(atomic.LoadUint64(&r.valueBits))
}

// Get returns the last capture with its Unix time in milliseconds.
func (s *Sample) Get() (value float64, timestamp uint64) {
	v := s.value.Load()
	if v == nil {
		return
	}
	m := v.(measurement)
	return m.value, m.timestamp
}

// Set replaces the current value.
func (z *Integer) Set(update int64) {
	atomic.StoreInt64(&z.value, update)
}

// Set replaces the current value.
func (r *Real) Set(update float64) {
	atomic.StoreUint64(&r.valueBits, math.Float64bits(update))
}

// Set replaces the current value.
func (s *Sample) Set(value float64, timestamp time.Time) {
	s.value.Store(measurement{value, uint64(timestamp.UnixNano()) / 1e6})
}

// Add increments the current value with diff.
func (c *Counter) Add(diff uint64) {
	atomic.AddUint64(&c.value, diff)
}

// Add sums the current value with diff.
// Note that diff can be negative (for subtraction).
func (z *Integer) Add(diff int64) {
	atomic.AddInt64(&z.value, diff)
}

// Histogram samples observations and counts them in configurable buckets.
// It also provides a sum of all observed values.
// Multiple goroutines may invoke methods on a Histogram simultaneously.
type Histogram struct {
	// Fields with atomic access first! (alignment constraint)

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
	// total number of observations
	hotAndColdCounts [2]uint64
	// sums all observed values
	hotAndColdSumBits [2]uint64
	// counts for each BucketBounds, +Inf omitted
	hotAndColdBuckets [2][]uint64

	// locked on hotAndCold switch (by reads)
	switchMutex sync.Mutex

	// End of fields with atomic access! (alignment constraint)

	// metric identifier
	name string

	// Upper value for each bucket, sorted, +Inf omitted.
	// This field is read-only.
	BucketBounds []float64

	// corresponding name + label serials for each BucketBounds, including +Inf
	bucketPrefixes         []string
	sumPrefix, countPrefix string
}

// Add applies value to the countings.
func (h *Histogram) Add(value float64) {
	// define bucket index
	i := sort.SearchFloat64s(h.BucketBounds, value)

	// start transaction with count increment & resolve hot index [0 or 1]
	hotIndex := atomic.AddUint64(&h.countAndHotIndex, 1) >> 63

	// update hot buckets; skips +Inf
	if buckets := h.hotAndColdBuckets[hotIndex]; i < len(buckets) {
		atomic.AddUint64(&buckets[i], 1)
	}

	// update hot sum
	for {
		p := &h.hotAndColdSumBits[hotIndex]
		oldBits := atomic.LoadUint64(p)
		newBits := math.Float64bits(math.Float64frombits(oldBits) + value)
		if atomic.CompareAndSwapUint64(p, oldBits, newBits) {
			break
		}
		// lost race
		runtime.Gosched()
	}

	// end transaction by matching count(AndHotIndex).
	atomic.AddUint64(&h.hotAndColdCounts[hotIndex], 1)
}

// AddSince applies the duration since start (in seconds) to the countings.
// E.g., the following one-liner measures function delay.
//
//	defer DurationHistogram.AddSince(time.Now())
//
func (h *Histogram) AddSince(start time.Time) {
	h.Add(float64(time.Now().UnixNano()-start.UnixNano()) * 1e-9)
}

var negativeInfinity = math.Inf(-1)

func newHistogram(name string, buckets []float64) *Histogram {
	// sort, dedupelicate, drop not-a-number, drop infinities
	sort.Float64s(buckets)
	writeIndex := 0
	last := negativeInfinity
	for _, f := range buckets {
		if f > last && f <= math.MaxFloat64 {
			buckets[writeIndex] = f
			writeIndex++
			last = f
		}
	}
	buckets = buckets[:writeIndex]

	h := &Histogram{name: name, BucketBounds: buckets}

	// One memory allocation for hot & cold.
	// Must be alligned for atomic access!
	bothBuckets := make([]uint64, 2*len(buckets))
	h.hotAndColdBuckets[0] = bothBuckets[:len(buckets)]
	h.hotAndColdBuckets[1] = bothBuckets[len(buckets):]

	h.bucketPrefixes = make([]string, len(buckets)+1)

	return h
}

func (h *Histogram) prefix(name string) {
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
}

// Get returns the current value. The buckets return is the observation count
// for each corresponding Histogram.BucketBounds element and count is the total
// observation count.
func (h *Histogram) Get(bucketAppend []uint64) (buckets []uint64, count uint64, sum float64) {
	buckets = bucketAppend

	// see struct comments for algorithm description
	h.switchMutex.Lock()
	defer h.switchMutex.Unlock()

	// Adding 1<<63 swaps the index of hotAndCold from 0 to 1,
	// or from 1 to 0, without touching the initiation counter.
	updated := atomic.AddUint64(&h.countAndHotIndex, 1<<63)

	// write destination after switch
	hotIndex := updated >> 63
	coldIndex := (^hotIndex) & 1

	// number of writes to cold
	count = updated &^ (1 << 63)

	// cooldown: await initiated writes to complete
	for count > atomic.LoadUint64(&h.hotAndColdCounts[coldIndex]) {
		runtime.Gosched()
	}

	// all: merge into hot and reset cold to zero

	// count
	atomic.StoreUint64(&h.hotAndColdCounts[coldIndex], 0)
	atomic.AddUint64(&h.hotAndColdCounts[hotIndex], count)

	// buckets
	hotBuckets := h.hotAndColdBuckets[hotIndex]
	coldBuckets := h.hotAndColdBuckets[coldIndex]
	for i := range coldBuckets {
		n := atomic.LoadUint64(&coldBuckets[i])
		atomic.StoreUint64(&coldBuckets[i], 0)
		atomic.AddUint64(&hotBuckets[i], n)

		buckets = append(buckets, n)
	}

	// sum
	sum = math.Float64frombits(atomic.LoadUint64(&h.hotAndColdSumBits[coldIndex]))
	atomic.StoreUint64(&h.hotAndColdSumBits[coldIndex], 0)
	for p := &h.hotAndColdSumBits[hotIndex]; ; {
		oldBits := atomic.LoadUint64(p)
		newBits := math.Float64bits(math.Float64frombits(oldBits) + sum)
		if atomic.CompareAndSwapUint64(p, oldBits, newBits) {
			break
		}
	}

	return
}
