package metrics

import (
	"strings"
	"sync"
)

// FNV-1a
const (
	hashOffset = 14695981039346656037
	hashPrime  = 1099511628211
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

// Map1LabelCounter is a Counter composition with a fixed label.
// Multiple goroutines may invoke methods on a Map1LabelCounter simultaneously.
type Map1LabelCounter struct {
	map1Label
	counters []*Counter
}

// Map2LabelCounter is a Counter composition with 2 fixed labels.
// Multiple goroutines may invoke methods on a Map2LabelCounter simultaneously.
type Map2LabelCounter struct {
	map2Label
	counters []*Counter
}

// Map3LabelCounter is a Counter composition with 3 fixed labels.
// Multiple goroutines may invoke methods on a Map3LabelCounter simultaneously.
type Map3LabelCounter struct {
	map3Label
	counters []*Counter
}

// Map1LabelGauge is a Gauge composition with a fixed label.
// Multiple goroutines may invoke methods on a Map1LabelGauge simultaneously.
type Map1LabelGauge struct {
	map1Label
	gauges []*Gauge
}

// Map2LabelGauge is a Gauge composition with 2 fixed labels.
// Multiple goroutines may invoke methods on a Map2LabelGauge simultaneously.
type Map2LabelGauge struct {
	map2Label
	gauges []*Gauge
}

// Map3LabelGauge is a Gauge composition with 3 fixed labels.
// Multiple goroutines may invoke methods on a Map3LabelGauge simultaneously.
type Map3LabelGauge struct {
	map3Label
	gauges []*Gauge
}

// Map1LabelHistogram is a Histogram composition with a fixed label.
// Multiple goroutines may invoke methods on a Map1LabelHistogram simultaneously.
type Map1LabelHistogram struct {
	map1Label
	buckets    []float64
	histograms []*Histogram
}

// Map2LabelHistogram is a Histogram composition with 2 fixed labels.
// Multiple goroutines may invoke methods on a Map2LabelHistogram simultaneously.
type Map2LabelHistogram struct {
	map2Label
	buckets    []float64
	histograms []*Histogram
}

// Map3LabelHistogram is a Histogram composition with 3 fixed labels.
// Multiple goroutines may invoke methods on a Map3LabelHistogram simultaneously.
type Map3LabelHistogram struct {
	map3Label
	histograms []*Histogram
}

// Map1LabelSample is a Sample composition with a fixed label.
// Multiple goroutines may invoke methods on a Map1LabelSample simultaneously.
type Map1LabelSample struct {
	map1Label
	samples []*Sample
}

// Map2LabelSample is a Sample composition with 2 fixed labels.
// Multiple goroutines may invoke methods on a Map2LabelSample simultaneously.
type Map2LabelSample struct {
	map2Label
	samples []*Sample
}

// Map3LabelSample is a Sample composition with 3 fixed labels.
// Multiple goroutines may invoke methods on a Map3LabelSample simultaneously.
type Map3LabelSample struct {
	map3Label
	samples []*Sample
}

// With returns a dedicated Counter for a label. The value
// maps to the name as defined at Must2LabelCounter. With
// registers a new Counter if the label hasn't been used before.
// Remember that each label represents a new time series,
// which can dramatically increase the amount of data stored.
func (l1 *Map1LabelCounter) With(value string) *Counter {
	hash := uint64(hashOffset)
	for i := 0; i < len(value); i++ {
		hash ^= uint64(value[i])
		hash *= hashPrime
	}

	l1.mutex.Lock()

	for i, h := range l1.labelHashes {
		if h == hash {
			hit := l1.counters[i]

			l1.mutex.Unlock()
			return hit
		}
	}

	l1.labelHashes = append(l1.labelHashes, hash)
	c := &Counter{prefix: format1Prefix(l1.name, l1.labelName, value)}
	l1.counters = append(l1.counters, c)

	l1.mutex.Unlock()
	return c
}

// With returns a dedicated Counter for a label combination. The values
// map to the names (in order) as defined at Must2LabelCounter. With
// registers a new Counter if the combination hasn't been used before.
// Remember that each label combination represents a new time series,
// which can dramatically increase the amount of data stored.
func (l2 *Map2LabelCounter) With(value1, value2 string) *Counter {
	hash := uint64(hashOffset)
	hash ^= uint64(len(value1))
	hash *= hashPrime
	for i := 0; i < len(value1); i++ {
		hash ^= uint64(value1[i])
		hash *= hashPrime
	}
	for i := 0; i < len(value2); i++ {
		hash ^= uint64(value2[i])
		hash *= hashPrime
	}

	l2.mutex.Lock()

	for i, h := range l2.labelHashes {
		if h == hash {
			hit := l2.counters[i]

			l2.mutex.Unlock()
			return hit
		}
	}

	l2.labelHashes = append(l2.labelHashes, hash)
	c := &Counter{prefix: format2Prefix(l2.name, &l2.labelNames, value1, value2)}
	l2.counters = append(l2.counters, c)

	l2.mutex.Unlock()
	return c
}

// With returns a dedicated Counter for a label combination. The values
// map to the names (in order) as defined at Must3LabelCounter. With
// registers a new Counter if the combination hasn't been used before.
// Remember that each label combination represents a new time series,
// which can dramatically increase the amount of data stored.
func (l3 *Map3LabelCounter) With(value1, value2, value3 string) *Counter {
	hash := uint64(hashOffset)
	hash ^= uint64(len(value1))
	hash *= hashPrime
	for i := 0; i < len(value1); i++ {
		hash ^= uint64(value1[i])
		hash *= hashPrime
	}
	hash ^= uint64(len(value2))
	hash *= hashPrime
	for i := 0; i < len(value2); i++ {
		hash ^= uint64(value2[i])
		hash *= hashPrime
	}
	for i := 0; i < len(value3); i++ {
		hash ^= uint64(value3[i])
		hash *= hashPrime
	}

	l3.mutex.Lock()

	for i, h := range l3.labelHashes {
		if h == hash {
			hit := l3.counters[i]

			l3.mutex.Unlock()
			return hit
		}
	}

	l3.labelHashes = append(l3.labelHashes, hash)
	c := &Counter{prefix: format3Prefix(l3.name, &l3.labelNames, value1, value2, value3)}
	l3.counters = append(l3.counters, c)

	l3.mutex.Unlock()
	return c
}

// With returns a dedicated Gauge for a label. The value
// maps to the name as defined at Must2LabelGauge. With
// registers a new Gauge if the label hasn't been used before.
// Remember that each label represents a new time series,
// which can dramatically increase the amount of data stored.
func (l1 *Map1LabelGauge) With(value string) *Gauge {
	hash := uint64(hashOffset)
	for i := 0; i < len(value); i++ {
		hash ^= uint64(value[i])
		hash *= hashPrime
	}

	l1.mutex.Lock()

	for i, h := range l1.labelHashes {
		if h == hash {
			hit := l1.gauges[i]

			l1.mutex.Unlock()
			return hit
		}
	}

	l1.labelHashes = append(l1.labelHashes, hash)
	g := &Gauge{prefix: format1Prefix(l1.name, l1.labelName, value)}
	l1.gauges = append(l1.gauges, g)

	l1.mutex.Unlock()
	return g
}

// With returns a dedicated Gauge for a label combination. The values
// map to the names (in order) as defined at Must2LabelGauge. With
// registers a new Gauge if the combination hasn't been used before.
// Remember that each label combination represents a new time series,
// which can dramatically increase the amount of data stored.
func (l2 *Map2LabelGauge) With(value1, value2 string) *Gauge {
	hash := uint64(hashOffset)
	hash ^= uint64(len(value1))
	hash *= hashPrime
	for i := 0; i < len(value1); i++ {
		hash ^= uint64(value1[i])
		hash *= hashPrime
	}
	for i := 0; i < len(value2); i++ {
		hash ^= uint64(value2[i])
		hash *= hashPrime
	}

	l2.mutex.Lock()

	for i, h := range l2.labelHashes {
		if h == hash {
			hit := l2.gauges[i]

			l2.mutex.Unlock()
			return hit
		}
	}

	l2.labelHashes = append(l2.labelHashes, hash)
	g := &Gauge{prefix: format2Prefix(l2.name, &l2.labelNames, value1, value2)}
	l2.gauges = append(l2.gauges, g)

	l2.mutex.Unlock()
	return g
}

// With returns a dedicated Gauge for a label combination. The values
// map to the names (in order) as defined at Must3LabelGauge. With
// registers a new Gauge if the combination hasn't been used before.
// Remember that each label combination represents a new time series,
// which can dramatically increase the amount of data stored.
func (l3 *Map3LabelGauge) With(value1, value2, value3 string) *Gauge {
	hash := uint64(hashOffset)
	hash ^= uint64(len(value1))
	hash *= hashPrime
	for i := 0; i < len(value1); i++ {
		hash ^= uint64(value1[i])
		hash *= hashPrime
	}
	hash ^= uint64(len(value2))
	hash *= hashPrime
	for i := 0; i < len(value2); i++ {
		hash ^= uint64(value2[i])
		hash *= hashPrime
	}
	for i := 0; i < len(value3); i++ {
		hash ^= uint64(value3[i])
		hash *= hashPrime
	}

	l3.mutex.Lock()

	for i, h := range l3.labelHashes {
		if h == hash {
			hit := l3.gauges[i]

			l3.mutex.Unlock()
			return hit
		}
	}

	l3.labelHashes = append(l3.labelHashes, hash)
	g := &Gauge{prefix: format3Prefix(l3.name, &l3.labelNames, value1, value2, value3)}
	l3.gauges = append(l3.gauges, g)

	l3.mutex.Unlock()
	return g
}

// With returns a dedicated Histogram for a label. The value
// maps to the name as defined at Must2LabelHistogram. With
// registers a new Histogram if the label hasn't been used before.
// Remember that each label represents a new time series,
// which can dramatically increase the amount of data stored.
func (l1 *Map1LabelHistogram) With(value string, buckets ...float64) *Histogram {
	hash := uint64(hashOffset)
	for i := 0; i < len(value); i++ {
		hash ^= uint64(value[i])
		hash *= hashPrime
	}

	l1.mutex.Lock()

	for i, h := range l1.labelHashes {
		if h == hash {
			hit := l1.histograms[i]

			l1.mutex.Unlock()
			return hit
		}
	}

	l1.labelHashes = append(l1.labelHashes, hash)
	h := newHistogram(l1.name, buckets)
	l1.histograms = append(l1.histograms, h)

	l1.mutex.Unlock()
	return h
}

// With returns a dedicated Histogram for a label combination. The values
// map to the names (in order) as defined at Must2LabelHistogram. With
// registers a new Histogram if the combination hasn't been used before.
// Remember that each label combination represents a new time series,
// which can dramatically increase the amount of data stored.
func (l2 *Map2LabelHistogram) With(value1, value2 string, buckets ...float64) *Histogram {
	hash := uint64(hashOffset)
	hash ^= uint64(len(value1))
	hash *= hashPrime
	for i := 0; i < len(value1); i++ {
		hash ^= uint64(value1[i])
		hash *= hashPrime
	}
	for i := 0; i < len(value2); i++ {
		hash ^= uint64(value2[i])
		hash *= hashPrime
	}

	l2.mutex.Lock()

	for i, h := range l2.labelHashes {
		if h == hash {
			hit := l2.histograms[i]

			l2.mutex.Unlock()
			return hit
		}
	}

	l2.labelHashes = append(l2.labelHashes, hash)
	h := newHistogram(l2.name, buckets)
	l2.histograms = append(l2.histograms, h)

	l2.mutex.Unlock()
	return h
}

// With returns a dedicated Histogram for a label combination. The values
// map to the names (in order) as defined at Must3LabelHistogram. With
// registers a new Histogram if the combination hasn't been used before.
// Remember that each label combination represents a new time series,
// which can dramatically increase the amount of data stored.
func (l3 *Map3LabelHistogram) With(value1, value2, value3 string, buckets ...float64) *Histogram {
	hash := uint64(hashOffset)
	hash ^= uint64(len(value1))
	hash *= hashPrime
	for i := 0; i < len(value1); i++ {
		hash ^= uint64(value1[i])
		hash *= hashPrime
	}
	hash ^= uint64(len(value2))
	hash *= hashPrime
	for i := 0; i < len(value2); i++ {
		hash ^= uint64(value2[i])
		hash *= hashPrime
	}
	for i := 0; i < len(value3); i++ {
		hash ^= uint64(value3[i])
		hash *= hashPrime
	}

	l3.mutex.Lock()

	for i, h := range l3.labelHashes {
		if h == hash {
			hit := l3.histograms[i]

			l3.mutex.Unlock()
			return hit
		}
	}

	l3.labelHashes = append(l3.labelHashes, hash)
	h := newHistogram(l3.name, buckets)
	l3.histograms = append(l3.histograms, h)

	l3.mutex.Unlock()
	return h
}

// With returns a dedicated Sample for a label. The value
// maps to the name as defined at Must2LabelSample. With
// registers a new Sample if the label hasn't been used before.
// Remember that each label represents a new time series,
// which can dramatically increase the amount of data stored.
func (l1 *Map1LabelSample) With(value string) *Sample {
	hash := uint64(hashOffset)
	for i := 0; i < len(value); i++ {
		hash ^= uint64(value[i])
		hash *= hashPrime
	}

	l1.mutex.Lock()

	for i, h := range l1.labelHashes {
		if h == hash {
			hit := l1.samples[i]

			l1.mutex.Unlock()
			return hit
		}
	}

	l1.labelHashes = append(l1.labelHashes, hash)
	s := &Sample{prefix: format1Prefix(l1.name, l1.labelName, value)}
	l1.samples = append(l1.samples, s)

	l1.mutex.Unlock()
	return s
}

// With returns a dedicated Sample for a label combination. The values
// map to the names (in order) as defined at Must2LabelSample. With
// registers a new Sample if the combination hasn't been used before.
// Remember that each label combination represents a new time series,
// which can dramatically increase the amount of data stored.
func (l2 *Map2LabelSample) With(value1, value2 string) *Sample {
	hash := uint64(hashOffset)
	hash ^= uint64(len(value1))
	hash *= hashPrime
	for i := 0; i < len(value1); i++ {
		hash ^= uint64(value1[i])
		hash *= hashPrime
	}
	for i := 0; i < len(value2); i++ {
		hash ^= uint64(value2[i])
		hash *= hashPrime
	}

	l2.mutex.Lock()

	for i, h := range l2.labelHashes {
		if h == hash {
			hit := l2.samples[i]

			l2.mutex.Unlock()
			return hit
		}
	}

	l2.labelHashes = append(l2.labelHashes, hash)
	s := &Sample{prefix: format2Prefix(l2.name, &l2.labelNames, value1, value2)}
	l2.samples = append(l2.samples, s)

	l2.mutex.Unlock()
	return s
}

// With returns a dedicated Sample for a label combination. The values
// map to the names (in order) as defined at Must3LabelSample. With
// registers a new Sample if the combination hasn't been used before.
// Remember that each label combination represents a new time series,
// which can dramatically increase the amount of data stored.
func (l3 *Map3LabelSample) With(value1, value2, value3 string) *Sample {
	hash := uint64(hashOffset)
	hash ^= uint64(len(value1))
	hash *= hashPrime
	for i := 0; i < len(value1); i++ {
		hash ^= uint64(value1[i])
		hash *= hashPrime
	}
	hash ^= uint64(len(value2))
	hash *= hashPrime
	for i := 0; i < len(value2); i++ {
		hash ^= uint64(value2[i])
		hash *= hashPrime
	}
	for i := 0; i < len(value3); i++ {
		hash ^= uint64(value3[i])
		hash *= hashPrime
	}

	l3.mutex.Lock()

	for i, h := range l3.labelHashes {
		if h == hash {
			hit := l3.samples[i]

			l3.mutex.Unlock()
			return hit
		}
	}

	l3.labelHashes = append(l3.labelHashes, hash)
	s := &Sample{prefix: format3Prefix(l3.name, &l3.labelNames, value1, value2, value3)}
	l3.samples = append(l3.samples, s)

	l3.mutex.Unlock()
	return s
}

// Must1LabelCounter returns a composition with one fixed label.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]* or
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]*
func Must1LabelCounter(name, labelName string) *Map1LabelCounter {
	mustValidName(name)
	mustValidLabelName(labelName)

	mutex.Lock()

	var l1 *Map1LabelCounter
	if index, ok := indices[name]; !ok {
		l1 = &Map1LabelCounter{map1Label: map1Label{
			name: name, labelName: labelName}}

		indices[name] = uint32(len(metrics))
		metrics = append(metrics, &metric{
			typeComment: typePrefix + name + counterTypeLineEnd,
			counterL1s:  []*Map1LabelCounter{l1},
		})
	} else {
		m := metrics[index]
		if m.typeID() != counterType {
			panic("metrics: name in use as another type")
		}

		for _, o := range m.counterL1s {
			if o.labelName == labelName {
				l1 = o
				break
			}
		}
		if l1 == nil {
			l1 = &Map1LabelCounter{map1Label: map1Label{
				name: name, labelName: labelName}}
			m.counterL1s = append(m.counterL1s, l1)
		}
	}

	mutex.Unlock()
	return l1
}

