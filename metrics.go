// Package metrics provides atomic measures and Prometheus exposition.
//
// Gauge, Counter and Histogram represent live running values and Sample covers
// captures. Metrics are permanent-the API has no delete.
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

// SkipTimestamp controls time inclusion with sample serialisation.
// That is, when false, live running values get the current time and
// Sample provides its own.
var SkipTimestamp bool

const (
	// special comment line starts
	typePrefix = "\n# TYPE "
	helpPrefix = "# HELP "
	// termination of TYPE comment
	gaugeTypeLineEnd     = " gauge\n"
	counterTypeLineEnd   = " counter\n"
	histogramTypeLineEnd = " histogram\n"

	// last letter of TYPE comment
	gaugeType     = 'e'
	counterType   = 'r'
	histogramType = 'm'
)

// Serialisation Byte Limits
const (
	maxFloat64Text = 24
	maxUint64Text  = 20
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

// Gauge is a metric that represents a single numerical value that can
// arbitrarily go up and down.
// Multiple goroutines may invoke methods on a Gauge simultaneously.
type Gauge struct {
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
	// value holds the latest measurement
	value atomic.Value
	// sample line start as in <name> <label-map>? ' '
	prefix string
}

// Measurement is a Sample capture.
type measurement struct {
	value     float64
	timestamp uint64
}

func (c *Counter) name() string {
	return c.prefix[:strings.IndexAny(c.prefix, " {")]
}

func (g *Gauge) name() string {
	return g.prefix[:strings.IndexAny(g.prefix, " {")]
}

func (s *Sample) name() string {
	return s.prefix[:strings.IndexAny(s.prefix, " {")]
}

// Get returns the current value.
func (c *Counter) Get() uint64 {
	return atomic.LoadUint64(&c.value)
}

// Get returns the current value.
func (g *Gauge) Get() float64 {
	return math.Float64frombits(atomic.LoadUint64(&g.valueBits))
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

// Set replaces the current value with an update.
func (g *Gauge) Set(update float64) {
	atomic.StoreUint64(&g.valueBits, math.Float64bits(update))
}

// Set replaces the capture.
func (s *Sample) Set(value float64, timestamp time.Time) {
	s.value.Store(measurement{value, uint64(timestamp.UnixNano()) / 1e6})
}

// Add increments the value with diff.
func (c *Counter) Add(diff uint64) {
	atomic.AddUint64(&c.value, diff)
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
		runtime.Gosched()
	}
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
	// counts for each bucketBounds, +Inf omitted
	hotAndColdBuckets [2][]uint64

	// locked on hotAndCold switch (by reads)
	switchMutex sync.Mutex

	// End of fields with atomic access! (alignment constraint)

	// metric identifier
	name string

	// upper value for each bucket, sorted, +Inf omitted
	bucketBounds []float64

	// corresponding name + label serials for each bucketBounds, including +Inf
	bucketPrefixes         []string
	sumPrefix, countPrefix string
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

	gauge     *Gauge
	counter   *Counter
	gaugeL1s  []*Map1LabelGauge
	gaugeL2s  []*Map2LabelGauge
	gaugeL3s  []*Map3LabelGauge
	histogram *Histogram
	sample    *Sample
}

func (m *metric) typeID() byte {
	return m.typeComment[len(m.typeComment)-2]
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

// MustNewCounter registers a new Counter. The function panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Combinations of Counter, Sample and the various
// label options are allowed though. The Sample is ignored once a Counter is
// registered under the same name. This fallback allows warm starts.
func MustNewCounter(name string) *Counter {
	mustValidName(name)

	mutex.Lock()

	var m *metric
	if index, ok := indices[name]; ok {
		m = metrics[index]
		if m.typeID() != counterType || m.counter != nil {
			panic("metrics: name already in use")
		}
	} else {
		indices[name] = uint32(len(metrics))
		m = &metric{typeComment: typePrefix + name + counterTypeLineEnd}
		metrics = append(metrics, m)
	}

	c := &Counter{prefix: name + " "}
	m.counter = c

	mutex.Unlock()

	return c
}

// MustNewGauge registers a new Gauge. The function panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Combinations of Gauge, Sample and the various
// label options are allowed though. The Sample is ignored once a Gauge is
// registered under the same name. This fallback allows warm starts.
func MustNewGauge(name string) *Gauge {
	mustValidName(name)

	mutex.Lock()

	var m *metric
	if index, ok := indices[name]; ok {
		m = metrics[index]
		if m.typeID() != gaugeType || m.gauge != nil {
			panic("metrics: name already in use")
		}
	} else {
		indices[name] = uint32(len(metrics))
		m = &metric{typeComment: typePrefix + name + gaugeTypeLineEnd}
		metrics = append(metrics, m)
	}

	g := &Gauge{prefix: name + " "}
	m.gauge = g

	mutex.Unlock()

	return g
}

var negativeInfinity = math.Inf(-1)

// MustNewHistogram registers a new Histogram. Buckets define the upper
// boundaries, preferably in ascending order. Special cases not-a-number
// and both infinities are ignored.
// The function panics when name was registered before, or when name
// doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*.
func MustNewHistogram(name string, buckets ...float64) *Histogram {
	mustValidName(name)

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

	mutex.Lock()

	var m *metric
	if index, ok := indices[name]; ok {
		m = metrics[index]
		if m.typeID() != histogramType || m.histogram != nil {
			panic("metrics: name already in use")
		}
	} else {
		indices[name] = uint32(len(metrics))
		m = &metric{typeComment: typePrefix + name + histogramTypeLineEnd}
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

	// serialise prefixes only once here
	h.bucketPrefixes = make([]string, len(buckets), len(buckets)+1)
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
	h.bucketPrefixes = append(h.bucketPrefixes, name+`{le="+Inf"} `)
	h.sumPrefix = name + "_sum "
	h.countPrefix = name + "_count "

	return h
}

// MustNewGaugeSample registers a new Sample. The function panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Combinations of Gauge, Sample and the various
// label options are allowed though. The Sample is ignored once a Gauge is
// registered under the same name. This fallback allows warm starts.
func MustNewGaugeSample(name string) *Sample {
	mustValidName(name)

	mutex.Lock()

	var m *metric
	if index, ok := indices[name]; ok {
		m = metrics[index]
		if m.typeID() != gaugeType || m.sample != nil {
			panic("metrics: name already in use")
		}
	} else {
		indices[name] = uint32(len(metrics))
		m = &metric{typeComment: typePrefix + name + gaugeTypeLineEnd}
		metrics = append(metrics, m)
	}

	s := &Sample{prefix: name + " "}
	m.sample = s

	mutex.Unlock()

	return s
}

// MustNewCounterSample registers a new Sample. The function panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Combinations of Counter, Sample and the various
// label options are allowed though. The Sample is ignored once a Counter is
// registered under the same name. This fallback allows warm starts.
func MustNewCounterSample(name string) *Sample {
	mustValidName(name)

	mutex.Lock()

	var m *metric
	if index, ok := indices[name]; ok {
		m = metrics[index]
		if m.typeID() != counterType || m.sample != nil {
			panic("metrics: name already in use")
		}
	} else {
		indices[name] = uint32(len(metrics))
		m = &metric{typeComment: typePrefix + name + counterTypeLineEnd}
		metrics = append(metrics, m)
	}

	s := &Sample{prefix: name + " "}
	m.sample = s

	mutex.Unlock()

	return s
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
func (c *Counter) Help(text string) *Counter {
	help(c.name(), text)
	return c
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
func (h *Histogram) Help(text string) *Histogram {
	help(h.name, text)
	return h
}

// Help sets the comment. Any previous text is replaced.
func (s *Sample) Help(text string) *Sample {
	help(s.name(), text)
	return s
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
