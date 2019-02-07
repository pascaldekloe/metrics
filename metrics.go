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

// Special Comments
const (
	helpPrefix       = "# HELP "
	typePrefix       = "# TYPE "
	gaugeLineEnd     = " gauge\n"
	counterLineEnd   = " counter\n"
	histogramLineEnd = " histogram\n"
	untypedLineEnd   = " untyped\n"

	gaugeType     = 'g'
	counterType   = 'c'
	histogramType = 'h'
	untypedType   = 'u'
)

// Gauge is a metric that represents a single numerical value that can
// arbitrarily go up and down.
type Gauge struct {
	value uint64
	label string
}

// Counter is a cumulative metric that represents a single monotonically
// increasing counter whose value can only increase or be reset to zero on
// restart.
type Counter struct {
	value uint64
	label string
}

// Get returns the current value.
// Multiple goroutines may invoke this method simultaneously.
func (g *Gauge) Get() float64 {
	return math.Float64frombits(atomic.LoadUint64(&g.value))
}

// Get returns the current value.
// Multiple goroutines may invoke this method simultaneously.
func (c *Counter) Get() uint64 {
	return atomic.LoadUint64(&c.value)
}

// Set replaces the current value with an update.
// Multiple goroutines may invoke this method simultaneously.
func (g *Gauge) Set(update float64) {
	atomic.StoreUint64(&g.value, math.Float64bits(update))
}

// Add sets the value to the sum of the current value and summand.
// Multiple goroutines may invoke this method simultaneously.
func (g *Gauge) Add(summand float64) {
	for {
		current := atomic.LoadUint64(&g.value)
		update := math.Float64bits(math.Float64frombits(current) + summand)
		if atomic.CompareAndSwapUint64(&g.value, current, update) {
			return
		}
	}
}

// Add increments the value with diff.
// Multiple goroutines may invoke this method simultaneously.
func (c *Counter) Add(diff uint64) {
	atomic.AddUint64(&c.value, diff)
}

type Histogram struct {
	// Fields with atomic access first! (alignment constraints)

	// The most significant bit is the hot index [0 or 1] of hotAndCold.
	// Writes update the hot one. The remaining 63 least significant bits
	// count the number of writes initiated.
	//
	// Reads swap the hot–cold in a switchMutex lock. A cooldown is awaited
	// in that lock by comparing the number of writes with the initiation
	// count (from countAndHotIndex). Once they match the last write on the
	// now cool one completed. Data may be consumed at that point. Data must
	// be merged into the new hot before unlock.
	countAndHotIndex uint64

	hotAndColdCounts  [2]uint64
	hotAndColdSumBits [2]uint64
	hotAndColdBuckets [2][]uint64

	// locked on hotAndCold change [reads]
	switchMutex sync.Mutex

	name string
	// upper value for each bucket: sorted, +Inf excluded
	bucketBounds []float64
	// corresponding name + label serials for bucketBounds [map]
	bucketLabels []string
}