// Must2LabelCounter returns a composition with two fixed labels.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) the label names do not appear in ascending order.
func Must2LabelCounter(name, label1Name, label2Name string) *Map2LabelCounter {
	mustValidName(name)
	mustValidLabelName(label1Name)
	mustValidLabelName(label2Name)
	if label1Name > label2Name {
		panic("metrics: label name arguments aren't sorted")
	}

	mutex.Lock()

	var l2 *Map2LabelCounter
	if index, ok := indices[name]; !ok {
		l2 = &Map2LabelCounter{map2Label: map2Label{
			name: name, labelNames: [...]string{label1Name, label2Name}}}

		indices[name] = uint32(len(metrics))
		metrics = append(metrics, &metric{
			typeComment: typePrefix + name + counterTypeLineEnd,
			counterL2s:  []*Map2LabelCounter{l2},
		})
	} else {
		m := metrics[index]
		if m.typeID() != counterType {
			panic("metrics: name in use as another type")
		}

		for _, o := range m.counterL2s {
			if o.labelNames[0] == label1Name && o.labelNames[1] == label2Name {
				l2 = o
				break
			}
		}
		if l2 == nil {
			l2 = &Map2LabelCounter{map2Label: map2Label{
				name: name, labelNames: [...]string{label1Name, label2Name}}}
			m.counterL2s = append(m.counterL2s, l2)
		}
	}

	mutex.Unlock()
	return l2
}

