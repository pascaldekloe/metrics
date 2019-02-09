// Package metrics provides Prometheus exposition.
package metrics

import (
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// SkipTimestamp controls inclusion with sample production.
var SkipTimestamp bool

const (
	// special comment line starts
	helpPrefix = "# HELP "
	typePrefix = "# TYPE "
	// termination of TYPE comment
	gaugeTypeLineEnd     = " gauge\n"
	counterTypeLineEnd   = " counter\n"
	histogramTypeLineEnd = " histogram\n"

	gaugeType     = 'g'
	counterType   = 'c'
	histogramType = 'h'
)

// Serialisation Byte Limits
const (
	maxFloat64Text = 24
	maxUint64Text  = 20
)

// Gauge is a metric that represents a single numerical value that can
// arbitrarily go up and down.
// Multiple goroutines may invoke methods on a Gauge simultaneously.
type Gauge struct {
	// value first due atomic alignment requirement
	valueBits uint64
	// sample line start as in <name> <label-map>? ' '
	prefix string
}

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

// Copy is a specialised metric for recording snapshots, as opposed to holding
// the current value at all times. The precision of measurements is enhanced
// with timestamps, at the cost of performance degradation.
// Multiple goroutines may invoke methods on a Copy simultaneously.
type Copy struct {
	// value holds the latest snapshot
	value atomic.Value
	// sample line start as in <name> <label-map>? ' '
	prefix string
}

func (g *Gauge) name() string {
	return g.prefix[:strings.IndexAny(g.prefix, " {")]
}

func (c *Counter) name() string {
	return c.prefix[:strings.IndexAny(c.prefix, " {")]
}

func (c *Copy) name() string {
	return c.prefix[:strings.IndexAny(c.prefix, " {")]
}

// Snapshot is a Copy value.
type snapshot struct {
	value     float64
	timestamp uint64
}

// Get returns the current value.
func (g *Gauge) Get() float64 {
	return math.Float64frombits(atomic.LoadUint64(&g.valueBits))
}

// Get returns the current value.
func (c *Counter) Get() uint64 {
	return atomic.LoadUint64(&c.value)
}

// Get returns the latest snapshot with its Unix time in milliseconds.
func (c *Copy) Get() (value float64, timestamp uint64) {
	v := c.value.Load()
	if v == nil {
		return
	}
	ss := v.(snapshot)
	return ss.value, ss.timestamp
}

// Set replaces the current value with an update.
func (g *Gauge) Set(update float64) {
	atomic.StoreUint64(&g.valueBits, math.Float64bits(update))
}

// Set defines the latest snapshot.
func (c *Copy) Set(value float64, timestamp time.Time) {
	c.value.Store(snapshot{value, uint64(timestamp.UnixNano()) / 1e6})
}

// Add sets the value to the sum of the current value and summand.
// Note that summand can be negative (for subtraction).
func (g *Gauge) Add(summand float64) {
	for {
		oldBits := atomic.LoadUint64(&g.valueBits)
		newBits := math.Float64bits(math.Float64frombits(oldBits) + summand)
		if atomic.CompareAndSwapUint64(&g.valueBits, oldBits, newBits) {
			return
		}
		// lost race
	}
}

// Add increments the value with diff.
func (c *Counter) Add(diff uint64) {
	atomic.AddUint64(&c.value, diff)
}

// Histogram samples observations and counts them in configurable buckets.
// It also provides a sum of all observed values.
// Multiple goroutines may invoke methods on a Histogram simultaneously.
type Histogram struct {
	// Fields with atomic access first! (alignment constraint)

	// The most significant bit is the hot index [0 or 1] of each hotAndCold.
	// Writes update the hot one. The remaining 63 least significant bits
	// count the number of writes initiated. Write transactions start with
	// incrementing the counter, and finnish with incrementing the respective
	// hotAndColdCounts, as a marker for completion.
	//
	// Reads swap the hot–cold in a switchMutex lock. A cooldown is awaited
	// in that lock by comparing the number of writes with the initiation
	// count. Once they match the last write transaction on the now cool one
	// completed. Data may be consumed at that point. Data must be merged
	// into the new hot before unlock.
	countAndHotIndex uint64
	hotAndColdCounts [2]uint64

	// sums all observed values
	hotAndColdSumBits [2]uint64
	// counts for each bucketBucketBounds, +Inf excluded
	hotAndColdBuckets [2][]uint64

	// upper value for each bucket: sorted, +Inf excluded
	bucketBounds []float64

	// locked on hotAndCold change [reads]
	switchMutex sync.Mutex

	// metric identifier
	name string

	// corresponding name + label serials for each bucketBounds
	bucketPrefixes []string
}

// Add applies the value to the countings.
func (h *Histogram) Add(value float64) {
	// define bucket index
	i := sort.SearchFloat64s(h.bucketBounds, value)

	// start transaction with count increment & resolve hot index [0 or 1]
	hotIndex := atomic.AddUint64(&h.countAndHotIndex, 1) >> 63

	// update hot buckets; skips +Inf
	if buckets := h.hotAndColdBuckets[hotIndex]; i < len(buckets) {
		atomic.AddUint64(&buckets[i], 1)
	}

	// update hot sum
	for {
		oldBits := atomic.LoadUint64(&h.hotAndColdSumBits[hotIndex])
		newBits := math.Float64bits(math.Float64frombits(oldBits) + value)
		if atomic.CompareAndSwapUint64(&h.hotAndColdSumBits[hotIndex], oldBits, newBits) {
			break
		}
		// lost race
	}

	// end transaction by matching count(AndHotIndex).
	atomic.AddUint64(&h.hotAndColdCounts[hotIndex], 1)
}

// AddSince applies the duration since start (in seconds) to the countings.
// E.g., the following one-liner tracks function latencies.
//
//	defer Latency.AddSince(time.Now())
//
func (h *Histogram) AddSince(start time.Time) {
	h.Add(float64(time.Now().UnixNano()-start.UnixNano()) * 1e-9)
}

// Metric is a named record.
type metric struct {
	typeComment string
	helpComment string

	typeID int

	counter      *Counter
	gauge        *Gauge
	gaugeL1s     []*Map1LabelGauge
	gaugeL2s     []*Map2LabelGauge
	gaugeL3s     []*Map3LabelGauge
	histogram    *Histogram
	copyFallback *Copy
}

// Register
var (
	// register lock
	mutex sync.Mutex
	// mapping by name
	indices = make(map[string]uint32)
	// consistent order
	metrics []*metric
)

// MustNewGauge registers a new Gauge. The function panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Combinations of Gauge, Copy and the various
// label options are allowed though. The Copy is ignored once a Gauge is
// registered under the same name. This fallback allows warm starts.
func MustNewGauge(name string) *Gauge {
	mustValidName(name)

	mutex.Lock()

	var m *metric
	if index, ok := indices[name]; ok {
		m = metrics[index]
		if m.typeID != gaugeType || m.gauge != nil {
			panic("metrics: name already in use")
		}
	} else {
		indices[name] = uint32(len(metrics))
		m = &metric{
			typeComment: typePrefix + name + gaugeTypeLineEnd,
			typeID:      gaugeType,
		}
		metrics = append(metrics, m)
	}

	g := &Gauge{prefix: name + " "}
	m.gauge = g

	mutex.Unlock()

	return g
}

// MustNewCounter registers a new Counter. The function panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Combinations of Counter, Copy and the various
// label options are allowed though. The Copy is ignored once a Counter is
// registered under the same name. This fallback allows warm starts.
func MustNewCounter(name string) *Counter {
	mustValidName(name)

	mutex.Lock()

	var m *metric
	if index, ok := indices[name]; ok {
		m = metrics[index]
		if m.typeID != counterType || m.counter != nil {
			panic("metrics: name already in use")
		}
	} else {
		indices[name] = uint32(len(metrics))
		m = &metric{
			typeComment: typePrefix + name + counterTypeLineEnd,
			typeID:      counterType,
		}
		metrics = append(metrics, m)
	}

	c := &Counter{prefix: name + " "}
	m.counter = c

	mutex.Unlock()

	return c
}

var negativeInfinity = math.Inf(-1)
var positiveInfinity = math.Inf(0)

// MustNewHistogram registers a new Histogram. Buckets define the upper
// boundaries, preferably in ascending order. Special cases not-a-number
// and both infinities are ignored.
// The function panics when name was registered before, or when name
// doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*.
func MustNewHistogram(name string, buckets ...float64) *Histogram {
	mustValidName(name)

	// sort, dedupelicate, drop not-a-number, drop infinities
	sort.Float64s(buckets)
	last := negativeInfinity
	for i, f := range buckets {
		if f > last {
			last = f
			continue
		}
		// delete
		buckets = append(buckets[:i], buckets[i+1:]...)
	}

	if i := len(buckets) - 1; i >= 0 && buckets[i] > math.MaxFloat64 {
		buckets = buckets[:i]
	}

	mutex.Lock()

	var m *metric
	if index, ok := indices[name]; ok {
		m = metrics[index]
		if m.typeID != histogramType || m.histogram != nil {
			panic("metrics: name already in use")
		}
	} else {
		indices[name] = uint32(len(metrics))
		m = &metric{
			typeComment: typePrefix + name + histogramTypeLineEnd,
			typeID:      histogramType,
		}
		metrics = append(metrics, m)
	}

	h := &Histogram{name: name, bucketBounds: buckets}
	m.histogram = h

	mutex.Unlock()

	// One memory allocation for hot & cold.
	// Must be alligned for atomic access!
	bothBuckets := make([]uint64, 2*len(buckets))
	h.hotAndColdBuckets[0] = bothBuckets[:len(buckets)]
	h.hotAndColdBuckets[1] = bothBuckets[len(buckets):]

	// serialise prefixes for each bucket once
	h.bucketPrefixes = make([]string, len(buckets))
	for i, f := range buckets {
		const suffixHead, suffixTail = `{le="`, `"} `
		var buf strings.Builder
		buf.Grow(len(name) + len(suffixHead) + maxFloat64Text + len(suffixTail))
		buf.WriteString(name)
		buf.WriteString(suffixHead)
		buf.WriteString(strconv.FormatFloat(f, 'g', -1, 64))
		buf.WriteString(suffixTail)
		h.bucketPrefixes[i] = buf.String()
	}

	return h
}

// MustNewGaugeCopy registers a new Copy. The function panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Combinations of Gauge, Copy and the various
// label options are allowed though. The Copy is ignored once a Gauge is
// registered under the same name. This fallback allows warm starts.
func MustNewGaugeCopy(name string) *Copy {
	mustValidName(name)

	mutex.Lock()

	var m *metric
	if index, ok := indices[name]; ok {
		m = metrics[index]
		if m.typeID != gaugeType || m.copyFallback != nil {
			panic("metrics: name already in use")
		}
	} else {
		indices[name] = uint32(len(metrics))
		m = &metric{
			typeComment: typePrefix + name + gaugeTypeLineEnd,
			typeID:      gaugeType,
		}
		metrics = append(metrics, m)
	}

	c := &Copy{prefix: name + " "}
	m.copyFallback = c

	mutex.Unlock()

	return c
}

// MustNewCounterCopy registers a new Copy. The function panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Combinations of Counter, Copy and the various
// label options are allowed though. The Copy is ignored once a Counter is
// registered under the same name. This fallback allows warm starts.
func MustNewCounterCopy(name string) *Copy {
	mustValidName(name)

	mutex.Lock()

	var m *metric
	if index, ok := indices[name]; ok {
		m = metrics[index]
		if m.typeID != counterType || m.copyFallback != nil {
			panic("metrics: name already in use")
		}
	} else {
		indices[name] = uint32(len(metrics))
		m = &metric{
			typeComment: typePrefix + name + counterTypeLineEnd,
			typeID:      counterType,
		}
		metrics = append(metrics, m)
	}

	c := &Copy{prefix: name + " "}
	m.copyFallback = c

	mutex.Unlock()

	return c
}

func mustValidName(s string) {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '_' || c == ':' {
			continue
		}
		if i == 0 || c < '0' || c > '9' {
			panic("metrics: name doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*")
		}
	}
}