// Add applies the value to the countings.
func (h *Histogram) Add(value float64) {
	// define bucket index
	i := sort.SearchFloat64s(h.bucketBounds, value)

	// start transaction with count increment & apply hot index [0 or 1]
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
	gaugeL1s  []*Link1LabelGauge
	gaugeL2s  []*Link2LabelGauge
	gaugeL3s  []*Link3LabelGauge
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

// MustPlaceGauge registers a new Gauge if name hasn't been used before.
// The function panics when name is in use as another metric type or when
// name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*.
func MustPlaceGauge(name string) *Gauge {
	mustValidName(name)

	mutex.Lock()

	var m *metric
	if index, ok := indices[name]; ok {
		m = metrics[index]
		if m.typeID != gaugeType {
			panic("metrics: name in use as another type")
		}
	} else {
		indices[name] = uint32(len(metrics))
		m = &metric{
			typeComment: typePrefix + name + gaugeLineEnd,
			typeID:      gaugeType,
		}
		metrics = append(metrics, m)
	}

	g := m.gauge
	if g == nil {
		g = &Gauge{label: name + " "}
		m.gauge = g
	}

	mutex.Unlock()

	return g
}

// MustPlaceCounter registers a new Counter if name hasn't been used before.
// The function panics when name is in use as another metric type or when
// name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*.
func MustPlaceCounter(name string) *Counter {
	mustValidName(name)

	mutex.Lock()

	var m *metric
	if index, ok := indices[name]; ok {
		m = metrics[index]
		if m.typeID != counterType {
			panic("metrics: name in use as another type")
		}
	} else {
		indices[name] = uint32(len(metrics))
		m = &metric{
			typeComment: typePrefix + name + counterLineEnd,
			typeID:      counterType,
		}
		metrics = append(metrics, m)
	}

	c := m.counter
	if c == nil {
		c = &Counter{label: name + " "}
		m.counter = c
	}

	mutex.Unlock()

	return c
}

var negativeInfinity = math.Inf(-1)
var positiveInfinity = math.Inf(0)

// MustPlaceHistogram registers a new Histogram if name hasn't been used
// before. Buckets define the upper boundaries, preferably in ascending
// order. Special cases not-a-number and both infinities are ignored.
// The function panics when name is in use as another metric type or when
// name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*.
func MustPlaceHistogram(name string, buckets ...float64) *Histogram {
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
	if i := len(buckets) - 1; i < 0 || buckets[i] > math.MaxFloat64 {
		buckets = buckets[:i]
	}

	mutex.Lock()

	var m *metric
	if index, ok := indices[name]; ok {
		m = metrics[index]
		if m.typeID != histogramType {
			panic("metrics: name in use as another type")
		}
	} else {
		indices[name] = uint32(len(metrics))
		m = &metric{
			typeComment: typePrefix + name + histogramLineEnd,
			typeID:      histogramType,
		}
		metrics = append(metrics, m)
	}

	h := m.histogram
	if h == nil {
		// One memory allocation for hot & cold.
		// Must be alligned for atomic access!
		bothBuckets := make([]uint64, 2*len(buckets))

		h = &Histogram{
			name: name,
			hotAndColdBuckets: [...][]uint64{
				bothBuckets[:len(buckets)],
				bothBuckets[len(buckets):],
			},
			bucketBounds: buckets,
			bucketLabels: make([]string, len(buckets)),
		}
		m.histogram = h

		// serialise labels for each bucket once
		for i, f := range buckets {
			var buf strings.Builder
			buf.Grow(len(name) + 29)
			buf.WriteString(name)
			buf.WriteString(`{le="`)
			buf.WriteString(strconv.FormatFloat(f, 'g', -1, 64))
			buf.WriteString(`"} `)
			h.bucketLabels[i] = buf.String()
		}
	}

	mutex.Unlock()

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
	buf.Grow(len(helpPrefix) + len(name) + +len(text) + 2)
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

	lineEnd := sampleLineEnd(make([]byte, 21))
	buf := make([]byte, 0, 4096)

	// snapshot
	mutex.Lock()
	view := metrics[:]
	mutex.Unlock()

	for _, m := range view {
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
			if m.counter != nil {
				buf, lineEnd = m.counter.sample(w, buf, lineEnd)
			}

		case histogramType:
			if m.histogram != nil {
				buf, lineEnd = m.histogram.sample(w, buf, lineEnd)
			}
		}
	}

	w.Write(buf)
}

func (g *Gauge) sample(w http.ResponseWriter, buf, lineEnd []byte) ([]byte, []byte) {
	if len(buf) > 3900 {
		w.Write(buf)
		buf = buf[:0]
		// need fresh timestamp after Write
		lineEnd = sampleLineEnd(lineEnd)
	}

	buf = append(buf, g.label...)
	buf = strconv.AppendFloat(buf, g.Get(), 'g', -1, 64)
	buf = append(buf, lineEnd...)
	return buf, lineEnd
}

