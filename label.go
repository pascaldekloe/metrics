package metrics

import (
	"strconv"
	"strings"
	"sync"
)

type map1Label struct {
	mutex       sync.Mutex
	name        string
	labelName   string
	labelHashes []uint64
}

type map2Label struct {
	mutex       sync.Mutex
	name        string
	labelNames  [2]string
	labelHashes []uint64
}

type map3Label struct {
	mutex       sync.Mutex
	name        string
	labelNames  [3]string
	labelHashes []uint64
}

type map1LabelCounter struct {
	map1Label
	counters []*Counter
}

type map2LabelCounter struct {
	map2Label
	counters []*Counter
}

type map3LabelCounter struct {
	map3Label
	counters []*Counter
}

type map1LabelInteger struct {
	map1Label
	integers []*Integer
}

type map2LabelInteger struct {
	map2Label
	integers []*Integer
}

type map3LabelInteger struct {
	map3Label
	integers []*Integer
}

type map1LabelReal struct {
	map1Label
	reals []*Real
}

type map2LabelReal struct {
	map2Label
	reals []*Real
}

type map3LabelReal struct {
	map3Label
	reals []*Real
}

type map1LabelHistogram struct {
	map1Label
	buckets    []float64
	histograms []*Histogram
}

type map2LabelHistogram struct {
	map2Label
	buckets    []float64
	histograms []*Histogram
}

type map1LabelSample struct {
	map1Label
	samples []*Sample
}

type map2LabelSample struct {
	map2Label
	samples []*Sample
}

type map3LabelSample struct {
	map3Label
	samples []*Sample
}

// FNV-1a
const (
	hashOffset = 14695981039346656037
	hashPrime  = 1099511628211
)

func hash1(value string) uint64 {
	hash := uint64(hashOffset)
	hash ^= uint64(len(value))
	hash *= hashPrime
	for i := 0; i < len(value); i++ {
		hash ^= uint64(value[i])
		hash *= hashPrime
	}
	return hash
}

func hash2(value1, value2 string) uint64 {
	hash := hash1(value1)
	hash ^= uint64(len(value2))
	hash *= hashPrime
	for i := 0; i < len(value2); i++ {
		hash ^= uint64(value2[i])
		hash *= hashPrime
	}
	return hash
}

func hash3(value1, value2, value3 string) uint64 {
	hash := hash2(value1, value2)
	hash ^= uint64(len(value3))
	hash *= hashPrime
	for i := 0; i < len(value3); i++ {
		hash ^= uint64(value3[i])
		hash *= hashPrime
	}
	return hash
}

func (l1 *map1LabelCounter) with(value string) *Counter {
	hash := hash1(value)

	l1.mutex.Lock()

	for i, h := range l1.labelHashes {
		if h == hash {
			hit := l1.counters[i]

			l1.mutex.Unlock()
			return hit
		}
	}

	l1.labelHashes = append(l1.labelHashes, hash)
	c := &Counter{prefix: format1LabelPrefix(l1.name, l1.labelName, value)}
	l1.counters = append(l1.counters, c)

	l1.mutex.Unlock()
	return c
}

func (l2 *map2LabelCounter) with(value1, value2 string) *Counter {
	hash := hash2(value1, value2)

	l2.mutex.Lock()

	for i, h := range l2.labelHashes {
		if h == hash {
			hit := l2.counters[i]

			l2.mutex.Unlock()
			return hit
		}
	}

	l2.labelHashes = append(l2.labelHashes, hash)
	c := &Counter{prefix: format2LabelPrefix(l2.name, l2.labelNames[0], value1, l2.labelNames[1], value2)}
	l2.counters = append(l2.counters, c)

	l2.mutex.Unlock()
	return c
}

func (l3 *map3LabelCounter) with(value1, value2, value3 string) *Counter {
	hash := hash3(value1, value2, value3)

	l3.mutex.Lock()

	for i, h := range l3.labelHashes {
		if h == hash {
			hit := l3.counters[i]

			l3.mutex.Unlock()
			return hit
		}
	}

	l3.labelHashes = append(l3.labelHashes, hash)
	c := &Counter{prefix: format3LabelPrefix(l3.name, l3.labelNames[0], value1, l3.labelNames[1], value2, l3.labelNames[2], value3)}
	l3.counters = append(l3.counters, c)

	l3.mutex.Unlock()
	return c
}

func (l1 *map1LabelInteger) with(value string) *Integer {
	hash := hash1(value)

	l1.mutex.Lock()

	for i, h := range l1.labelHashes {
		if h == hash {
			hit := l1.integers[i]

			l1.mutex.Unlock()
			return hit
		}
	}

	l1.labelHashes = append(l1.labelHashes, hash)
	z := &Integer{prefix: format1LabelPrefix(l1.name, l1.labelName, value)}
	l1.integers = append(l1.integers, z)

	l1.mutex.Unlock()
	return z
}

