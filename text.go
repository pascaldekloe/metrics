package metrics

import (
	"io"
	"net/http"
	"strconv"
	"time"
)

// SkipTimestamp controls time inclusion with sample serialisation.
// When false, then live running values are stamped with the current
// time and Sample provides its own.
var SkipTimestamp bool

const headerLine = "# Prometheus Samples\n"

// ServeHTTP provides samples all metrics.
func ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	std.ServeHTTP(resp, req)
}

// ServeHTTP provides samples all metrics.
func (r *Register) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet && req.Method != http.MethodHead {
		resp.Header().Set("Allow", http.MethodOptions+", "+http.MethodGet+", "+http.MethodHead)
		if req.Method != http.MethodOptions {
			http.Error(resp, "read-only resource", http.StatusMethodNotAllowed)
		}

		return
	}

	resp.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=UTF-8")
	r.WriteText(resp)
}

// WriteText serialises all metrics using a simple text-based exposition format.
// All errors returned by Writer are ignored by design.
func WriteText(w io.Writer) {
	std.WriteText(w)
}

// WriteText serialises all metrics using a simple text-based exposition format.
// All errors returned by Writer are ignored by design.
func (r *Register) WriteText(w io.Writer) {
	// write buffer
	buf := make([]byte, len(headerLine), 4096)
	copy(buf, headerLine)

	var buckets []uint64 // reusable

	// snapshot
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// reuse to minimise time lookups
	lineEnd := sampleLineEnd(make([]byte, 21))

	// serialise samples in order of appearance
	for _, m := range r.metrics {
		if cap(buf)-len(buf) < len(m.typeComment)+len(m.helpComment) {
			w.Write(buf)
			buf = buf[:0]
			// need fresh timestamp after Write
			lineEnd = sampleLineEnd(lineEnd)
		}

		buf = append(buf, m.typeComment...)
		buf = append(buf, m.helpComment...)

		switch m.typeID() {
		case counterType:
			var written bool

			if m.counter != nil {
				buf, lineEnd = m.counter.sample(w, buf, lineEnd)
				written = true
			}

			for _, l1 := range m.counterL1s {
				l1.mutex.Lock()
				view := l1.counters
				l1.mutex.Unlock()
				for _, c := range view {
					buf, lineEnd = c.sample(w, buf, lineEnd)
					written = true
				}
			}
			for _, l2 := range m.counterL2s {
				l2.mutex.Lock()
				view := l2.counters
				l2.mutex.Unlock()
				for _, c := range view {
					buf, lineEnd = c.sample(w, buf, lineEnd)
					written = true
				}
			}
			for _, l3 := range m.counterL3s {
				l3.mutex.Lock()
				view := l3.counters
				l3.mutex.Unlock()
				for _, c := range view {
					buf, lineEnd = c.sample(w, buf, lineEnd)
					written = true
				}
			}

			if !written {
				buf, lineEnd = m.writeSamplesText(w, buf, lineEnd)
			}

		case gaugeType:
			var written bool

			if m.integer != nil {
				buf, lineEnd = m.integer.sample(w, buf, lineEnd)
				written = true
			}
			if m.real != nil {
				buf, lineEnd = m.real.sample(w, buf, lineEnd)
				written = true
			}

			for _, l1 := range m.integerL1s {
				l1.mutex.Lock()
				view := l1.integers
				l1.mutex.Unlock()
				for _, g := range view {
					buf, lineEnd = g.sample(w, buf, lineEnd)
					written = true
				}
			}
			for _, l2 := range m.integerL2s {
				l2.mutex.Lock()
				view := l2.integers
				l2.mutex.Unlock()
				for _, g := range view {
					buf, lineEnd = g.sample(w, buf, lineEnd)
					written = true
				}
			}
			for _, l3 := range m.integerL3s {
				l3.mutex.Lock()
				view := l3.integers
				l3.mutex.Unlock()
				for _, g := range view {
					buf, lineEnd = g.sample(w, buf, lineEnd)
					written = true
				}
			}

			for _, l1 := range m.realL1s {
				l1.mutex.Lock()
				view := l1.reals
				l1.mutex.Unlock()
				for _, g := range view {
					buf, lineEnd = g.sample(w, buf, lineEnd)
					written = true
				}
			}
			for _, l2 := range m.realL2s {
				l2.mutex.Lock()
				view := l2.reals
				l2.mutex.Unlock()
				for _, g := range view {
					buf, lineEnd = g.sample(w, buf, lineEnd)
					written = true
				}
			}
			for _, l3 := range m.realL3s {
				l3.mutex.Lock()
				view := l3.reals
				l3.mutex.Unlock()
				for _, g := range view {
					buf, lineEnd = g.sample(w, buf, lineEnd)
					written = true
				}
			}

			if !written {
				buf, lineEnd = m.writeSamplesText(w, buf, lineEnd)
			}

		case histogramType:
			if m.histogram != nil {
				buf, lineEnd, buckets = m.histogram.sample(w, buf, lineEnd, buckets)
			}

			for _, l1 := range m.histogramL1s {
				l1.mutex.Lock()
				view := l1.histograms
				l1.mutex.Unlock()
				for _, h := range view {
					buf, lineEnd, buckets = h.sample(w, buf, lineEnd, buckets)
				}
			}
			for _, l2 := range m.histogramL2s {
				l2.mutex.Lock()
				view := l2.histograms
				l2.mutex.Unlock()
				for _, h := range view {
					buf, lineEnd, buckets = h.sample(w, buf, lineEnd, buckets)
				}
			}
		}
	}

	if len(buf) != 0 {
		w.Write(buf)
	}
}

