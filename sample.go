package metrics

import (
	"math"
	"net/http"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"
)

const headerLine = "# Prometheus Samples\n"

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
	buf := make([]byte, len(headerLine), 4096)
	copy(buf, headerLine)

	// snapshot
	mutex.Lock()
	view := metrics[:]
	mutex.Unlock()

	// reuse to minimise time lookups
	lineEnd := sampleLineEnd(make([]byte, 21))

	// serialise samples in order of appearance
	for _, m := range view {
		if cap(buf)-len(buf) < 1+len(m.typeComment)+len(m.helpComment) {
			w.Write(buf)
			buf = buf[:0]
			// need fresh timestamp after Write
			lineEnd = sampleLineEnd(lineEnd)
		}

		buf = append(buf, '\n')
		buf = append(buf, m.typeComment...)
		buf = append(buf, m.helpComment...)

		switch m.typeID {
		case gaugeType:
			var sample bool

			if m.gauge != nil {
				buf, lineEnd = m.gauge.sample(w, buf, lineEnd)
				sample = true
			}

			for _, l1 := range m.gaugeL1s {
				for _, g := range l1.gauges {
					buf, lineEnd = g.sample(w, buf, lineEnd)
					sample = true
				}
			}
			for _, l2 := range m.gaugeL2s {
				for _, g := range l2.gauges {
					buf, lineEnd = g.sample(w, buf, lineEnd)
					sample = true
				}
			}
			for _, l3 := range m.gaugeL3s {
				for _, g := range l3.gauges {
					buf, lineEnd = g.sample(w, buf, lineEnd)
					sample = true
				}
			}

			if !sample && m.copyFallback != nil {
				buf, lineEnd = m.copyFallback.sample(w, buf, lineEnd)
			}

		case counterType:
			if m.counter != nil {
				buf, lineEnd = m.counter.sample(w, buf, lineEnd)
			} else if m.copyFallback != nil {
				buf, lineEnd = m.copyFallback.sample(w, buf, lineEnd)
			}

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

func (c *Copy) sample(w http.ResponseWriter, buf, lineEnd []byte) ([]byte, []byte) {
	if cap(buf)-len(buf) < len(c.prefix)+maxFloat64Text+21 {
		w.Write(buf)
		buf = buf[:0]
		// need fresh timestamp after Write
		lineEnd = sampleLineEnd(lineEnd)
	}

	if value, timestamp := c.Get(); timestamp != 0 {
		buf = append(buf, c.prefix...)
		buf = strconv.AppendFloat(buf, value, 'g', -1, 64)
		if !SkipTimestamp {
			buf = append(buf, ' ')
			buf = strconv.AppendUint(buf, timestamp, 10)
		}
		buf = append(buf, '\n')
	}

	return buf, lineEnd
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
