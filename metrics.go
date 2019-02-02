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

// Special Comments
const (
	helpPrefix      = "# HELP "
	typePrefix      = "# TYPE "
	gaugeTypeTail   = " gauge\n"
	counterTypeTail = " counter\n"
)

// Gauge is a metric that represents a single numerical value that can
// arbitrarily go up and down.
type Gauge struct {
	value int64
	head  string
	help  []byte
}

// RealGauge is a Gauge, but for real numbers instead of integers.
// arbitrarily go up and down.
type RealGauge struct {
	value uint64
	head  string
	help  []byte
}

// Counter is a cumulative metric that represents a single monotonically
// increasing counter whose value can only increase or be reset to zero on
// restart.
type Counter struct {
	value uint64
	head  string
	help  []byte
}

// Set updates the value.
// Multiple goroutines may invoke this method simultaneously.
func (g *Gauge) Set(update int64) {
	atomic.StoreInt64(&g.value, update)
}

// Set updates the value.
// Multiple goroutines may invoke this method simultaneously.
func (g *RealGauge) Set(update float64) {
	atomic.StoreUint64(&g.value, math.Float64bits(update))
}

// Add sets the value to the sum of the current value and summand.
// Multiple goroutines may invoke this method simultaneously.
func (g *Gauge) Add(summand int64) {
	atomic.AddInt64(&g.value, summand)
}