// Help sets the comment. Any previous text is replaced.
func (g *Gauge) Help(text string) *Gauge {
	help(g.name(), text)
	return g
}

// Help sets the comment for the metric name. Any previous text is replaced.
func (m *Map1LabelGauge) Help(text string) *Map1LabelGauge {
	help(m.name, text)
	return m
}

// Help sets the comment for the metric name. Any previous text is replaced.
func (m *Map2LabelGauge) Help(text string) *Map2LabelGauge {
	help(m.name, text)
	return m
}

// Help sets the comment for the metric name. Any previous text is replaced.
func (m *Map3LabelGauge) Help(text string) *Map3LabelGauge {
	help(m.name, text)
	return m
}

// Help sets the comment. Any previous text is replaced.
func (c *Counter) Help(text string) *Counter {
	help(c.name(), text)
	return c
}

// Help sets the comment. Any previous text is replaced.
func (h *Histogram) Help(text string) *Histogram {
	help(h.name, text)
	return h
}

// Help sets the comment. Any previous text is replaced.
func (c *Copy) Help(text string) *Copy {
	help(c.name(), text)
	return c
}

var helpEscapes = strings.NewReplacer("\n", `\n`, `\`, `\\`)

func help(name, text string) {
	var buf strings.Builder
	buf.Grow(len(helpPrefix) + len(name) + len(text) + 2)
	buf.WriteString(helpPrefix)
	buf.WriteString(name)
	buf.WriteByte(' ')
	helpEscapes.WriteString(&buf, text)
	buf.WriteByte('\n')

	mutex.Lock()
	metrics[indices[name]].helpComment = buf.String()
	mutex.Unlock()
}