func (l2 *map2LabelInteger) with(value1, value2 string) *Integer {
	hash := hash2(value1, value2)

	l2.mutex.Lock()

	for i, h := range l2.labelHashes {
		if h == hash {
			hit := l2.integers[i]

			l2.mutex.Unlock()
			return hit
		}
	}

	l2.labelHashes = append(l2.labelHashes, hash)
	z := &Integer{prefix: format2LabelPrefix(l2.name, l2.labelNames[0], value1, l2.labelNames[1], value2)}
	l2.integers = append(l2.integers, z)

	l2.mutex.Unlock()
	return z
}

func (l3 *map3LabelInteger) with(value1, value2, value3 string) *Integer {
	hash := hash3(value1, value2, value3)

	l3.mutex.Lock()

	for i, h := range l3.labelHashes {
		if h == hash {
			hit := l3.integers[i]

			l3.mutex.Unlock()
			return hit
		}
	}

	l3.labelHashes = append(l3.labelHashes, hash)
	z := &Integer{prefix: format3LabelPrefix(l3.name, l3.labelNames[0], value1, l3.labelNames[1], value2, l3.labelNames[2], value3)}
	l3.integers = append(l3.integers, z)

	l3.mutex.Unlock()
	return z
}

func (l1 *map1LabelReal) with(value string) *Real {
	hash := hash1(value)

	l1.mutex.Lock()

	for i, h := range l1.labelHashes {
		if h == hash {
			hit := l1.reals[i]

			l1.mutex.Unlock()
			return hit
		}
	}

	l1.labelHashes = append(l1.labelHashes, hash)
	r := &Real{prefix: format1LabelPrefix(l1.name, l1.labelName, value)}
	l1.reals = append(l1.reals, r)

	l1.mutex.Unlock()
	return r
}

func (l2 *map2LabelReal) with(value1, value2 string) *Real {
	hash := hash2(value1, value2)

	l2.mutex.Lock()

	for i, h := range l2.labelHashes {
		if h == hash {
			hit := l2.reals[i]

			l2.mutex.Unlock()
			return hit
		}
	}

	l2.labelHashes = append(l2.labelHashes, hash)
	r := &Real{prefix: format2LabelPrefix(l2.name, l2.labelNames[0], value1, l2.labelNames[1], value2)}
	l2.reals = append(l2.reals, r)

	l2.mutex.Unlock()
	return r
}

func (l3 *map3LabelReal) with(value1, value2, value3 string) *Real {
	hash := hash3(value1, value2, value3)

	l3.mutex.Lock()

	for i, h := range l3.labelHashes {
		if h == hash {
			hit := l3.reals[i]

			l3.mutex.Unlock()
			return hit
		}
	}

	l3.labelHashes = append(l3.labelHashes, hash)
	r := &Real{prefix: format3LabelPrefix(l3.name, l3.labelNames[0], value1, l3.labelNames[1], value2, l3.labelNames[2], value3)}
	l3.reals = append(l3.reals, r)

	l3.mutex.Unlock()
	return r
}

func (l1 *map1LabelHistogram) with(value string) *Histogram {
	hash := hash1(value)

	l1.mutex.Lock()

	for i, h := range l1.labelHashes {
		if h == hash {
			hit := l1.histograms[i]

			l1.mutex.Unlock()
			return hit
		}
	}

	l1.labelHashes = append(l1.labelHashes, hash)
	h := newHistogram(l1.name, l1.buckets)
	l1.histograms = append(l1.histograms, h)

	for i, f := range h.BucketBounds {
		h.bucketPrefixes[i] = format2LabelPrefix(l1.name, "le", strconv.FormatFloat(f, 'g', -1, 64), l1.labelName, value)
	}
	h.bucketPrefixes[len(h.BucketBounds)] = format2LabelPrefix(l1.name, "le", "+Inf", l1.labelName, value)
	h.countPrefix = format1LabelPrefix(l1.name+"_count", l1.labelName, value)
	h.sumPrefix = format1LabelPrefix(l1.name+"_sum", l1.labelName, value)

	l1.mutex.Unlock()
	return h
}

