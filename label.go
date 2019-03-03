package metrics

import (
	"strconv"
	"strings"
	"sync"
)

type labelMapping struct {
	sync.Mutex
	name        string
	labelNames  [3]string
	labelHashes []uint64

	counters   []*Counter
	integers   []*Integer
	reals      []*Real
	buckets    []float64
	histograms []*Histogram
	samples    []*Sample
}

func (m *labelMapping) counter1(value string) *Counter {
	i := m.lockIndex1(value)
	if i < len(m.counters) {
		m.Unlock()
		return m.counters[i]
	}

	c := &Counter{prefix: m.format1LabelPrefix(value)}
	m.counters = append(m.counters, c)
	m.Unlock()
	return c
}

func (m *labelMapping) counter2(value1, value2 string) *Counter {
	i := m.lockIndex2(value1, value2)
	if i < len(m.counters) {
		m.Unlock()
		return m.counters[i]
	}

	c := &Counter{prefix: m.format2LabelPrefix(value1, value2)}
	m.counters = append(m.counters, c)
	m.Unlock()
	return c
}

func (m *labelMapping) counter3(value1, value2, value3 string) *Counter {
	i := m.lockIndex3(value1, value2, value3)
	if i < len(m.counters) {
		m.Unlock()
		return m.counters[i]
	}

	c := &Counter{prefix: m.format3LabelPrefix(value1, value2, value3)}
	m.counters = append(m.counters, c)
	m.Unlock()
	return c
}

func (m *labelMapping) integer1(value string) *Integer {
	i := m.lockIndex1(value)
	if i < len(m.integers) {
		m.Unlock()
		return m.integers[i]
	}

	z := &Integer{prefix: m.format1LabelPrefix(value)}
	m.integers = append(m.integers, z)
	m.Unlock()
	return z
}

func (m *labelMapping) integer2(value1, value2 string) *Integer {
	i := m.lockIndex2(value1, value2)
	if i < len(m.integers) {
		m.Unlock()
		return m.integers[i]
	}

	z := &Integer{prefix: m.format2LabelPrefix(value1, value2)}
	m.integers = append(m.integers, z)
	m.Unlock()
	return z
}

func (m *labelMapping) integer3(value1, value2, value3 string) *Integer {
	i := m.lockIndex3(value1, value2, value3)
	if i < len(m.integers) {
		m.Unlock()
		return m.integers[i]
	}

	z := &Integer{prefix: m.format3LabelPrefix(value1, value2, value3)}
	m.integers = append(m.integers, z)
	m.Unlock()
	return z
}

func (m *labelMapping) real1(value string) *Real {
	i := m.lockIndex1(value)
	if i < len(m.reals) {
		m.Unlock()
		return m.reals[i]
	}

	r := &Real{prefix: m.format1LabelPrefix(value)}
	m.reals = append(m.reals, r)
	m.Unlock()
	return r
}

func (m *labelMapping) real2(value1, value2 string) *Real {
	i := m.lockIndex2(value1, value2)
	if i < len(m.reals) {
		m.Unlock()
		return m.reals[i]
	}

	r := &Real{prefix: m.format2LabelPrefix(value1, value2)}
	m.reals = append(m.reals, r)
	m.Unlock()
	return r
}

func (m *labelMapping) real3(value1, value2, value3 string) *Real {
	i := m.lockIndex3(value1, value2, value3)
	if i < len(m.reals) {
		m.Unlock()
		return m.reals[i]
	}

	r := &Real{prefix: m.format3LabelPrefix(value1, value2, value3)}
	m.reals = append(m.reals, r)
	m.Unlock()
	return r
}

func (m *labelMapping) sample1(value string) *Sample {
	i := m.lockIndex1(value)
	if i < len(m.samples) {
		m.Unlock()
		return m.samples[i]
	}

	r := &Sample{prefix: m.format1LabelPrefix(value)}
	m.samples = append(m.samples, r)
	m.Unlock()
	return r
}

func (m *labelMapping) sample2(value1, value2 string) *Sample {
	i := m.lockIndex2(value1, value2)
	if i < len(m.samples) {
		m.Unlock()
		return m.samples[i]
	}

	r := &Sample{prefix: m.format2LabelPrefix(value1, value2)}
	m.samples = append(m.samples, r)
	m.Unlock()
	return r
}

func (m *labelMapping) sample3(value1, value2, value3 string) *Sample {
	i := m.lockIndex3(value1, value2, value3)
	if i < len(m.samples) {
		m.Unlock()
		return m.samples[i]
	}

	r := &Sample{prefix: m.format3LabelPrefix(value1, value2, value3)}
	m.samples = append(m.samples, r)
	m.Unlock()
	return r
}

func (m *labelMapping) histogram1(value string) *Histogram {
	i := m.lockIndex1(value)
	if i < len(m.histograms) {
		m.Unlock()
		return m.histograms[i]
	}

	h := newHistogram(m.name, m.buckets)

	// set prefixes
	tail := `",` + m.labelNames[0] + `="` + valueEscapes.Replace(value) + `"} `
	for i, f := range h.BucketBounds {
		h.bucketPrefixes[i] = m.name + `{le="` + strconv.FormatFloat(f, 'g', -1, 64) + tail
	}
	h.bucketPrefixes[len(h.BucketBounds)] = m.name + `{le="+Inf` + tail
	h.countPrefix = m.name + "_count{" + tail[2:]
	h.sumPrefix = m.name + "_sum{" + tail[2:]

	m.histograms = append(m.histograms, h)

	m.Unlock()
	return h
}