func (m *metric) writeSamplesText(w io.Writer, buf, lineEnd []byte) ([]byte, []byte) {
	if m.sample != nil {
		buf, lineEnd = m.sample.sample(w, buf, lineEnd)
	}
	for _, l1 := range m.sampleL1s {
		l1.mutex.Lock()
		view := l1.samples
		l1.mutex.Unlock()
		for _, v := range view {
			buf, lineEnd = v.sample(w, buf, lineEnd)
		}
	}
	for _, l2 := range m.sampleL2s {
		l2.mutex.Lock()
		view := l2.samples
		l2.mutex.Unlock()
		for _, v := range view {
			buf, lineEnd = v.sample(w, buf, lineEnd)
		}
	}
	for _, l3 := range m.sampleL3s {
		l3.mutex.Lock()
		view := l3.samples
		l3.mutex.Unlock()
		for _, v := range view {
			buf, lineEnd = v.sample(w, buf, lineEnd)
		}
	}

	return buf, lineEnd
}

func (c *Counter) sample(w io.Writer, buf, lineEnd []byte) ([]byte, []byte) {
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

func (g *Integer) sample(w io.Writer, buf, lineEnd []byte) ([]byte, []byte) {
	if cap(buf)-len(buf) < len(g.prefix)+maxInt64Text+len(lineEnd) {
		w.Write(buf)
		buf = buf[:0]
		// need fresh timestamp after Write
		lineEnd = sampleLineEnd(lineEnd)
	}

	buf = append(buf, g.prefix...)
	buf = strconv.AppendInt(buf, g.Get(), 10)
	buf = append(buf, lineEnd...)

	return buf, lineEnd
}

func (g *Real) sample(w io.Writer, buf, lineEnd []byte) ([]byte, []byte) {
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

func (h *Histogram) sample(w io.Writer, buf, lineEnd []byte, buckets []uint64) ([]byte, []byte, []uint64) {
	// calculate buffer space required
	fit := len(h.sumPrefix) + len(h.countPrefix) + maxFloat64Text + maxUint64Text + 2*len(lineEnd)
	for _, prefix := range h.bucketPrefixes {
		fit += len(prefix) + maxUint64Text + len(lineEnd)
	}
	if len(buf) != 0 && cap(buf)-len(buf) < fit {
		w.Write(buf)
		buf = buf[:0]
	}
	// buf either fits or empty (to minimise memory allocation)

	buckets, count, sum := h.Get(buckets[:0])

	// need fresh timestamp after Write and/or Lock
	lineEnd = sampleLineEnd(lineEnd)

	buf = append(buf, h.countPrefix...)
	offset := len(buf)
	buf = strconv.AppendUint(buf, count, 10)
	countSerial := buf[offset:]
	buf = append(buf, lineEnd...)

	// buckets
	var cum uint64
	for i, prefix := range h.bucketPrefixes {
		if i >= len(buckets) {
			// (redundant) +Inf bucket
			buf = append(buf, prefix...)
			buf = append(buf, countSerial...)
			buf = append(buf, lineEnd...)
			break
		}

		cum += buckets[i]
		buf = append(buf, prefix...)
		buf = strconv.AppendUint(buf, cum, 10)
		buf = append(buf, lineEnd...)
	}

	// sum
	buf = append(buf, h.sumPrefix...)
	buf = strconv.AppendFloat(buf, sum, 'g', -1, 64)
	buf = append(buf, lineEnd...)

	return buf, lineEnd, buckets
}

func (s *Sample) sample(w io.Writer, buf, lineEnd []byte) ([]byte, []byte) {
	if cap(buf)-len(buf) < len(s.prefix)+maxFloat64Text+21 {
		w.Write(buf)
		buf = buf[:0]
		// need fresh timestamp after Write
		lineEnd = sampleLineEnd(lineEnd)
	}

	if value, timestamp := s.Get(); timestamp != 0 {
		buf = append(buf, s.prefix...)
		buf = strconv.AppendFloat(buf, value, 'g', -1, 64)
		if !SkipTimestamp {
			buf = append(buf, ' ')
			buf = strconv.AppendUint(buf, timestamp, 10)
		}
		buf = append(buf, '\n')
	}

	return buf, lineEnd
}

// SampleLineEnd may include a timestamp, and terminates with a double line feed.
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