// Must3LabelCounter returns a composition with three fixed labels.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) the label names do not appear in ascending order.
func Must3LabelCounter(name, label1Name, label2Name, label3Name string) *Map3LabelCounter {
	mustValidName(name)
	mustValidLabelName(label1Name)
	mustValidLabelName(label2Name)
	mustValidLabelName(label3Name)
	if label1Name > label2Name || label2Name > label3Name {
		panic("metrics: label name arguments aren't sorted")
	}

	mutex.Lock()

	var l3 *Map3LabelCounter
	if index, ok := indices[name]; !ok {
		l3 = &Map3LabelCounter{map3Label: map3Label{
			name: name, labelNames: [...]string{label1Name, label2Name, label3Name}}}

		indices[name] = uint32(len(metrics))
		metrics = append(metrics, &metric{
			typeComment: typePrefix + name + counterTypeLineEnd,
			counterL3s:  []*Map3LabelCounter{l3},
		})
	} else {
		m := metrics[index]
		if m.typeID() != counterType {
			panic("metrics: name in use as another type")
		}

		for _, o := range m.counterL3s {
			if o.labelNames[0] == label1Name && o.labelNames[1] == label2Name && o.labelNames[2] == label3Name {
				l3 = o
				break
			}
		}
		if l3 == nil {
			l3 = &Map3LabelCounter{map3Label: map3Label{
				name: name, labelNames: [...]string{label1Name, label2Name, label3Name}}}
			m.counterL3s = append(m.counterL3s, l3)
		}
	}

	mutex.Unlock()
	return l3
}

