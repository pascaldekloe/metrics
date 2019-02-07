// Package metrics provides Prometheus exposition.
package metrics

import (
	"math"
	"net/http"
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
	helpPrefix     = "# HELP "
	typePrefix     = "# TYPE "
	gaugeLineEnd   = " gauge\n"
	counterLineEnd = " counter\n"
	untypedLineEnd = " untyped\n"

	gaugeType   = 'g'
	counterType = 'c'
	untypedType = 'u'
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

// Metric is a named record.
type metric struct {
	typeComment string
	helpComment string

	typeID int

	counter  *Counter
	gauge    *Gauge
	gaugeL1s []*Link1LabelGauge
	gaugeL2s []*Link2LabelGauge
	gaugeL3s []*Link3LabelGauge
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

	var g *Gauge
	if index, ok := indices[name]; ok {
		m := metrics[index]
		if m.typeID != gaugeType {
			panic("metrics: name in use as another type")
		}
		g = m.gauge
	} else {
		g = &Gauge{label: name + " "}
		indices[name] = uint32(len(metrics))
		metrics = append(metrics, &metric{
			typeComment: typePrefix + name + gaugeLineEnd,
			typeID:      gaugeType,
			gauge:       g,
		})
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

	var c *Counter
	if index, ok := indices[name]; ok {
		m := metrics[index]
		if m.typeID != counterType {
			panic("metrics: name in use as another type")
		}
		c = m.counter
	} else {
		c = &Counter{label: name + " "}
		indices[name] = uint32(len(metrics))
		metrics = append(metrics, &metric{
			typeComment: typePrefix + name + counterLineEnd,
			typeID:      counterType,
			counter:     c,
		})
	}

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

func sampleTail(buf []byte) []byte {
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

	tail := sampleTail(make([]byte, 15))
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
				buf, tail = m.gauge.sample(w, buf, tail)
			}
			for _, l1 := range m.gaugeL1s {
				for _, g := range l1.gauges {
					buf, tail = g.sample(w, buf, tail)
				}
			}
			for _, l2 := range m.gaugeL2s {
				for _, g := range l2.gauges {
					buf, tail = g.sample(w, buf, tail)
				}
			}
			for _, l3 := range m.gaugeL3s {
				for _, g := range l3.gauges {
					buf, tail = g.sample(w, buf, tail)
				}
			}

		case counterType:
			if m.counter != nil {
				buf, tail = m.counter.sample(w, buf, tail)
			}
		}
	}

	w.Write(buf)
}

func (g *Gauge) sample(w http.ResponseWriter, buf, tail []byte) ([]byte, []byte) {
	if len(buf) > 3900 {
		w.Write(buf)
		buf = buf[:0]
		tail = sampleTail(tail)
	}

	buf = append(buf, g.label...)
	buf = strconv.AppendFloat(buf, g.Get(), 'g', -1, 64)
	buf = append(buf, tail...)
	return buf, tail
}

func (c *Counter) sample(w http.ResponseWriter, buf, tail []byte) ([]byte, []byte) {
	if len(buf) > 3900 {
		w.Write(buf)
		buf = buf[:0]
		tail = sampleTail(tail)
	}

	buf = append(buf, c.label...)
	buf = strconv.AppendUint(buf, c.Get(), 10)
	buf = append(buf, tail...)
	return buf, tail
}