func (m *labelMapping) histogram2(value1, value2 string) *Histogram {
	i := m.lockIndex2(value1, value2)
	if i < len(m.histograms) {
		m.Unlock()
		return m.histograms[i]
	}

	h := newHistogram(m.name, m.buckets)

	// set prefixes
	tail := `",` + m.labelNames[0] + `="` + valueEscapes.Replace(value1)
	tail += `",` + m.labelNames[1] + `="` + valueEscapes.Replace(value2) + `"} `
	for i, f := range h.BucketBounds {
		h.bucketPrefixes[i] = m.name + `{le="` + strconv.FormatFloat(f, 'g', -1, 64) + tail
	}
	h.bucketPrefixes[len(h.BucketBounds)] = m.name + `{le="+Inf` + tail
	h.countPrefix = m.name + "_count{" + tail[2:]
	h.sumPrefix = m.name + "_sum{" + tail[2:]

	m.histograms = append(m.histograms, h)

	m.Unlock()
	return h
}

// FNV-1a
const (
	hashOffset = 14695981039346656037
	hashPrime  = 1099511628211
)

func (m *labelMapping) lockIndex1(value string) int {
	hash := uint64(hashOffset)
	hash ^= uint64(len(value))
	hash *= hashPrime
	for i := 0; i < len(value); i++ {
		hash ^= uint64(value[i])
		hash *= hashPrime
	}

	return m.lockIndex(hash)
}

func (m *labelMapping) lockIndex2(value1, value2 string) int {
	hash := uint64(hashOffset)
	hash ^= uint64(len(value1))
	hash *= hashPrime
	hash ^= uint64(len(value2))
	hash *= hashPrime
	for i := 0; i < len(value1); i++ {
		hash ^= uint64(value1[i])
		hash *= hashPrime
	}
	for i := 0; i < len(value2); i++ {
		hash ^= uint64(value2[i])
		hash *= hashPrime
	}

	return m.lockIndex(hash)
}

func (m *labelMapping) lockIndex3(value1, value2, value3 string) int {
	hash := uint64(hashOffset)
	hash ^= uint64(len(value1))
	hash *= hashPrime
	hash ^= uint64(len(value2))
	hash *= hashPrime
	hash ^= uint64(len(value3))
	hash *= hashPrime
	for i := 0; i < len(value1); i++ {
		hash ^= uint64(value1[i])
		hash *= hashPrime
	}
	for i := 0; i < len(value2); i++ {
		hash ^= uint64(value2[i])
		hash *= hashPrime
	}
	for i := 0; i < len(value3); i++ {
		hash ^= uint64(value3[i])
		hash *= hashPrime
	}

	return m.lockIndex(hash)
}

func (m *labelMapping) lockIndex(hash uint64) int {
	m.Lock()

	for i, h := range m.labelHashes {
		if h == hash {
			return i
		}
	}

	i := len(m.labelHashes)
	m.labelHashes = append(m.labelHashes, hash)
	return i
}

var valueEscapes = strings.NewReplacer("\n", `\n`, `"`, `\"`, `\`, `\\`)

func (m *labelMapping) format1LabelPrefix(labelValue string) string {
	var buf strings.Builder
	buf.Grow(6 + len(m.name) + len(m.labelNames[0]) + len(labelValue))

	buf.WriteString(m.name)
	buf.WriteByte('{')
	buf.WriteString(m.labelNames[0])
	buf.WriteString(`="`)
	valueEscapes.WriteString(&buf, labelValue)
	buf.WriteString(`"} `)

	return buf.String()
}

func (m *labelMapping) format2LabelPrefix(labelValue1, labelValue2 string) string {
	var buf strings.Builder
	buf.Grow(10 + len(m.name) + len(m.labelNames[0]) + len(m.labelNames[1]) + len(labelValue1) + len(labelValue2))

	buf.WriteString(m.name)
	buf.WriteByte('{')
	buf.WriteString(m.labelNames[0])
	buf.WriteString(`="`)
	valueEscapes.WriteString(&buf, labelValue1)
	buf.WriteString(`",`)
	buf.WriteString(m.labelNames[1])
	buf.WriteString(`="`)
	valueEscapes.WriteString(&buf, labelValue2)
	buf.WriteString(`"} `)

	return buf.String()
}

func (m *labelMapping) format3LabelPrefix(labelValue1, labelValue2, labelValue3 string) string {
	var buf strings.Builder
	buf.Grow(14 + len(m.name) + len(m.labelNames[0]) + len(m.labelNames[1]) + len(m.labelNames[2]) + len(labelValue1) + len(labelValue2) + len(labelValue3))

	buf.WriteString(m.name)
	buf.WriteByte('{')
	buf.WriteString(m.labelNames[0])
	buf.WriteString(`="`)
	valueEscapes.WriteString(&buf, labelValue1)
	buf.WriteString(`",`)
	buf.WriteString(m.labelNames[1])
	buf.WriteString(`="`)
	valueEscapes.WriteString(&buf, labelValue2)
	buf.WriteString(`",`)
	buf.WriteString(m.labelNames[2])
	buf.WriteString(`="`)
	valueEscapes.WriteString(&buf, labelValue3)
	buf.WriteString(`"} `)

	return buf.String()
}