// Must1LabelGauge returns a composition with one fixed label.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]* or
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]*
func Must1LabelGauge(name, labelName string) *Map1LabelGauge {
	mustValidName(name)
	mustValidLabelName(labelName)

	mutex.Lock()

	var l1 *Map1LabelGauge
	if index, ok := indices[name]; !ok {
		l1 = &Map1LabelGauge{map1Label: map1Label{
			name: name, labelName: labelName}}

		indices[name] = uint32(len(metrics))
		metrics = append(metrics, &metric{
			typeComment: typePrefix + name + gaugeTypeLineEnd,
			gaugeL1s:    []*Map1LabelGauge{l1},
		})
	} else {
		m := metrics[index]
		if m.typeID() != gaugeType {
			panic("metrics: name in use as another type")
		}

		for _, o := range m.gaugeL1s {
			if o.labelName == labelName {
				l1 = o
				break
			}
		}
		if l1 == nil {
			l1 = &Map1LabelGauge{map1Label: map1Label{
				name: name, labelName: labelName}}
			m.gaugeL1s = append(m.gaugeL1s, l1)
		}
	}

	mutex.Unlock()
	return l1
}

// Must2LabelGauge returns a composition with two fixed labels.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) the label names do not appear in ascending order.
func Must2LabelGauge(name, label1Name, label2Name string) *Map2LabelGauge {
	mustValidName(name)
	mustValidLabelName(label1Name)
	mustValidLabelName(label2Name)
	if label1Name > label2Name {
		panic("metrics: label name arguments aren't sorted")
	}

	mutex.Lock()

	var l2 *Map2LabelGauge
	if index, ok := indices[name]; !ok {
		l2 = &Map2LabelGauge{map2Label: map2Label{
			name: name, labelNames: [...]string{label1Name, label2Name}}}

		indices[name] = uint32(len(metrics))
		metrics = append(metrics, &metric{
			typeComment: typePrefix + name + gaugeTypeLineEnd,
			gaugeL2s:    []*Map2LabelGauge{l2},
		})
	} else {
		m := metrics[index]
		if m.typeID() != gaugeType {
			panic("metrics: name in use as another type")
		}

		for _, o := range m.gaugeL2s {
			if o.labelNames[0] == label1Name && o.labelNames[1] == label2Name {
				l2 = o
				break
			}
		}
		if l2 == nil {
			l2 = &Map2LabelGauge{map2Label: map2Label{
				name: name, labelNames: [...]string{label1Name, label2Name}}}
			m.gaugeL2s = append(m.gaugeL2s, l2)
		}
	}

	mutex.Unlock()
	return l2
}

