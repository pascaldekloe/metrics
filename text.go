package metrics

import (
	"io"
	"net/http"
	"strconv"
	"time"
)

// SkipTimestamp controls time inclusion with sample serialisation.
// When false, then live running values are stamped with the current
// time and Samples provide their own time.
var SkipTimestamp = false

const headerLine = "# Prometheus Samples\n"

// ServeHTTP provides a sample of each metric.
func ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	std.ServeHTTP(resp, req)
}

// ServeHTTP provides a sample of each metric.
func (reg *Register) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet && req.Method != http.MethodHead {
		resp.Header().Set("Allow", http.MethodOptions+", "+http.MethodGet+", "+http.MethodHead)
		if req.Method != http.MethodOptions {
			http.Error(resp, "read-only resource", http.StatusMethodNotAllowed)
		}

		return
	}

	resp.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=UTF-8")
	reg.WriteText(resp)
}

// WriteText serialises a sample of each metric in a simple text
// format. All errors returned by Writer are ignored by design.
func WriteText(w io.Writer) {
	std.WriteText(w)
}

// WriteText serialises a sample of each metric in a simple text
// format. All errors returned by Writer are ignored by design.
func (reg *Register) WriteText(w io.Writer) {
	// write buffer
	buf := make([]byte, len(headerLine), 4096)
	copy(buf, headerLine)

	var buckets []uint64 // reusable

	// snapshot
	reg.mutex.RLock()
	defer reg.mutex.RUnlock()

	// reuse to minimise time lookups
	lineEnd := sampleLineEnd(make([]byte, 21))

	// serialise samples in order of appearance
	for _, m := range reg.metrics {
		if cap(buf)-len(buf) < len(m.typeComment)+len(m.helpComment) {
			w.Write(buf)
			buf = buf[:0]
			// need fresh timestamp after Write
			lineEnd = sampleLineEnd(lineEnd)
		}

		buf = append(buf, m.typeComment...)
		buf = append(buf, m.helpComment...)

		switch m.typeID {
		case counterID:
			if m.counter != nil {
				buf, lineEnd = m.counter.sample(w, buf, lineEnd)
			}

			for _, l := range m.labels {
				l.Lock()
				view := l.counters
				l.Unlock()
				for _, c := range view {
					buf, lineEnd = c.sample(w, buf, lineEnd)
				}
			}

		case integerID:
			if m.integer != nil {
				buf, lineEnd = m.integer.sample(w, buf, lineEnd)
			}

			for _, l := range m.labels {
				l.Lock()
				view := l.integers
				l.Unlock()
				for _, g := range view {
					buf, lineEnd = g.sample(w, buf, lineEnd)
				}
			}

		case realID:
			if m.real != nil {
				buf, lineEnd = m.real.sample(w, buf, lineEnd)
			}

			for _, l := range m.labels {
				l.Lock()
				view := l.reals
				l.Unlock()
				for _, g := range view {
					buf, lineEnd = g.sample(w, buf, lineEnd)
				}
			}

		case counterSampleID, realSampleID:
			if m.sample != nil {
				buf, lineEnd = m.sample.sample(w, buf, lineEnd)
			}

			for _, l := range m.labels {
				l.Lock()
				view := l.samples
				l.Unlock()
				for _, v := range view {
					buf, lineEnd = v.sample(w, buf, lineEnd)
				}
			}

		case histogramID:
			if m.histogram != nil {
				buf, lineEnd, buckets = m.histogram.sample(w, buf, lineEnd, buckets)
			}

			for _, l := range m.labels {
				l.Lock()
				view := l.histograms
				l.Unlock()
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

func (m *Counter) sample(w io.Writer, buf, lineEnd []byte) ([]byte, []byte) {
	if cap(buf)-len(buf) < len(m.prefix)+maxUint64Text+len(lineEnd) {
		w.Write(buf)
		buf = buf[:0]
		// need fresh timestamp after Write
		lineEnd = sampleLineEnd(lineEnd)
	}

	buf = append(buf, m.prefix...)
	buf = strconv.AppendUint(buf, m.Get(), 10)
	buf = append(buf, lineEnd...)

	return buf, lineEnd
}

func (m *Integer) sample(w io.Writer, buf, lineEnd []byte) ([]byte, []byte) {
	if cap(buf)-len(buf) < len(m.prefix)+maxInt64Text+len(lineEnd) {
		w.Write(buf)
		buf = buf[:0]
		// need fresh timestamp after Write
		lineEnd = sampleLineEnd(lineEnd)
	}

	buf = append(buf, m.prefix...)
	buf = strconv.AppendInt(buf, m.Get(), 10)
	buf = append(buf, lineEnd...)

	return buf, lineEnd
}

func (m *Real) sample(w io.Writer, buf, lineEnd []byte) ([]byte, []byte) {
	if cap(buf)-len(buf) < len(m.prefix)+maxFloat64Text+len(lineEnd) {
		w.Write(buf)
		buf = buf[:0]
		// need fresh timestamp after Write
		lineEnd = sampleLineEnd(lineEnd)
	}

	buf = append(buf, m.prefix...)
	buf = strconv.AppendFloat(buf, m.Get(), 'g', -1, 64)
	buf = append(buf, lineEnd...)

	return buf, lineEnd
}

func (m *Sample) sample(w io.Writer, buf, lineEnd []byte) ([]byte, []byte) {
	if cap(buf)-len(buf) < len(m.prefix)+maxFloat64Text+21 {
		w.Write(buf)
		buf = buf[:0]
		// need fresh timestamp after Write
		lineEnd = sampleLineEnd(lineEnd)
	}

	if value, timestamp := m.Get(); timestamp != 0 {
		buf = append(buf, m.prefix...)
		buf = strconv.AppendFloat(buf, value, 'g', -1, 64)
		if !SkipTimestamp {
			buf = append(buf, ' ')
			buf = strconv.AppendUint(buf, timestamp, 10)
		}
		buf = append(buf, '\n')
	}

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

// SampleLineEnd may include a timestamp, and terminates with a line feed.
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
