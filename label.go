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

// Map1LabelGauge is a gauge composition with a fixed label.
// Multiple goroutines may invoke methods on a Map1LabelGauge simultaneously.
type Map1LabelGauge struct {
	mutex       sync.Mutex
	name        string
	labelName   string
	labelHashes []uint64
	gauges      []*Gauge
}

// Map2LabelGauge is a gauge composition with 2 fixed labels.
// Multiple goroutines may invoke methods on a Map2LabelGauge simultaneously.
type Map2LabelGauge struct {
	mutex       sync.Mutex
	name        string
	labelNames  [2]string
	labelHashes []uint64
	gauges      []*Gauge
}

// Map3LabelGauge is a gauge composition with 3 fixed labels.
// Multiple goroutines may invoke methods on a Map3LabelGauge simultaneously.
type Map3LabelGauge struct {
	mutex       sync.Mutex
	name        string
	labelNames  [3]string
	labelHashes []uint64
	gauges      []*Gauge
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
		l1 = &Map1LabelGauge{name: name, labelName: labelName}

		indices[name] = uint32(len(metrics))
		metrics = append(metrics, &metric{
			typeComment: typePrefix + name + gaugeTypeLineEnd,
			typeID:      gaugeType,
			gaugeL1s:    []*Map1LabelGauge{l1},
		})
	} else {
		m := metrics[index]
		if m.typeID != gaugeType {
			panic("metrics: name in use as another type")
		}

		for _, o := range m.gaugeL1s {
			if o.labelName == labelName {
				l1 = o
				break
			}
		}
		if l1 == nil {
			l1 = &Map1LabelGauge{name: name, labelName: labelName}
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
		l2 = &Map2LabelGauge{name: name, labelNames: [...]string{label1Name, label2Name}}

		indices[name] = uint32(len(metrics))
		metrics = append(metrics, &metric{
			typeComment: typePrefix + name + gaugeTypeLineEnd,
			typeID:      gaugeType,
			gaugeL2s:    []*Map2LabelGauge{l2},
		})
	} else {
		m := metrics[index]
		if m.typeID != gaugeType {
			panic("metrics: name in use as another type")
		}

		for _, o := range m.gaugeL2s {
			if o.labelNames[0] == label1Name && o.labelNames[1] == label2Name {
				l2 = o
				break
			}
		}
		if l2 == nil {
			l2 = &Map2LabelGauge{name: name, labelNames: [...]string{label1Name, label2Name}}
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
		l3 = &Map3LabelGauge{
			name:       name,
			labelNames: [...]string{label1Name, label2Name, label3Name},
		}

		indices[name] = uint32(len(metrics))
		metrics = append(metrics, &metric{
			typeComment: typePrefix + name + gaugeTypeLineEnd,
			typeID:      gaugeType,
			gaugeL3s:    []*Map3LabelGauge{l3},
		})
	} else {
		m := metrics[index]
		if m.typeID != gaugeType {
			panic("metrics: name in use as another type")
		}

		for _, o := range m.gaugeL3s {
			if o.labelNames[0] == label1Name && o.labelNames[1] == label2Name && o.labelNames[2] == label3Name {
				l3 = o
				break
			}
		}
		if l3 == nil {
			l3 = &Map3LabelGauge{
				name:       name,
				labelNames: [...]string{label1Name, label2Name, label3Name},
			}
			m.gaugeL3s = append(m.gaugeL3s, l3)
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