// Must3LabelGauge returns a composition with three fixed labels.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) the label names do not appear in ascending order.
func Must3LabelGauge(name, label1Name, label2Name, label3Name string) *Map3LabelGauge {
	mustValidName(name)
	mustValidLabelName(label1Name)
	mustValidLabelName(label2Name)
	mustValidLabelName(label3Name)
	if label1Name > label2Name || label2Name > label3Name {
		panic("metrics: label name arguments aren't sorted")
	}

	mutex.Lock()

	var l3 *Map3LabelGauge
	if index, ok := indices[name]; !ok {
		l3 = &Map3LabelGauge{map3Label: map3Label{
			name: name, labelNames: [...]string{label1Name, label2Name, label3Name}}}

		indices[name] = uint32(len(metrics))
		metrics = append(metrics, &metric{
			typeComment: typePrefix + name + gaugeTypeLineEnd,
			gaugeL3s:    []*Map3LabelGauge{l3},
		})
	} else {
		m := metrics[index]
		if m.typeID() != gaugeType {
			panic("metrics: name in use as another type")
		}

		for _, o := range m.gaugeL3s {
			if o.labelNames[0] == label1Name && o.labelNames[1] == label2Name && o.labelNames[2] == label3Name {
				l3 = o
				break
			}
		}
		if l3 == nil {
			l3 = &Map3LabelGauge{map3Label: map3Label{
				name: name, labelNames: [...]string{label1Name, label2Name, label3Name}}}
			m.gaugeL3s = append(m.gaugeL3s, l3)
		}
	}

	mutex.Unlock()
	return l3
}

