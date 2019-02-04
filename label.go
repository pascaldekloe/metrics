package metrics

import (
	"strings"
	"sync"
)

// RealGauge1 is a RealGauge with a fixed label.
// Remember that every unique combination of key-value label pairs represents a
// new time series, which can dramatically increase the amount of data stored.
type RealGauge1 struct {
	name        string
	mutex       sync.Mutex
	labelKey    string
	labelValues []string
	gauges      []*RealGauge
}

// RealGauge2 is a RealGauge with 2 fixed labels.
// Remember that every unique combination of key-value label pairs represents a
// new time series, which can dramatically increase the amount of data stored.
type RealGauge2 struct {
	name        string
	mutex       sync.Mutex
	labelKeys   [2]string
	labelValues []*[2]string
	gauges      []*RealGauge
}

// RealGauge3 is a RealGauge with 3 fixed labels.
// Remember that every unique combination of key-value label pairs represents a
// new time series, which can dramatically increase the amount of data stored.
type RealGauge3 struct {
	name        string
	mutex       sync.Mutex
	labelKeys   [3]string
	labelValues []*[3]string
	gauges      []*RealGauge
}

func (g *RealGauge1) forLabel(label string) *RealGauge {
	g.mutex.Lock()

	for i, combo := range g.labelValues {
		if combo == label {
			g.mutex.Unlock()

			return g.gauges[i]
		}
	}

	g.labelValues = append(g.labelValues, label)
	entry := &RealGauge{name: g.name, head: formatHead1(g.name, g.labelKey, label)}
	g.gauges = append(g.gauges, entry)

	g.mutex.Unlock()

	return entry
}

func (g *RealGauge2) forLabels(label1, label2 string) *RealGauge {
	g.mutex.Lock()

	for i, combo := range g.labelValues {
		if combo[0] == label1 && combo[1] == label2 {
			g.mutex.Unlock()

			return g.gauges[i]
		}
	}

	combo := [2]string{label1, label2}
	entry := &RealGauge{name: g.name, head: formatHead2(g.name, &g.labelKeys, &combo)}
	g.labelValues = append(g.labelValues, &combo)
	g.gauges = append(g.gauges, entry)

	g.mutex.Unlock()

	return entry
}

func (g *RealGauge3) forLabels(label1, label2, label3 string) *RealGauge {
	g.mutex.Lock()

	for i, combo := range g.labelValues {
		if combo[0] == label1 && combo[1] == label2 && combo[2] == label3 {
			g.mutex.Unlock()

			return g.gauges[i]
		}
	}

	combo := [3]string{label1, label2, label3}
	entry := &RealGauge{name: g.name, head: formatHead3(g.name, &g.labelKeys, &combo)}
	g.labelValues = append(g.labelValues, &combo)
	g.gauges = append(g.gauges, entry)

	g.mutex.Unlock()

	return entry
}

// Add is like RealGauge.Add, with the addition of a label value.
func (g *RealGauge1) Add(summand float64, label string) {
	g.forLabel(label).Add(summand)
}

// Add is like RealGauge.Add, with the addition of 2 label values.
func (g *RealGauge2) Add(summand float64, label1, label2 string) {
	g.forLabels(label1, label2).Add(summand)
}

// Add is like RealGauge.Add, with the addition of 3 label values.
func (g *RealGauge3) Add(summand float64, label1, label2, label3 string) {
	g.forLabels(label1, label2, label3).Add(summand)
}

// Set is like RealGauge.Set, with the addition of a label value.
func (g *RealGauge1) Set(update float64, label string) {
	g.forLabel(label).Set(update)
}

// Set is like RealGauge.Set, with the addition of 2 label values.
func (g *RealGauge2) Set(update float64, label1, label2 string) {
	g.forLabels(label1, label2).Set(update)
}

// Set is like RealGauge.Set, with the addition of 3 label values.
func (g *RealGauge3) Set(update float64, label1, label2, label3 string) {
	g.forLabels(label1, label2, label3).Set(update)
}

type labeled struct {
	name    string
	gauge1s []*RealGauge1
	gauge2s []*RealGauge2
	gauge3s []*RealGauge3
}

var (
	labeledMutex   sync.Mutex
	labeledIndices = make(map[string]uint32)
	labeleds       []*labeled
)

// MustPlaceRealGauge1 registers a new RealGauge1 if the label key has not
// been used before on name. The function panics when name does not match
// regular expression [a-zA-Z_:][a-zA-Z0-9_:]* or when the label key does
// not match regular expression [a-zA-Z_][a-zA-Z0-9_]*.
func MustPlaceRealGauge1(name, key string) *RealGauge1 {
	mustValidName(name)
	mustValidKey(key)

	labeledMutex.Lock()

	var l *labeled
	if index, ok := labeledIndices[name]; ok {
		l = labeleds[index]
		if len(l.gauge1s) == 0 && len(l.gauge2s) == 0 && len(l.gauge3s) == 0 {
			panic("metrics: name in use as another type")
		}
	} else {
		l = &labeled{name: name}
		labeledIndices[name] = uint32(len(labeleds))
		labeleds = append(labeleds, l)
	}

	var entry *RealGauge1
	for _, l1 := range l.gauge1s {
		if l1.labelKey == key {
			entry = l1
			break
		}
	}
	if entry == nil {
		entry = &RealGauge1{name: name, labelKey: key}
		l.gauge1s = append(l.gauge1s, entry)
	}

	labeledMutex.Unlock()

	return entry
}

