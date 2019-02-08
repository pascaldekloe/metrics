// Package metrics provides Prometheus exposition.
package metrics

import (
	"math"
	"net/http"
	"runtime"
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

// Get returns the current value.
func (g *Gauge) Get() float64 {
	return math.Float64frombits(atomic.LoadUint64(&g.valueBits))
}

// Get returns the current value.
func (c *Counter) Get() uint64 {
	return atomic.LoadUint64(&c.value)
}

// Set replaces the current value with an update.
func (g *Gauge) Set(update float64) {
	atomic.StoreUint64(&g.valueBits, math.Float64bits(update))
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

// Metric is a named record.
type metric struct {
	typeComment string
	helpComment string

	typeID int

	counter   *Counter
	gauge     *Gauge
	gaugeL1s  []*Map1LabelGauge
	gaugeL2s  []*Map2LabelGauge
	gaugeL3s  []*Map3LabelGauge
	histogram *Histogram
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
// [a-zA-Z_:][a-zA-Z0-9_:]*. Combinations of labeled and unlabeled
// instances are allowed though.
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
// [a-zA-Z_:][a-zA-Z0-9_:]*.
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

var helpEscapes = strings.NewReplacer("\n", `\n`, `\`, `\\`)

// MustHelp sets the comment text for a metric name. Any previous text value is
// discarded. The function panics when name is not in use.
func MustHelp(name, text string) {
	var buf strings.Builder
	buf.Grow(len(helpPrefix) + len(name) + len(text) + 2)
	buf.WriteString(helpPrefix)
	buf.WriteString(name)
	buf.WriteByte(' ')
	helpEscapes.WriteString(&buf, text)
	buf.WriteByte('\n')

	mutex.Lock()

	index, ok := indices[name]
	if !ok {
		panic("metrics: name not in use")
	}
	metrics[index].helpComment = buf.String()

	mutex.Unlock()
}

func sampleLineEnd(buf []byte) []byte {
	buf = buf[:1]

	if SkipTimestamp {
		buf[0] = '\n'
	} else {
		buf[0] = ' '
		ms := time.Now().UnixNano() / 1e6
		buf = strconv.AppendInt(buf, ms, 10)
		buf = append(buf, '\n')
	}

	return buf
}

// HTTPHandler serves all metrics using a simple text-based exposition format.
func HTTPHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", http.MethodOptions+", "+http.MethodGet+", "+http.MethodHead)
		if r.Method != http.MethodOptions {
			http.Error(w, "read-only resource", http.StatusMethodNotAllowed)
		}

		return
	}

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=UTF-8")

	// write buffer
	buf := make([]byte, 0, 4096)

	// snapshot
	mutex.Lock()
	view := metrics[:]
	mutex.Unlock()

	// reuse to minimise time lookups
	lineEnd := sampleLineEnd(make([]byte, 21))

	// serialise samples in order of appearance
	for _, m := range view {
		if cap(buf)-len(buf) < len(m.typeComment)+len(m.helpComment) {
			w.Write(buf)
			buf = buf[:0]
			// need fresh timestamp after Write
			lineEnd = sampleLineEnd(lineEnd)
		}

		// Comments
		buf = append(buf, m.typeComment...)
		buf = append(buf, m.helpComment...)

		switch m.typeID {
		case gaugeType:
			if m.gauge != nil {
				buf, lineEnd = m.gauge.sample(w, buf, lineEnd)
			}
			for _, l1 := range m.gaugeL1s {
				for _, g := range l1.gauges {
					buf, lineEnd = g.sample(w, buf, lineEnd)
				}
			}
			for _, l2 := range m.gaugeL2s {
				for _, g := range l2.gauges {
					buf, lineEnd = g.sample(w, buf, lineEnd)
				}
			}
			for _, l3 := range m.gaugeL3s {
				for _, g := range l3.gauges {
					buf, lineEnd = g.sample(w, buf, lineEnd)
				}
			}

		case counterType:
			buf, lineEnd = m.counter.sample(w, buf, lineEnd)

		case histogramType:
			buf, lineEnd = m.histogram.sample(w, buf, lineEnd)
		}
	}

	w.Write(buf)
}

func (g *Gauge) sample(w http.ResponseWriter, buf, lineEnd []byte) ([]byte, []byte) {
	if cap(buf)-len(buf) < len(g.prefix)+maxFloat64Text+len(lineEnd) {
		w.Write(buf)
		buf = buf[:0]
		// need fresh timestamp after Write
		lineEnd = sampleLineEnd(lineEnd)
	}

	buf = append(buf, g.prefix...)
	buf = strconv.AppendFloat(buf, g.Get(), 'g', -1, 64)
	buf = append(buf, lineEnd...)
	return buf, lineEnd
}

func (c *Counter) sample(w http.ResponseWriter, buf, lineEnd []byte) ([]byte, []byte) {
	if cap(buf)-len(buf) < len(c.prefix)+maxUint64Text+len(lineEnd) {
		w.Write(buf)
		buf = buf[:0]
		// need fresh timestamp after Write
		lineEnd = sampleLineEnd(lineEnd)
	}

	buf = append(buf, c.prefix...)
	buf = strconv.AppendUint(buf, c.Get(), 10)
	buf = append(buf, lineEnd...)
	return buf, lineEnd
}

func (h *Histogram) sample(w http.ResponseWriter, buf, lineEnd []byte) ([]byte, []byte) {
	const infSuffix, sumSuffix, countSuffix = `{le="+Inf"} `, "_sum ", "_count "

	// calculate buffer space; start with static
	fit := len(infSuffix) + len(sumSuffix) + len(countSuffix) + 2*maxFloat64Text + maxUint64Text
	fit += 3 * (len(h.name) + len(lineEnd))
	for _, prefix := range h.bucketPrefixes {
		fit += len(prefix) + maxUint64Text + len(lineEnd)
	}
	if len(buf) != 0 && cap(buf)-len(buf) < fit {
		w.Write(buf)
		buf = buf[:0]
	}
	// buf either fits or empty (to minimise memory allocation)

	// see struct comments for algorithm description
	h.switchMutex.Lock()

	// need fresh timestamp after Write and/or Lock
	lineEnd = sampleLineEnd(lineEnd)

	// Adding 1<<63 swaps the index of hotAndCold from 0 to 1,
	// or from 1 to 0, without touching the initiation counter.
	updated := atomic.AddUint64(&h.countAndHotIndex, 1<<63)

	// write destination after switch
	hotIndex := updated >> 63
	coldIndex := (^hotIndex) & 1

	// number of writes to cold
	initiated := updated &^ (1 << 63)

	// cooldown: await initiated writes to complete
	for initiated > atomic.LoadUint64(&h.hotAndColdCounts[coldIndex]) {
		runtime.Gosched()
	}

	// for all:
	// * reset cold to zero
	// * merge into hot
	// * serialise to buf

	atomic.StoreUint64(&h.hotAndColdCounts[coldIndex], 0)
	atomic.AddUint64(&h.hotAndColdCounts[hotIndex], initiated)
	buf = append(buf, h.name...)
	buf = append(buf, countSuffix...)
	offset := len(buf)
	buf = strconv.AppendUint(buf, initiated, 10)
	countSerial := buf[offset:]
	buf = append(buf, lineEnd...)

	// each non-infinite bucket
	hotBuckets := h.hotAndColdBuckets[hotIndex]
	coldBuckets := h.hotAndColdBuckets[coldIndex]
	var count uint64
	for i := range coldBuckets {
		n := atomic.LoadUint64(&coldBuckets[i])
		atomic.StoreUint64(&coldBuckets[i], 0)
		atomic.AddUint64(&hotBuckets[i], n)

		count += n
		buf = append(buf, h.bucketPrefixes[i]...)
		buf = strconv.AppendUint(buf, count, 10)
		buf = append(buf, lineEnd...)
	}

	// no idea why we need the redundant infinite bucket?
	buf = append(buf, h.name...)
	buf = append(buf, infSuffix...)
	buf = append(buf, countSerial...)
	buf = append(buf, lineEnd...)

	// sum
	sumBits := atomic.LoadUint64(&h.hotAndColdSumBits[coldIndex])
	atomic.StoreUint64(&h.hotAndColdSumBits[coldIndex], 0)
	sum := math.Float64frombits(sumBits)
	lostRace := false
	for p := &h.hotAndColdSumBits[hotIndex]; ; {
		oldBits := atomic.LoadUint64(p)
		newBits := math.Float64bits(math.Float64frombits(oldBits) + sum)
		if atomic.CompareAndSwapUint64(p, oldBits, newBits) {
			break
		}
		lostRace = true
	}

	if lostRace {
		// refresh timestamp
		lineEnd = sampleLineEnd(lineEnd)
	}

	h.switchMutex.Unlock()

	buf = append(buf, h.name...)
	buf = append(buf, sumSuffix...)
	buf = strconv.AppendFloat(buf, sum, 'g', -1, 64)
	buf = append(buf, lineEnd...)

	return buf, lineEnd
}
