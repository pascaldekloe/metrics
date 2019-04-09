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
	samples    []*Sample
	histograms []*Histogram

	buckets []float64
}

func (mapping *labelMapping) counter1(value string) *Counter {
	i := mapping.lockIndex1(value)
	if i < len(mapping.counters) {
		mapping.Unlock()
		return mapping.counters[i]
	}

	m := &Counter{prefix: mapping.format1LabelPrefix(value)}
	mapping.counters = append(mapping.counters, m)
	mapping.Unlock()
	return m
}

func (mapping *labelMapping) counter2(value1, value2 string) *Counter {
	i := mapping.lockIndex2(value1, value2)
	if i < len(mapping.counters) {
		mapping.Unlock()
		return mapping.counters[i]
	}

	m := &Counter{prefix: mapping.format2LabelPrefix(value1, value2)}
	mapping.counters = append(mapping.counters, m)
	mapping.Unlock()
	return m
}

func (mapping *labelMapping) counter3(value1, value2, value3 string) *Counter {
	i := mapping.lockIndex3(value1, value2, value3)
	if i < len(mapping.counters) {
		mapping.Unlock()
		return mapping.counters[i]
	}

	m := &Counter{prefix: mapping.format3LabelPrefix(value1, value2, value3)}
	mapping.counters = append(mapping.counters, m)
	mapping.Unlock()
	return m
}

func (mapping *labelMapping) integer1(value string) *Integer {
	i := mapping.lockIndex1(value)
	if i < len(mapping.integers) {
		mapping.Unlock()
		return mapping.integers[i]
	}

	m := &Integer{prefix: mapping.format1LabelPrefix(value)}
	mapping.integers = append(mapping.integers, m)
	mapping.Unlock()
	return m
}

func (mapping *labelMapping) integer2(value1, value2 string) *Integer {
	i := mapping.lockIndex2(value1, value2)
	if i < len(mapping.integers) {
		mapping.Unlock()
		return mapping.integers[i]
	}

	m := &Integer{prefix: mapping.format2LabelPrefix(value1, value2)}
	mapping.integers = append(mapping.integers, m)
	mapping.Unlock()
	return m
}

func (mapping *labelMapping) integer3(value1, value2, value3 string) *Integer {
	i := mapping.lockIndex3(value1, value2, value3)
	if i < len(mapping.integers) {
		mapping.Unlock()
		return mapping.integers[i]
	}

	m := &Integer{prefix: mapping.format3LabelPrefix(value1, value2, value3)}
	mapping.integers = append(mapping.integers, m)
	mapping.Unlock()
	return m
}

func (mapping *labelMapping) real1(value string) *Real {
	i := mapping.lockIndex1(value)
	if i < len(mapping.reals) {
		mapping.Unlock()
		return mapping.reals[i]
	}

	m := &Real{prefix: mapping.format1LabelPrefix(value)}
	mapping.reals = append(mapping.reals, m)
	mapping.Unlock()
	return m
}

func (mapping *labelMapping) real2(value1, value2 string) *Real {
	i := mapping.lockIndex2(value1, value2)
	if i < len(mapping.reals) {
		mapping.Unlock()
		return mapping.reals[i]
	}

	m := &Real{prefix: mapping.format2LabelPrefix(value1, value2)}
	mapping.reals = append(mapping.reals, m)
	mapping.Unlock()
	return m
}

func (mapping *labelMapping) real3(value1, value2, value3 string) *Real {
	i := mapping.lockIndex3(value1, value2, value3)
	if i < len(mapping.reals) {
		mapping.Unlock()
		return mapping.reals[i]
	}

	m := &Real{prefix: mapping.format3LabelPrefix(value1, value2, value3)}
	mapping.reals = append(mapping.reals, m)
	mapping.Unlock()
	return m
}

func (mapping *labelMapping) sample1(value string) *Sample {
	i := mapping.lockIndex1(value)
	if i < len(mapping.samples) {
		mapping.Unlock()
		return mapping.samples[i]
	}

	m := &Sample{prefix: mapping.format1LabelPrefix(value)}
	mapping.samples = append(mapping.samples, m)
	mapping.Unlock()
	return m
}

func (mapping *labelMapping) sample2(value1, value2 string) *Sample {
	i := mapping.lockIndex2(value1, value2)
	if i < len(mapping.samples) {
		mapping.Unlock()
		return mapping.samples[i]
	}

	m := &Sample{prefix: mapping.format2LabelPrefix(value1, value2)}
	mapping.samples = append(mapping.samples, m)
	mapping.Unlock()
	return m
}

func (mapping *labelMapping) sample3(value1, value2, value3 string) *Sample {
	i := mapping.lockIndex3(value1, value2, value3)
	if i < len(mapping.samples) {
		mapping.Unlock()
		return mapping.samples[i]
	}

	m := &Sample{prefix: mapping.format3LabelPrefix(value1, value2, value3)}
	mapping.samples = append(mapping.samples, m)
	mapping.Unlock()
	return m
}