func (l2 *map2LabelHistogram) with(value1, value2 string) *Histogram {
	hash := hash2(value1, value2)

	l2.mutex.Lock()

	for i, h := range l2.labelHashes {
		if h == hash {
			hit := l2.histograms[i]

			l2.mutex.Unlock()
			return hit
		}
	}

	l2.labelHashes = append(l2.labelHashes, hash)
	h := newHistogram(l2.name, l2.buckets)
	l2.histograms = append(l2.histograms, h)

	for i, f := range h.BucketBounds {
		h.bucketPrefixes[i] = format3LabelPrefix(l2.name, "le", strconv.FormatFloat(f, 'g', -1, 64), l2.labelNames[0], value1, l2.labelNames[1], value2)
	}
	h.bucketPrefixes[len(h.BucketBounds)] = format3LabelPrefix(l2.name, "le", "+Inf", l2.labelNames[0], value1, l2.labelNames[1], value2)
	h.countPrefix = format2LabelPrefix(l2.name+"_count", l2.labelNames[0], value1, l2.labelNames[1], value2)
	h.sumPrefix = format2LabelPrefix(l2.name+"_sum", l2.labelNames[0], value1, l2.labelNames[1], value2)

	l2.mutex.Unlock()
	return h
}

func (l1 *map1LabelSample) with(value string) *Sample {
	hash := hash1(value)

	l1.mutex.Lock()

	for i, h := range l1.labelHashes {
		if h == hash {
			hit := l1.samples[i]

			l1.mutex.Unlock()
			return hit
		}
	}

	l1.labelHashes = append(l1.labelHashes, hash)
	s := &Sample{prefix: format1LabelPrefix(l1.name, l1.labelName, value)}
	l1.samples = append(l1.samples, s)

	l1.mutex.Unlock()
	return s
}

func (l2 *map2LabelSample) with(value1, value2 string) *Sample {
	hash := hash2(value1, value2)

	l2.mutex.Lock()

	for i, h := range l2.labelHashes {
		if h == hash {
			hit := l2.samples[i]

			l2.mutex.Unlock()
			return hit
		}
	}

	l2.labelHashes = append(l2.labelHashes, hash)
	s := &Sample{prefix: format2LabelPrefix(l2.name, l2.labelNames[0], value1, l2.labelNames[1], value2)}
	l2.samples = append(l2.samples, s)

	l2.mutex.Unlock()
	return s
}

func (l3 *map3LabelSample) with(value1, value2, value3 string) *Sample {
	hash := hash3(value1, value2, value3)

	l3.mutex.Lock()

	for i, h := range l3.labelHashes {
		if h == hash {
			hit := l3.samples[i]

			l3.mutex.Unlock()
			return hit
		}
	}

	l3.labelHashes = append(l3.labelHashes, hash)
	s := &Sample{prefix: format3LabelPrefix(l3.name, l3.labelNames[0], value1, l3.labelNames[1], value2, l3.labelNames[2], value3)}
	l3.samples = append(l3.samples, s)

	l3.mutex.Unlock()
	return s
}

var valueEscapes = strings.NewReplacer("\n", `\n`, `"`, `\"`, `\`, `\\`)

func format1LabelPrefix(name, labelName, labelValue string) string {
	var buf strings.Builder
	buf.Grow(6 + len(name) + len(labelName) + len(labelValue))

	buf.WriteString(name)
	buf.WriteByte('{')
	buf.WriteString(labelName)
	buf.WriteString(`="`)
	valueEscapes.WriteString(&buf, labelValue)
	buf.WriteString(`"} `)

	return buf.String()
}

func format2LabelPrefix(name string, labelName1, labelValue1, labelName2, labelValue2 string) string {
	var buf strings.Builder
	buf.Grow(10 + len(name) + len(labelName1) + len(labelValue1) + len(labelName2) + len(labelValue2))

	buf.WriteString(name)
	buf.WriteByte('{')
	buf.WriteString(labelName1)
	buf.WriteString(`="`)
	valueEscapes.WriteString(&buf, labelValue1)
	buf.WriteString(`",`)
	buf.WriteString(labelName2)
	buf.WriteString(`="`)
	valueEscapes.WriteString(&buf, labelValue2)
	buf.WriteString(`"} `)

	return buf.String()
}

func format3LabelPrefix(name string, labelName1, labelValue1, labelName2, labelValue2, labelName3, labelValue3 string) string {
	var buf strings.Builder
	buf.Grow(14 + len(name) + len(labelName1) + len(labelValue1) + len(labelName2) + len(labelValue2) + len(labelName3) + len(labelValue3))

	buf.WriteString(name)
	buf.WriteByte('{')
	buf.WriteString(labelName1)
	buf.WriteString(`="`)
	valueEscapes.WriteString(&buf, labelValue1)
	buf.WriteString(`",`)
	buf.WriteString(labelName2)
	buf.WriteString(`="`)
	valueEscapes.WriteString(&buf, labelValue2)
	buf.WriteString(`",`)
	buf.WriteString(labelName3)
	buf.WriteString(`="`)
	valueEscapes.WriteString(&buf, labelValue3)
	buf.WriteString(`"} `)

	return buf.String()
}
