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

func (mapping *labelMapping) counter12(value1, value2 string) *Counter {
	i := mapping.lockIndex12(value1, value2)
	if i < len(mapping.counters) {
		mapping.Unlock()
		return mapping.counters[i]
	}

	m := &Counter{prefix: mapping.format2LabelPrefix(value1, value2)}
	mapping.counters = append(mapping.counters, m)
	mapping.Unlock()
	return m
}

func (mapping *labelMapping) counter123(value1, value2, value3 string) *Counter {
	i := mapping.lockIndex123(value1, value2, value3)
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

func (mapping *labelMapping) integer12(value1, value2 string) *Integer {
	i := mapping.lockIndex12(value1, value2)
	if i < len(mapping.integers) {
		mapping.Unlock()
		return mapping.integers[i]
	}

	m := &Integer{prefix: mapping.format2LabelPrefix(value1, value2)}
	mapping.integers = append(mapping.integers, m)
	mapping.Unlock()
	return m
}

func (mapping *labelMapping) integer123(value1, value2, value3 string) *Integer {
	i := mapping.lockIndex123(value1, value2, value3)
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

func (mapping *labelMapping) real12(value1, value2 string) *Real {
	i := mapping.lockIndex12(value1, value2)
	if i < len(mapping.reals) {
		mapping.Unlock()
		return mapping.reals[i]
	}

	m := &Real{prefix: mapping.format2LabelPrefix(value1, value2)}
	mapping.reals = append(mapping.reals, m)
	mapping.Unlock()
	return m
}

func (mapping *labelMapping) real123(value1, value2, value3 string) *Real {
	i := mapping.lockIndex123(value1, value2, value3)
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

func (mapping *labelMapping) sample12(value1, value2 string) *Sample {
	i := mapping.lockIndex12(value1, value2)
	if i < len(mapping.samples) {
		mapping.Unlock()
		return mapping.samples[i]
	}

	m := &Sample{prefix: mapping.format2LabelPrefix(value1, value2)}
	mapping.samples = append(mapping.samples, m)
	mapping.Unlock()
	return m
}

func (mapping *labelMapping) sample123(value1, value2, value3 string) *Sample {
	i := mapping.lockIndex123(value1, value2, value3)
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

func (mapping *labelMapping) histogram12(value1, value2 string) *Histogram {
	i := mapping.lockIndex12(value1, value2)
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

func (mapping *labelMapping) lockIndex12(value1, value2 string) int {
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

func (mapping *labelMapping) lockIndex123(value1, value2, value3 string) int {
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

// Labels values may have any [!] byte content, i.e., there is no illegal value.
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

// ParseMetricLabels returns a new map if s has labels.
func parseMetricLabels(s string) map[string]string {
	i := strings.IndexByte(s, '{')
	if i < 0 {
		return nil
	}
	s = s[i+1:]
	labels := make(map[string]string, 3)

	for {
		name := s[:strings.IndexByte(s, '=')]
		s = s[len(name)+2:] // skips double-quote check

		end := strings.IndexAny(s, `"\`)
		if s[end] == '"' {
			labels[name] = s[:end]
		} else {
			var buf strings.Builder
			for {
				buf.WriteString(s[:end])
				if s[end] == '"' {
					break
				}

				if s[end+1] == 'n' {
					buf.WriteByte('\n')
					s = s[end+2:]
					end = strings.IndexAny(s, `"\`)
				} else {
					s = s[end+1:]
					end = 1 + strings.IndexAny(s[1:], `"\`)
				}
			}
			labels[name] = buf.String()
		}

		if s[end+1] == '}' {
			return labels
		}
		s = s[end+2:] // skips comma check
	}
}

/////////////////////////
// Label Order Remappings

func (mapping *labelMapping) counter21(v2, v1 string) *Counter { return mapping.counter12(v1, v2) }
func (mapping *labelMapping) counter132(v1, v3, v2 string) *Counter {
	return mapping.counter123(v1, v2, v3)
}
func (mapping *labelMapping) counter213(v2, v1, v3 string) *Counter {
	return mapping.counter123(v1, v2, v3)
}
func (mapping *labelMapping) counter231(v2, v3, v1 string) *Counter {
	return mapping.counter123(v1, v2, v3)
}
func (mapping *labelMapping) counter312(v3, v1, v2 string) *Counter {
	return mapping.counter123(v1, v2, v3)
}
func (mapping *labelMapping) counter321(v3, v2, v1 string) *Counter {
	return mapping.counter123(v1, v2, v3)
}

func (mapping *labelMapping) integer21(v2, v1 string) *Integer { return mapping.integer12(v1, v2) }
func (mapping *labelMapping) integer132(v1, v3, v2 string) *Integer {
	return mapping.integer123(v1, v2, v3)
}
func (mapping *labelMapping) integer213(v2, v1, v3 string) *Integer {
	return mapping.integer123(v1, v2, v3)
}
func (mapping *labelMapping) integer231(v2, v3, v1 string) *Integer {
	return mapping.integer123(v1, v2, v3)
}
func (mapping *labelMapping) integer312(v3, v1, v2 string) *Integer {
	return mapping.integer123(v1, v2, v3)
}
func (mapping *labelMapping) integer321(v3, v2, v1 string) *Integer {
	return mapping.integer123(v1, v2, v3)
}

func (mapping *labelMapping) real21(v2, v1 string) *Real      { return mapping.real12(v1, v2) }
func (mapping *labelMapping) real132(v1, v3, v2 string) *Real { return mapping.real123(v1, v2, v3) }
func (mapping *labelMapping) real213(v2, v1, v3 string) *Real { return mapping.real123(v1, v2, v3) }
func (mapping *labelMapping) real231(v2, v3, v1 string) *Real { return mapping.real123(v1, v2, v3) }
func (mapping *labelMapping) real312(v3, v1, v2 string) *Real { return mapping.real123(v1, v2, v3) }
func (mapping *labelMapping) real321(v3, v2, v1 string) *Real { return mapping.real123(v1, v2, v3) }

func (mapping *labelMapping) sample21(v2, v1 string) *Sample { return mapping.sample12(v1, v2) }
func (mapping *labelMapping) sample132(v1, v3, v2 string) *Sample {
	return mapping.sample123(v1, v2, v3)
}
func (mapping *labelMapping) sample213(v2, v1, v3 string) *Sample {
	return mapping.sample123(v1, v2, v3)
}
func (mapping *labelMapping) sample231(v2, v3, v1 string) *Sample {
	return mapping.sample123(v1, v2, v3)
}
func (mapping *labelMapping) sample312(v3, v1, v2 string) *Sample {
	return mapping.sample123(v1, v2, v3)
}
func (mapping *labelMapping) sample321(v3, v2, v1 string) *Sample {
	return mapping.sample123(v1, v2, v3)
}

func (mapping *labelMapping) histogram21(v2, v1 string) *Histogram {
	return mapping.histogram12(v1, v2)
}