// Add sets the value to the sum of the current value and summand.
// Multiple goroutines may invoke this method simultaneously.
func (g *RealGauge) Add(summand float64) {
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

func (g *Gauge) name() string {
	return g.head[strings.LastIndexByte(g.head, '\n')+1 : len(g.head)-1]
}

func (g *RealGauge) name() string {
	return g.head[strings.LastIndexByte(g.head, '\n')+1 : len(g.head)-1]
}

func (c *Counter) name() string {
	return c.head[strings.LastIndexByte(c.head, '\n')+1 : len(c.head)-1]
}

var (
	mutex      sync.Mutex
	indices    = make(map[string]uint32)
	gauges     []*Gauge
	realGauges []*RealGauge
	counters   []*Counter
)

// MustPlaceGauge registers a new Gauge if name hasn't been used before.
// The function panics when name is in use as onther metric type or when
// name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*.
func MustPlaceGauge(name string) *Gauge {
	mustValidName(name)
	head := formatGaugeHead(name)

	mutex.Lock()

	var g *Gauge
	if index, ok := indices[name]; ok {
		if int(index) >= len(gauges) || gauges[index].name() != name {
			panic("metrics: name in use as another type")
		}
		g = gauges[index]
	} else {
		g = &Gauge{head: head}
		indices[name] = uint32(len(gauges))
		gauges = append(gauges, g)
	}

	mutex.Unlock()

	return g
}

// MustPlaceRealGauge registers a new RealGauge if name hasn't been used before.
// The function panics when name is in use as onther metric type or when
// name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*.
func MustPlaceRealGauge(name string) *RealGauge {
	mustValidName(name)
	head := formatGaugeHead(name)

	mutex.Lock()

	var g *RealGauge
	if index, ok := indices[name]; ok {
		if int(index) >= len(realGauges) || realGauges[index].name() != name {
			panic("metrics: name in use as another type")
		}
		g = realGauges[index]
	} else {
		g = &RealGauge{head: head}
		indices[name] = uint32(len(realGauges))
		realGauges = append(realGauges, g)
	}

	mutex.Unlock()

	return g
}

func formatGaugeHead(name string) string {
	var buf strings.Builder
	buf.Grow(15 + 2*len(name))
	buf.WriteString(typePrefix)
	buf.WriteString(name)
	buf.WriteString(gaugeTypeTail)
	buf.WriteString(name)
	buf.WriteByte(' ')
	return buf.String()
}

// MustPlaceCounter registers a new Counter if name hasn't been used before.
// The function panics when name is in use as onther metric type or when
// name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*.
func MustPlaceCounter(name string) *Counter {
	mustValidName(name)

	var head strings.Builder
	head.Grow(17 + 2*len(name))
	head.WriteString(typePrefix)
	head.WriteString(name)
	head.WriteString(counterTypeTail)
	head.WriteString(name)
	head.WriteByte(' ')

	mutex.Lock()

	var c *Counter
	if index, ok := indices[name]; ok {
		if int(index) >= len(counters) || counters[index].name() != name {
			panic("metrics: name in use as another type")
		}
		c = counters[index]
	} else {
		c = &Counter{head: head.String()}
		indices[name] = uint32(len(counters))
		counters = append(counters, c)
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

// Help sets the text.
func (g *Gauge) Help(text string) (this *Gauge) {
	help(g.name(), text)
	return g
}

// Help sets the text.
func (c *Counter) Help(text string) (this *Counter) {
	help(c.name(), text)
	return c
}

var (
	helpMutex   sync.RWMutex
	helpIndices = make(map[string]uint32)
	helps       [][]byte
)

func help(name, text string) {
	headLen := len(helpPrefix) + len(name) + 1
	help := make([]byte, headLen, headLen+len(text)+1)

	copy(help, helpPrefix)
	copy(help[len(helpPrefix):], name)
	help[headLen-1] = ' '

	// add escaped text
	var offset int
	for i := 0; i < len(text); i++ {
		switch c := text[i]; c {
		case '\n':
			help = append(help, text[offset:i]...)
			help = append(help, '\\', 'n')
			offset = i + 1
		case '\\':
			help = append(help, text[offset:i]...)
			help = append(help, '\\', '\\')
			offset = i + 1
		}
	}
	help = append(help, text[offset:]...)

	// terminate help line
	help = append(help, '\n')

	helpMutex.Lock()
	if i, ok := helpIndices[name]; ok {
		helps[i] = help
	} else {
		helpIndices[name] = uint32(len(helps))
		helps = append(helps, help)
	}
	helpMutex.Unlock()
}

var appendTimeTail = func(buf []byte) []byte {
	ms := time.Now().UnixNano() / 1e6
	buf = strconv.AppendInt(buf, ms, 10)
	return append(buf, '\n')
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

	helpMutex.RLock()
	for _, line := range helps {
		w.Write(line)
	}
	helpMutex.RUnlock()

	buf := make([]byte, 0, 4096)
	timeTail := make([]byte, 1, 15)
	timeTail[0] = ' '

	mutex.Lock()
	gaugesView := gauges[:]
	realGaugesView := realGauges[:]
	countersView := counters[:]
	mutex.Unlock()

	timeTail = appendTimeTail(timeTail)

	for _, g := range gaugesView {
		buf = append(buf, g.head...)
		buf = strconv.AppendInt(buf, atomic.LoadInt64(&g.value), 10)
		buf = append(buf, timeTail...)
		if len(buf) > 3900 {
			w.Write(buf)
			buf = buf[:0]
			timeTail = appendTimeTail(timeTail[:1])
		}
	}

	for _, g := range realGaugesView {
		buf = append(buf, g.head...)
		buf = strconv.AppendFloat(buf, math.Float64frombits(atomic.LoadUint64(&g.value)), 'g', -1, 64)
		buf = append(buf, timeTail...)
		if len(buf) > 3900 {
			w.Write(buf)
			buf = buf[:0]
			timeTail = appendTimeTail(timeTail[:1])
		}
	}

	for _, c := range countersView {
		buf = append(buf, c.head...)
		buf = strconv.AppendUint(buf, atomic.LoadUint64(&c.value), 10)
		buf = append(buf, timeTail...)
		if len(buf) > 3900 {
			w.Write(buf)
			buf = buf[:0]
			timeTail = appendTimeTail(timeTail[:1])
		}
	}

	w.Write(buf)
}
