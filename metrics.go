// Package metrics provides Prometheus exposition.
package metrics

import (
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const (
	helpPrefix = "# HELP "
	typePrefix = "# TYPE "
)

// Gauge is a metric that represents a single numerical value that can
// arbitrarily go up and down.
type Gauge struct {
	n     int64
	label []byte
	help  []byte
}

// Counter is a cumulative metric that represents a single monotonically
// increasing counter whose value can only increase or be reset to zero on
// restart.
type Counter struct {
	n     uint64
	label []byte
	help  []byte
}

// Set updates the value.
// Multiple goroutines may invoke this method simultaneously.
func (g *Gauge) Set(update int64) { atomic.StoreInt64(&g.n, update) }

// Add increments the value with diff.
// Multiple goroutines may invoke this method simultaneously.
func (g *Gauge) Add(diff int64) { atomic.AddInt64(&g.n, diff) }

// Add increments the value with diff.
// Multiple goroutines may invoke this method simultaneously.
func (c *Counter) Add(diff uint64) { atomic.AddUint64(&c.n, diff) }

var (
	gaugeMutex   sync.Mutex
	gauges       []*Gauge
	gaugeIndices = make(map[string]int)

	counterMutex   sync.Mutex
	counters       []*Counter
	counterIndices = make(map[string]int)
)

// MustPlaceGauge registers a new Gauge if name hasn't been used before.
// The function panics when name does not match [a-zA-Z_:][a-zA-Z0-9_:]*
// or when name is already in use as another metric type.
func MustPlaceGauge(name string) *Gauge {
	mustValidName(name)

	label := make([]byte, 15+2*len(name))
	copy(label, typePrefix)
	copy(label[7:], name)
	copy(label[7+len(name):], " gauge\n")
	copy(label[14+len(name):], name)
	label[len(label)-1] = ' '

	g := &Gauge{label: label}

	gaugeMutex.Lock()
	if i, ok := gaugeIndices[name]; ok {
		g = gauges[i]
	} else {
		gaugeIndices[name] = len(gauges)
		gauges = append(gauges, g)
	}
	gaugeMutex.Unlock()

	return g
}

// MustPlaceCounter registers a new Counter if name hasn't been used before.
// The function panics when name does not match [a-zA-Z_:][a-zA-Z0-9_:]*
// or when name is already in use as another metric type.
func MustPlaceCounter(name string) *Counter {
	mustValidName(name)

	label := make([]byte, 17+2*len(name))
	copy(label, typePrefix)
	copy(label[7:], name)
	copy(label[7+len(name):], " counter\n")
	copy(label[16+len(name):], name)
	label[len(label)-1] = ' '

	c := &Counter{label: label}

	counterMutex.Lock()
	if i, ok := counterIndices[name]; ok {
		c = counters[i]
	} else {
		counterIndices[name] = len(counters)
		counters = append(counters, c)
	}
	counterMutex.Unlock()

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
	g.help = labelHelp(g.label, text)
	return g
}

// Help sets the text.
func (c *Counter) Help(text string) (this *Counter) {
	c.help = labelHelp(c.label, text)
	return c
}

func labelHelp(label []byte, text string) []byte {
	// get name from label
	var name []byte
	for i, c := range label {
		if i > len(typePrefix) && c == ' ' {
			name = label[len(typePrefix):i]
			break
		}
	}

	// compose help in new buffer
	buf := make([]byte, len(helpPrefix)+len(name)+1, len(helpPrefix)+len(name)+len(text)+2)
	copy(buf, helpPrefix)
	copy(buf[len(helpPrefix):], name)
	buf[len(helpPrefix)+len(name)] = ' '

	// add escaped text
	for i := 0; i < len(text); i++ {
		c := text[i]
		switch c {
		case '\n':
			buf = append(buf, '\\', 'n')
		case '\\':
			buf = append(buf, '\\', '\\')
		default:
			buf = append(buf, c)
		}
	}

	// terminate help line
	return append(buf, '\n')
}

var epochMilliseconds = func() int64 { return time.Now().UnixNano() / 1e6 }

// HTTPHandler serves Prometheus on a /metrics endpoint.
func HTTPHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", http.MethodOptions+", "+http.MethodGet+", "+http.MethodHead)
		if r.Method != http.MethodOptions {
			http.Error(w, "read-only resource", http.StatusMethodNotAllowed)
		}

		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")

	timeTail := make([]byte, 1, 15)
	timeTail[0] = ' '

	buf := make([]byte, 0, 16)

	gaugeMutex.Lock()
	gaugesView := gauges[:]
	gaugeMutex.Unlock()
	for _, g := range gaugesView {
		w.Write(g.help)
		w.Write(g.label)
		w.Write(strconv.AppendInt(buf, atomic.LoadInt64(&g.n), 10))
		w.Write(append(strconv.AppendInt(timeTail, epochMilliseconds(), 10), '\n'))
	}

	counterMutex.Lock()
	countersView := counters[:]
	counterMutex.Unlock()
	for _, c := range countersView {
		w.Write(c.help)
		w.Write(c.label)
		w.Write(strconv.AppendUint(buf, atomic.LoadUint64(&c.n), 10))
		w.Write(append(strconv.AppendInt(timeTail, epochMilliseconds(), 10), '\n'))
	}
}
