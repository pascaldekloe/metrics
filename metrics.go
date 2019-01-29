// Package metrics provides Prometheus exposition.
package metrics

import (
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// Gauge is a metric that represents a single numerical value that can
// arbitrarily go up and down.
type Gauge struct {
	n     int64
	label []byte
}

// Counter is a cumulative metric that represents a single monotonically
// increasing counter whose value can only increase or be reset to zero on
// restart.
type Counter struct {
	n     uint64
	label []byte
}

// Set updates the state.
// Multiple goroutines may invoke this method simultaneously.
func (g *Gauge) Set(update int64) { atomic.StoreInt64(&g.n, update) }

// Add changes the state.
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
// A panic is raised when name is already in use as another metric type.
func MustPlaceGauge(name string) *Gauge {
	mustValidName(name)

	g := &Gauge{label: make([]byte, 14+len(name))}
	copy(g.label, "# TYPE gauge\n")
	copy(g.label[13:], name)
	g.label[13+len(name)] = ' '

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
// A panic is raised when name is already in use as another metric type.
func MustPlaceCounter(name string) *Counter {
	mustValidName(name)

	c := &Counter{label: make([]byte, 16+len(name))}
	copy(c.label, "# TYPE counter\n")
	copy(c.label[15:], name)
	c.label[15+len(name)] = ' '

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
	g.label = labelHelp(g.label, text)
	return g
}

// Help sets the text.
func (g *Counter) Help(text string) (this *Counter) {
	g.label = labelHelp(g.label, text)
	return g
}

func labelHelp(label []byte, text string) []byte {
	const helpPrefix = "# HELP "

	// drop previous if any
	for i, c := range label {
		if i < len(helpPrefix) {
			if helpPrefix[i] != c {
				break
			}
		} else if c == '\n' {
			label = label[i+1:]
			break
		}
	}

	// start help in new buffer
	buf := make([]byte, len(helpPrefix), len(helpPrefix)+len(text)+len(label)+2)
	copy(buf, helpPrefix)

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

	// terminate help
	buf = append(buf, '\n')

	// join with label
	return append(buf, label...)
}

var nowMilli = func() int64 { return time.Now().UnixNano() / 1e6 }

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
		w.Write(g.label)
		w.Write(strconv.AppendInt(buf, atomic.LoadInt64(&g.n), 10))
		w.Write(append(strconv.AppendInt(timeTail, nowMilli(), 10), '\n'))
	}

	counterMutex.Lock()
	countersView := counters[:]
	counterMutex.Unlock()

	for _, c := range countersView {
		w.Write(c.label)
		w.Write(strconv.AppendUint(buf, atomic.LoadUint64(&c.n), 10))
		w.Write(append(strconv.AppendInt(timeTail, nowMilli(), 10), '\n'))
	}
}