func (mapping *labelMapping) histogram1(value string) *Histogram {
	i := mapping.lockIndex1(value)
	if i < len(mapping.histograms) {
		mapping.Unlock()
		return mapping.histograms[i]
	}

	h := newHistogram(mapping.name, mapping.buckets)

	// set prefixes
	tail := `",` + mapping.labelNames[0] + `="` + valueEscapes.Replace(value) + `"} `
	for i, f := range h.BucketBounds {
		h.bucketPrefixes[i] = mapping.name + `{le="` + strconv.FormatFloat(f, 'g', -1, 64) + tail
	}
	h.bucketPrefixes[len(h.BucketBounds)] = mapping.name + `{le="+Inf` + tail
	h.countPrefix = mapping.name + "_count{" + tail[2:]
	h.sumPrefix = mapping.name + "_sum{" + tail[2:]

	mapping.histograms = append(mapping.histograms, h)

	mapping.Unlock()
	return h
}

func (mapping *labelMapping) histogram2(value1, value2 string) *Histogram {
	i := mapping.lockIndex2(value1, value2)
	if i < len(mapping.histograms) {
		mapping.Unlock()
		return mapping.histograms[i]
	}

	h := newHistogram(mapping.name, mapping.buckets)

	// set prefixes
	tail := `",` + mapping.labelNames[0] + `="` + valueEscapes.Replace(value1)
	tail += `",` + mapping.labelNames[1] + `="` + valueEscapes.Replace(value2) + `"} `
	for i, f := range h.BucketBounds {
		h.bucketPrefixes[i] = mapping.name + `{le="` + strconv.FormatFloat(f, 'g', -1, 64) + tail
	}
	h.bucketPrefixes[len(h.BucketBounds)] = mapping.name + `{le="+Inf` + tail
	h.countPrefix = mapping.name + "_count{" + tail[2:]
	h.sumPrefix = mapping.name + "_sum{" + tail[2:]

	mapping.histograms = append(mapping.histograms, h)

	mapping.Unlock()
	return h
}

// 64-Bit FNV
const (
	hashOffset = 14695981039346656037
	hashPrime  = 1099511628211
)

func (mapping *labelMapping) lockIndex1(value string) int {
	hash := uint64(hashOffset)
	hash ^= uint64(len(value))
	hash *= hashPrime
	for i := 0; i < len(value); i++ {
		hash ^= uint64(value[i])
		hash *= hashPrime
	}

	return mapping.lockIndex(hash)
}

func (mapping *labelMapping) lockIndex2(value1, value2 string) int {
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

	return mapping.lockIndex(hash)
}

func (mapping *labelMapping) lockIndex3(value1, value2, value3 string) int {
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

	return mapping.lockIndex(hash)
}

func (mapping *labelMapping) lockIndex(hash uint64) int {
	mapping.Lock()

	for i, h := range mapping.labelHashes {
		if h == hash {
			return i
		}
	}

	i := len(mapping.labelHashes)
	mapping.labelHashes = append(mapping.labelHashes, hash)
	return i
}

var valueEscapes = strings.NewReplacer("\n", `\n`, `"`, `\"`, `\`, `\\`)

func (mapping *labelMapping) format1LabelPrefix(labelValue string) string {
	var buf strings.Builder
	buf.Grow(6 + len(mapping.name) + len(mapping.labelNames[0]) + len(labelValue))

	buf.WriteString(mapping.name)
	buf.WriteByte('{')
	buf.WriteString(mapping.labelNames[0])
	buf.WriteString(`="`)
	valueEscapes.WriteString(&buf, labelValue)
	buf.WriteString(`"} `)

	return buf.String()
}

func (mapping *labelMapping) format2LabelPrefix(labelValue1, labelValue2 string) string {
	var buf strings.Builder
	buf.Grow(10 + len(mapping.name) + len(mapping.labelNames[0]) + len(mapping.labelNames[1]) + len(labelValue1) + len(labelValue2))

	buf.WriteString(mapping.name)
	buf.WriteByte('{')
	buf.WriteString(mapping.labelNames[0])
	buf.WriteString(`="`)
	valueEscapes.WriteString(&buf, labelValue1)
	buf.WriteString(`",`)
	buf.WriteString(mapping.labelNames[1])
	buf.WriteString(`="`)
	valueEscapes.WriteString(&buf, labelValue2)
	buf.WriteString(`"} `)

	return buf.String()
}

func (mapping *labelMapping) format3LabelPrefix(labelValue1, labelValue2, labelValue3 string) string {
	var buf strings.Builder
	buf.Grow(14 + len(mapping.name) + len(mapping.labelNames[0]) + len(mapping.labelNames[1]) + len(mapping.labelNames[2]) + len(labelValue1) + len(labelValue2) + len(labelValue3))

	buf.WriteString(mapping.name)
	buf.WriteByte('{')
	buf.WriteString(mapping.labelNames[0])
	buf.WriteString(`="`)
	valueEscapes.WriteString(&buf, labelValue1)
	buf.WriteString(`",`)
	buf.WriteString(mapping.labelNames[1])
	buf.WriteString(`="`)
	valueEscapes.WriteString(&buf, labelValue2)
	buf.WriteString(`",`)
	buf.WriteString(mapping.labelNames[2])
	buf.WriteString(`="`)
	valueEscapes.WriteString(&buf, labelValue3)
	buf.WriteString(`"} `)

	return buf.String()
}