// Must1LabelHistogram returns a composition with one fixed label.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]* or
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]*
func Must1LabelHistogram(name, labelName string) *Map1LabelHistogram {
	mustValidName(name)
	mustValidLabelName(labelName)

	mutex.Lock()

	var l1 *Map1LabelHistogram
	if index, ok := indices[name]; !ok {
		l1 = &Map1LabelHistogram{map1Label: map1Label{
			name: name, labelName: labelName}}

		indices[name] = uint32(len(metrics))
		metrics = append(metrics, &metric{
			typeComment:  typePrefix + name + histogramTypeLineEnd,
			histogramL1s: []*Map1LabelHistogram{l1},
		})
	} else {
		m := metrics[index]
		if m.typeID() != histogramType {
			panic("metrics: name in use as another type")
		}

		for _, o := range m.histogramL1s {
			if o.labelName == labelName {
				l1 = o
				break
			}
		}
		if l1 == nil {
			l1 = &Map1LabelHistogram{map1Label: map1Label{
				name: name, labelName: labelName}}
			m.histogramL1s = append(m.histogramL1s, l1)
		}
	}

	mutex.Unlock()
	return l1
}

// Must2LabelHistogram returns a composition with two fixed labels.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) the label names do not appear in ascending order.
func Must2LabelHistogram(name, label1Name, label2Name string) *Map2LabelHistogram {
	mustValidName(name)
	mustValidLabelName(label1Name)
	mustValidLabelName(label2Name)
	if label1Name > label2Name {
		panic("metrics: label name arguments aren't sorted")
	}

	mutex.Lock()

	var l2 *Map2LabelHistogram
	if index, ok := indices[name]; !ok {
		l2 = &Map2LabelHistogram{map2Label: map2Label{
			name: name, labelNames: [...]string{label1Name, label2Name}}}

		indices[name] = uint32(len(metrics))
		metrics = append(metrics, &metric{
			typeComment:  typePrefix + name + histogramTypeLineEnd,
			histogramL2s: []*Map2LabelHistogram{l2},
		})
	} else {
		m := metrics[index]
		if m.typeID() != histogramType {
			panic("metrics: name in use as another type")
		}

		for _, o := range m.histogramL2s {
			if o.labelNames[0] == label1Name && o.labelNames[1] == label2Name {
				l2 = o
				break
			}
		}
		if l2 == nil {
			l2 = &Map2LabelHistogram{map2Label: map2Label{
				name: name, labelNames: [...]string{label1Name, label2Name}}}
			m.histogramL2s = append(m.histogramL2s, l2)
		}
	}

	mutex.Unlock()
	return l2
}

// Must3LabelHistogram returns a composition with three fixed labels.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) the label names do not appear in ascending order.
func Must3LabelHistogram(name, label1Name, label2Name, label3Name string) *Map3LabelHistogram {
	mustValidName(name)
	mustValidLabelName(label1Name)
	mustValidLabelName(label2Name)
	mustValidLabelName(label3Name)
	if label1Name > label2Name || label2Name > label3Name {
		panic("metrics: label name arguments aren't sorted")
	}

	mutex.Lock()

	var l3 *Map3LabelHistogram
	if index, ok := indices[name]; !ok {
		l3 = &Map3LabelHistogram{map3Label: map3Label{
			name: name, labelNames: [...]string{label1Name, label2Name, label3Name}}}

		indices[name] = uint32(len(metrics))
		metrics = append(metrics, &metric{
			typeComment:  typePrefix + name + histogramTypeLineEnd,
			histogramL3s: []*Map3LabelHistogram{l3},
		})
	} else {
		m := metrics[index]
		if m.typeID() != histogramType {
			panic("metrics: name in use as another type")
		}

		for _, o := range m.histogramL3s {
			if o.labelNames[0] == label1Name && o.labelNames[1] == label2Name && o.labelNames[2] == label3Name {
				l3 = o
				break
			}
		}
		if l3 == nil {
			l3 = &Map3LabelHistogram{map3Label: map3Label{
				name: name, labelNames: [...]string{label1Name, label2Name, label3Name}}}
			m.histogramL3s = append(m.histogramL3s, l3)
		}
	}

	mutex.Unlock()
	return l3
}

