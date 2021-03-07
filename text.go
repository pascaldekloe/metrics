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

// ServeHTTP provides a sample of each metric as an http.HandlerFunc.
func ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	std.ServeHTTP(resp, req)
}

// ServeHTTP provides a sample of each metric as an http.Handler.
func (reg *Register) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet && req.Method != http.MethodHead {
		resp.Header().Set("Allow", http.MethodOptions+", "+http.MethodGet+", "+http.MethodHead)
		if req.Method != http.MethodOptions {
			http.Error(resp, "read-only resource", http.StatusMethodNotAllowed)
		}

		return
	}

	resp.Header().Set("Content-Type", "text/plain;version=0.0.4;charset=utf-8")
	reg.WriteTo(resp)
}

// WriteText serialises a sample of each metric in a simple text
// format. All errors returned by Writer are ignored by design.
// Deprecated: Use WriteTo instead.
func WriteText(w io.Writer) {
	std.WriteText(w)
}

// WriteText serialises a sample of each metric in a simple text
// format. All errors returned by Writer are ignored by design.
// Deprecated: Use WriteTo instead.
func (reg *Register) WriteText(w io.Writer) {
	reg.WriteTo(w)
}

// WriteTo serialises a sample of each metric in a simple text
// format as an io.WriterTo.
func WriteTo(w io.Writer) (n int64, err error) {
	return std.WriteTo(w)
}

// WriteTo serialises a sample of each metric in a simple text
// format as an io.WriterTo.
func (reg *Register) WriteTo(w io.Writer) (n int64, err error) {
	wn, err := io.WriteString(w, headerLine)
	n = int64(wn)
	if err != nil {
		return n, err
	}

	// resables
	var buf []byte
	var buckets []uint64

	// snapshot
	reg.mutex.RLock()
	defer reg.mutex.RUnlock()

	// serialise samples in order of appearance
	for _, m := range reg.metrics {
		buf = append(buf, m.typeComment...)
		buf = append(buf, m.helpComment...)

		switch m.typeID {
		case counterID:
			if m.counter != nil {
				buf = append(buf, m.counter.prefix...)
				buf = strconv.AppendUint(buf, m.counter.Get(), 10)
				buf = appendTimestamp(buf)
			}

			for _, l := range m.labels {
				l.Lock()
				view := l.counters
				l.Unlock()
				for _, v := range view {
					buf = append(buf, v.prefix...)
					buf = strconv.AppendUint(buf, v.Get(), 10)
					buf = appendTimestamp(buf)
				}
			}

		case integerID:
			if m.integer != nil {
				buf = append(buf, m.integer.prefix...)
				buf = strconv.AppendInt(buf, m.integer.Get(), 10)
				buf = appendTimestamp(buf)
			}

			for _, l := range m.labels {
				l.Lock()
				view := l.integers
				l.Unlock()
				for _, v := range view {
					buf = append(buf, v.prefix...)
					buf = strconv.AppendInt(buf, v.Get(), 10)
					buf = appendTimestamp(buf)
				}
			}

		case realID:
			if m.real != nil {
				buf = append(buf, m.real.prefix...)
				buf = strconv.AppendFloat(buf, m.real.Get(), 'g', -1, 64)
				buf = appendTimestamp(buf)
			}

			for _, l := range m.labels {
				l.Lock()
				view := l.reals
				l.Unlock()
				for _, v := range view {
					buf = append(buf, v.prefix...)
					buf = strconv.AppendFloat(buf, v.Get(), 'g', -1, 64)
					buf = appendTimestamp(buf)
				}
			}

		case counterSampleID, realSampleID:
			if m.sample != nil {
				buf = m.sample.append(buf)
			}

			for _, l := range m.labels {
				l.Lock()
				view := l.samples
				l.Unlock()
				for _, v := range view {
					buf = v.append(buf)
				}
			}

		case histogramID:
			if m.histogram != nil {
				buf = m.histogram.append(buf, &buckets)
			}

			for _, l := range m.labels {
				l.Lock()
				view := l.histograms
				l.Unlock()
				for _, v := range view {
					buf = v.append(buf, &buckets)
				}
			}
		}

		wn, err = w.Write(buf)
		n += int64(wn)
		if err != nil {
			return n, err
		}
		buf = buf[:0]
	}

	return n, nil
}

func (m *Sample) append(buf []byte) []byte {
	if value, timestamp := m.Get(); timestamp != 0 {
		buf = append(buf, m.prefix...)
		buf = strconv.AppendFloat(buf, value, 'g', -1, 64)
		if !SkipTimestamp {
			buf = append(buf, ' ')
			buf = strconv.AppendUint(buf, timestamp, 10)
		}
		buf = append(buf, '\n')
	}
	return buf
}

func (h *Histogram) append(buf []byte, buckets *[]uint64) []byte {
	var count uint64
	var sum float64
	*buckets, count, sum = h.Get((*buckets)[:0])

	buf = append(buf, h.countPrefix...)
	offset := len(buf)
	buf = strconv.AppendUint(buf, count, 10)
	countSerial := buf[offset:]

	timeOffset := len(buf)
	buf = appendTimestamp(buf)
	timestamp := buf[timeOffset:]

	// buckets
	var cum uint64
	for i, prefix := range h.bucketPrefixes {
		if i >= len(*buckets) {
			// (redundant) +Inf bucket
			buf = append(buf, prefix...)
			buf = append(buf, countSerial...)
			buf = append(buf, timestamp...)
			break
		}

		cum += (*buckets)[i]
		buf = append(buf, prefix...)
		buf = strconv.AppendUint(buf, cum, 10)
		buf = append(buf, timestamp...)
	}

	// sum
	buf = append(buf, h.sumPrefix...)
	buf = strconv.AppendFloat(buf, sum, 'g', -1, 64)
	buf = append(buf, timestamp...)

	return buf
}

func appendTimestamp(buf []byte) []byte {
	if !SkipTimestamp {
		buf = append(buf, ' ')
		ms := time.Now().UnixNano() / 1e6
		buf = strconv.AppendInt(buf, ms, 10)
	}

	buf = append(buf, '\n')
	return buf
}
