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
	value int64
	head  []byte
	help  []byte
}

// Counter is a cumulative metric that represents a single monotonically
// increasing counter whose value can only increase or be reset to zero on
// restart.
type Counter struct {
	value uint64
	head  []byte
	help  []byte
}

// Set updates the value.
// Multiple goroutines may invoke this method simultaneously.
func (g *Gauge) Set(update int64) { atomic.StoreInt64(&g.value, update) }

// Add increments the value with diff.
// Multiple goroutines may invoke this method simultaneously.
func (g *Gauge) Add(diff int64) { atomic.AddInt64(&g.value, diff) }

// Add increments the value with diff.
// Multiple goroutines may invoke this method simultaneously.
func (c *Counter) Add(diff uint64) { atomic.AddUint64(&c.value, diff) }

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

	head := make([]byte, 15+2*len(name))
	copy(head, typePrefix)
	copy(head[7:], name)
	copy(head[7+len(name):], " gauge\n")
	copy(head[14+len(name):], name)
	head[len(head)-1] = ' '

	g := &Gauge{head: head}

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

	head := make([]byte, 17+2*len(name))
	copy(head, typePrefix)
	copy(head[7:], name)
	copy(head[7+len(name):], " counter\n")
	copy(head[16+len(name):], name)
	head[len(head)-1] = ' '

	c := &Counter{head: head}

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
	g.help = headHelp(g.head, text)
	return g
}

// Help sets the text.
func (c *Counter) Help(text string) (this *Counter) {
	c.help = headHelp(c.head, text)
	return c
}

func headHelp(head []byte, text string) []byte {
	// get name from head
	var name []byte
	for i, c := range head {
		if i > len(typePrefix) && c == ' ' {
			name = head[len(typePrefix):i]
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

var appendTimeTail = func(buf []byte) []byte {
	ms := time.Now().UnixNano() / 1e6
	buf = strconv.AppendInt(buf, ms, 10)
	return append(buf, '\n')
}

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

	buf := make([]byte, 0, 4096)
	timeTail := make([]byte, 1, 15)
	timeTail[0] = ' '

	gaugeMutex.Lock()
	gaugesView := gauges[:]
	gaugeMutex.Unlock()

	timeTail = appendTimeTail(timeTail)
	for _, g := range gaugesView {
		buf = g.appendSample(buf)
		buf = append(buf, timeTail...)
		if len(buf) > 3800 {
			w.Write(buf)
			buf = buf[:0]
			timeTail = appendTimeTail(timeTail[:1])
		}
	}

	counterMutex.Lock()
	countersView := counters[:]
	counterMutex.Unlock()

	timeTail = appendTimeTail(timeTail[:1])
	for _, c := range countersView {
		buf = c.appendSample(buf)
		buf = append(buf, timeTail...)
		if len(buf) > 3800 {
			w.Write(buf)
			buf = buf[:0]
			timeTail = appendTimeTail(timeTail[:1])
		}
	}

	w.Write(buf)
}

// AppendSample applies the line format, open-ended.
func (g *Gauge) appendSample(buf []byte) []byte {
	buf = append(buf, g.help...)
	buf = append(buf, g.head...)
	return strconv.AppendInt(buf, atomic.LoadInt64(&g.value), 10)
}

// AppendSample applies the line format, open-ended.
func (c *Counter) appendSample(buf []byte) []byte {
	buf = append(buf, c.help...)
	buf = append(buf, c.head...)
	return strconv.AppendUint(buf, atomic.LoadUint64(&c.value), 10)
}