func (c *Counter) sample(w http.ResponseWriter, buf, lineEnd []byte) ([]byte, []byte) {
	if len(buf) > 3900 {
		w.Write(buf)
		buf = buf[:0]
		// need fresh timestamp after Write
		lineEnd = sampleLineEnd(lineEnd)
	}

	buf = append(buf, c.label...)
	buf = strconv.AppendUint(buf, c.Get(), 10)
	buf = append(buf, lineEnd...)
	return buf, lineEnd
}

func (h *Histogram) sample(w http.ResponseWriter, buf, lineEnd []byte) ([]byte, []byte) {
	fieldFit := len(h.name) + 50 + len(lineEnd)
	if len(buf) != 0 && (len(h.bucketLabels)+2)*fieldFit > cap(buf)-len(buf) {
		w.Write(buf)
		buf = buf[:0]
	}
	// buf either fits or empty (to minimise memory allocation)

	h.switchMutex.Lock()

	// need fresh timestamp after Write or Lock
	lineEnd = sampleLineEnd(lineEnd)

	// Adding 1<<63 swaps the index of hotAndCold from 0 to 1,
	// or from 1 to 0, without touching the initiation counter.
	updated := atomic.AddUint64(&h.countAndHotIndex, 1<<63)

	// write destination after switch
	hotIndex := updated >> 63
	coldIndex := (^hotIndex) & 1

	// number of writes to cold
	initiated := updated &^ (1 << 63)

	// await initiated writes to complete
	for p := &h.hotAndColdCounts[coldIndex]; initiated > atomic.LoadUint64(p); runtime.Gosched() {
	}
	atomic.AddUint64(&h.hotAndColdCounts[hotIndex], initiated)
	atomic.StoreUint64(&h.hotAndColdCounts[coldIndex], 0)

	// each cold value:
	// * serialise to buf
	// * merge to hot
	// * reset to zero

	// buckets first
	hotBuckets := h.hotAndColdBuckets[hotIndex]
	coldBuckets := h.hotAndColdBuckets[coldIndex]
	var count uint64
	for i := range coldBuckets {
		n := atomic.LoadUint64(&coldBuckets[i])
		atomic.StoreUint64(&coldBuckets[i], 0)
		atomic.AddUint64(&hotBuckets[i], n)

		count += n
		buf = append(buf, h.bucketLabels[i]...)
		buf = strconv.AppendUint(buf, count, 10)
		buf = append(buf, lineEnd...)
	}

	sum := math.Float64frombits(atomic.LoadUint64(&h.hotAndColdSumBits[coldIndex]))
	atomic.StoreUint64(&h.hotAndColdSumBits[coldIndex], 0)

	h.switchMutex.Unlock()

	// merge into hot sum
	for p := &h.hotAndColdSumBits[hotIndex]; ; {
		oldBits := atomic.LoadUint64(p)
		newBits := math.Float64bits(math.Float64frombits(oldBits) + sum)
		if atomic.CompareAndSwapUint64(p, oldBits, newBits) {
			break
		}
		// lost race
	}

	buf = append(buf, h.name...)
	buf = append(buf, `{le="+Inf"} `...)
	offset := len(buf)
	buf = strconv.AppendUint(buf, initiated, 10)
	countSerial := buf[offset:]
	buf = append(buf, lineEnd...)

	buf = append(buf, h.name...)
	buf = append(buf, "_sum "...)
	buf = strconv.AppendFloat(buf, sum, 'g', -1, 64)
	buf = append(buf, lineEnd...)

	buf = append(buf, h.name...)
	buf = append(buf, "_count "...)
	buf = append(buf, countSerial...)
	buf = append(buf, lineEnd...)

	if len(buf) > 3900 {
		w.Write(buf)
		buf = buf[:0]
	}
	// need fresh timestamp after optimistic locks and/or Write
	lineEnd = sampleLineEnd(lineEnd)

	return buf, lineEnd
}