// Must1LabelCounterSample returns a composition with one fixed label.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]* or
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]*
func Must1LabelCounterSample(name, labelName string) *Map1LabelSample {
	mustValidName(name)
	mustValidLabelName(labelName)

	mutex.Lock()

	var l1 *Map1LabelSample
	if index, ok := indices[name]; !ok {
		l1 = &Map1LabelSample{map1Label: map1Label{
			name: name, labelName: labelName}}

		indices[name] = uint32(len(metrics))
		metrics = append(metrics, &metric{
			typeComment: typePrefix + name + counterTypeLineEnd,
			sampleL1s:   []*Map1LabelSample{l1},
		})
	} else {
		m := metrics[index]
		if m.typeID() != counterType {
			panic("metrics: name in use as another type")
		}

		for _, o := range m.sampleL1s {
			if o.labelName == labelName {
				l1 = o
				break
			}
		}
		if l1 == nil {
			l1 = &Map1LabelSample{map1Label: map1Label{
				name: name, labelName: labelName}}
			m.sampleL1s = append(m.sampleL1s, l1)
		}
	}

	mutex.Unlock()
	return l1
}

// Must2LabelCounterSample returns a composition with two fixed labels.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) the label names do not appear in ascending order.
func Must2LabelCounterSample(name, label1Name, label2Name string) *Map2LabelSample {
	mustValidName(name)
	mustValidLabelName(label1Name)
	mustValidLabelName(label2Name)
	if label1Name > label2Name {
		panic("metrics: label name arguments aren't sorted")
	}

	mutex.Lock()

	var l2 *Map2LabelSample
	if index, ok := indices[name]; !ok {
		l2 = &Map2LabelSample{map2Label: map2Label{
			name: name, labelNames: [...]string{label1Name, label2Name}}}

		indices[name] = uint32(len(metrics))
		metrics = append(metrics, &metric{
			typeComment: typePrefix + name + counterTypeLineEnd,
			sampleL2s:   []*Map2LabelSample{l2},
		})
	} else {
		m := metrics[index]
		if m.typeID() != counterType {
			panic("metrics: name in use as another type")
		}

		for _, o := range m.sampleL2s {
			if o.labelNames[0] == label1Name && o.labelNames[1] == label2Name {
				l2 = o
				break
			}
		}
		if l2 == nil {
			l2 = &Map2LabelSample{map2Label: map2Label{
				name: name, labelNames: [...]string{label1Name, label2Name}}}
			m.sampleL2s = append(m.sampleL2s, l2)
		}
	}

	mutex.Unlock()
	return l2
}

// Must3LabelCounterSample returns a composition with three fixed labels.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) the label names do not appear in ascending order.
func Must3LabelCounterSample(name, label1Name, label2Name, label3Name string) *Map3LabelSample {
	mustValidName(name)
	mustValidLabelName(label1Name)
	mustValidLabelName(label2Name)
	mustValidLabelName(label3Name)
	if label1Name > label2Name || label2Name > label3Name {
		panic("metrics: label name arguments aren't sorted")
	}

	mutex.Lock()

	var l3 *Map3LabelSample
	if index, ok := indices[name]; !ok {
		l3 = &Map3LabelSample{map3Label: map3Label{
			name: name, labelNames: [...]string{label1Name, label2Name, label3Name}}}

		indices[name] = uint32(len(metrics))
		metrics = append(metrics, &metric{
			typeComment: typePrefix + name + counterTypeLineEnd,
			sampleL3s:   []*Map3LabelSample{l3},
		})
	} else {
		m := metrics[index]
		if m.typeID() != counterType {
			panic("metrics: name in use as another type")
		}

		for _, o := range m.sampleL3s {
			if o.labelNames[0] == label1Name && o.labelNames[1] == label2Name && o.labelNames[2] == label3Name {
				l3 = o
				break
			}
		}
		if l3 == nil {
			l3 = &Map3LabelSample{map3Label: map3Label{
				name: name, labelNames: [...]string{label1Name, label2Name, label3Name}}}
			m.sampleL3s = append(m.sampleL3s, l3)
		}
	}

	mutex.Unlock()
	return l3
}

// Must1LabelGaugeSample returns a composition with one fixed label.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]* or
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]*
func Must1LabelGaugeSample(name, labelName string) *Map1LabelSample {
	mustValidName(name)
	mustValidLabelName(labelName)

	mutex.Lock()

	var l1 *Map1LabelSample
	if index, ok := indices[name]; !ok {
		l1 = &Map1LabelSample{map1Label: map1Label{
			name: name, labelName: labelName}}

		indices[name] = uint32(len(metrics))
		metrics = append(metrics, &metric{
			typeComment: typePrefix + name + gaugeTypeLineEnd,
			sampleL1s:   []*Map1LabelSample{l1},
		})
	} else {
		m := metrics[index]
		if m.typeID() != gaugeType {
			panic("metrics: name in use as another type")
		}

		for _, o := range m.sampleL1s {
			if o.labelName == labelName {
				l1 = o
				break
			}
		}
		if l1 == nil {
			l1 = &Map1LabelSample{map1Label: map1Label{
				name: name, labelName: labelName}}
			m.sampleL1s = append(m.sampleL1s, l1)
		}
	}

	mutex.Unlock()
	return l1
}