// MustPlaceRealGauge2 registers a new RealGauge2 if the label keys have
// not been used before on name. The function panics when name does not match
// regular expression [a-zA-Z_:][a-zA-Z0-9_:]* or when a label key does not match
// regular expression [a-zA-Z_][a-zA-Z0-9_]* or when the label keys do not appear
// in sorted order.
func MustPlaceRealGauge2(name, key1, key2 string) *RealGauge2 {
	mustValidName(name)
	mustValidKey(key1)
	mustValidKey(key2)
	if key1 > key2 {
		panic("metrics: label key arguments aren't sorted")
	}

	labeledMutex.Lock()

	var l *labeled
	if index, ok := labeledIndices[name]; ok {
		l = labeleds[index]
		if len(l.gauge1s) == 0 && len(l.gauge2s) == 0 && len(l.gauge3s) == 0 {
			panic("metrics: name in use as another type")
		}
	} else {
		l = &labeled{name: name}
		labeledIndices[name] = uint32(len(labeleds))
		labeleds = append(labeleds, l)
	}

	var entry *RealGauge2
	for _, l2 := range l.gauge2s {
		if l2.labelKeys[0] == key1 && l2.labelKeys[1] == key2 {
			entry = l2
			break
		}
	}
	if entry == nil {
		entry = &RealGauge2{name: name, labelKeys: [...]string{key1, key2}}
		l.gauge2s = append(l.gauge2s, entry)
	}

	labeledMutex.Unlock()

	return entry
}

// MustPlaceRealGauge3 registers a new RealGauge3 if the label keys have
// not been used before on name. The function panics when name does not match
// regular expression [a-zA-Z_:][a-zA-Z0-9_:]* or when a label key does not match
// regular expression [a-zA-Z_][a-zA-Z0-9_]* or when the label keys do not appear
// in sorted order.
func MustPlaceRealGauge3(name, key1, key2, key3 string) *RealGauge3 {
	mustValidName(name)
	mustValidKey(key1)
	mustValidKey(key2)
	mustValidKey(key3)
	if key1 > key2 || key2 > key3 {
		panic("metrics: label key arguments aren't sorted")
	}

	labeledMutex.Lock()

	var l *labeled
	if index, ok := labeledIndices[name]; ok {
		l = labeleds[index]
		if len(l.gauge1s) == 0 && len(l.gauge2s) == 0 && len(l.gauge3s) == 0 {
			panic("metrics: name in use as another type")
		}
	} else {
		l = &labeled{name: name}
		labeledIndices[name] = uint32(len(labeleds))
		labeleds = append(labeleds, l)
	}

	var entry *RealGauge3
	for _, l3 := range l.gauge3s {
		if l3.labelKeys[0] == key1 && l3.labelKeys[1] == key2 && l3.labelKeys[2] == key3 {
			entry = l3
			break
		}
	}
	if entry == nil {
		entry = &RealGauge3{name: name, labelKeys: [...]string{key1, key2, key3}}
		l.gauge3s = append(l.gauge3s, entry)
	}

	labeledMutex.Unlock()

	return entry
}

func mustValidKey(s string) {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '_' {
			continue
		}
		if i == 0 || c < '0' || c > '9' {
			panic("metrics: key doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_]*")
		}
	}
}

var valueEscapes = strings.NewReplacer("\n", `\n`, `"`, `\"`, `\`, `\\`)

func formatHead1(name, key, value string) string {
	var buf strings.Builder
	buf.Grow(6 + len(name) + len(key) + len(value))

	buf.WriteString(name)
	buf.WriteByte('{')
	buf.WriteString(key)
	buf.WriteString(`="`)
	valueEscapes.WriteString(&buf, value)
	buf.WriteString(`"} `)

	return buf.String()
}

func formatHead2(name string, keys, values *[2]string) string {
	var buf strings.Builder
	buf.Grow(10 + len(name) + len(keys[0]) + len(keys[1]) + len(values[0]) + len(values[1]))

	buf.WriteString(name)
	buf.WriteByte('{')
	buf.WriteString(keys[0])
	buf.WriteString(`="`)
	valueEscapes.WriteString(&buf, values[0])
	buf.WriteString(`",`)
	buf.WriteString(keys[1])
	buf.WriteString(`="`)
	valueEscapes.WriteString(&buf, values[1])
	buf.WriteString(`"} `)

	return buf.String()
}

func formatHead3(name string, keys, values *[3]string) string {
	var buf strings.Builder
	buf.Grow(14 + len(name) + len(keys[0]) + len(keys[1]) + len(keys[2]) + len(values[0]) + len(values[1]) + len(values[2]))

	buf.WriteString(name)
	buf.WriteByte('{')
	buf.WriteString(keys[0])
	buf.WriteString(`="`)
	valueEscapes.WriteString(&buf, values[0])
	buf.WriteString(`",`)
	buf.WriteString(keys[1])
	buf.WriteString(`="`)
	valueEscapes.WriteString(&buf, values[1])
	buf.WriteString(`",`)
	buf.WriteString(keys[2])
	buf.WriteString(`="`)
	valueEscapes.WriteString(&buf, values[2])
	buf.WriteString(`"} `)

	return buf.String()
}