// Must2LabelGaugeSample returns a composition with two fixed labels.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) the label names do not appear in ascending order.
func Must2LabelGaugeSample(name, label1Name, label2Name string) *Map2LabelSample {
	mustValidName(name)
	mustValidLabelName(label1Name)
	mustValidLabelName(label2Name)
	if label1Name > label2Name {
		panic("metrics: label name arguments aren't sorted")
	}

	mutex.Lock()

	var l2 *Map2LabelSample
	if index, ok := indices[name]; !ok {
		l2 = &Map2LabelSample{map2Label: map2Label{
			name: name, labelNames: [...]string{label1Name, label2Name}}}

		indices[name] = uint32(len(metrics))
		metrics = append(metrics, &metric{
			typeComment: typePrefix + name + gaugeTypeLineEnd,
			sampleL2s:   []*Map2LabelSample{l2},
		})
	} else {
		m := metrics[index]
		if m.typeID() != gaugeType {
			panic("metrics: name in use as another type")
		}

		for _, o := range m.sampleL2s {
			if o.labelNames[0] == label1Name && o.labelNames[1] == label2Name {
				l2 = o
				break
			}
		}
		if l2 == nil {
			l2 = &Map2LabelSample{map2Label: map2Label{
				name: name, labelNames: [...]string{label1Name, label2Name}}}
			m.sampleL2s = append(m.sampleL2s, l2)
		}
	}

	mutex.Unlock()
	return l2
}

// Must3LabelGaugeSample returns a composition with three fixed labels.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) the label names do not appear in ascending order.
func Must3LabelGaugeSample(name, label1Name, label2Name, label3Name string) *Map3LabelSample {
	mustValidName(name)
	mustValidLabelName(label1Name)
	mustValidLabelName(label2Name)
	mustValidLabelName(label3Name)
	if label1Name > label2Name || label2Name > label3Name {
		panic("metrics: label name arguments aren't sorted")
	}

	mutex.Lock()

	var l3 *Map3LabelSample
	if index, ok := indices[name]; !ok {
		l3 = &Map3LabelSample{map3Label: map3Label{
			name: name, labelNames: [...]string{label1Name, label2Name, label3Name}}}

		indices[name] = uint32(len(metrics))
		metrics = append(metrics, &metric{
			typeComment: typePrefix + name + gaugeTypeLineEnd,
			sampleL3s:   []*Map3LabelSample{l3},
		})
	} else {
		m := metrics[index]
		if m.typeID() != gaugeType {
			panic("metrics: name in use as another type")
		}

		for _, o := range m.sampleL3s {
			if o.labelNames[0] == label1Name && o.labelNames[1] == label2Name && o.labelNames[2] == label3Name {
				l3 = o
				break
			}
		}
		if l3 == nil {
			l3 = &Map3LabelSample{map3Label: map3Label{
				name: name, labelNames: [...]string{label1Name, label2Name, label3Name}}}
			m.sampleL3s = append(m.sampleL3s, l3)
		}
	}

	mutex.Unlock()
	return l3
}

func mustValidLabelName(s string) {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '_' {
			continue
		}
		if i == 0 || c < '0' || c > '9' {
			panic("metrics: label name doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_]*")
		}
	}
}

var valueEscapes = strings.NewReplacer("\n", `\n`, `"`, `\"`, `\`, `\\`)

func format1Prefix(name, labelName, labelValue string) string {
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

func format2Prefix(name string, labelNames *[2]string, labelValue1, labelValue2 string) string {
	var buf strings.Builder
	buf.Grow(10 + len(name) + len(labelNames[0]) + len(labelNames[1]) + len(labelValue1) + len(labelValue2))

	buf.WriteString(name)
	buf.WriteByte('{')
	buf.WriteString(labelNames[0])
	buf.WriteString(`="`)
	valueEscapes.WriteString(&buf, labelValue1)
	buf.WriteString(`",`)
	buf.WriteString(labelNames[1])
	buf.WriteString(`="`)
	valueEscapes.WriteString(&buf, labelValue2)
	buf.WriteString(`"} `)

	return buf.String()
}

func format3Prefix(name string, labelNames *[3]string, labelValue1, labelValue2, labelValue3 string) string {
	var buf strings.Builder
	buf.Grow(14 + len(name) + len(labelNames[0]) + len(labelNames[1]) + len(labelNames[2]) + len(labelValue1) + len(labelValue2) + len(labelValue3))

	buf.WriteString(name)
	buf.WriteByte('{')
	buf.WriteString(labelNames[0])
	buf.WriteString(`="`)
	valueEscapes.WriteString(&buf, labelValue1)
	buf.WriteString(`",`)
	buf.WriteString(labelNames[1])
	buf.WriteString(`="`)
	valueEscapes.WriteString(&buf, labelValue2)
	buf.WriteString(`",`)
	buf.WriteString(labelNames[2])
	buf.WriteString(`="`)
	valueEscapes.WriteString(&buf, labelValue3)
	buf.WriteString(`"} `)

	return buf.String()
}
