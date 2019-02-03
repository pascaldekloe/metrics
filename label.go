package metrics

import (
	"math"
	"strings"
	"sync"
	"sync/atomic"
)

// RealGauge1 is a RealGauge with a fixed label.
// Remember that every unique combination of key-value label pairs represents a
// new time series, which can dramatically increase the amount of data stored.
type RealGauge1 struct {
	labelKey string

	mutex        sync.Mutex
	labelValues  []string
	labelSerials []string
	floatBits    []uint64
}

// RealGauge2 is a RealGauge with 2 fixed labelKeys.
// Remember that every unique combination of key-value label pairs represents a
// new time series, which can dramatically increase the amount of data stored.
type RealGauge2 struct {
	labelKeys [2]string

	mutex        sync.Mutex
	labelValues  []*[2]string
	labelSerials []string
	floatBits    []uint64
}

// RealGauge3 is a RealGauge with 3 fixed labelKeys.
// Remember that every unique combination of key-value label pairs represents a
// new time series, which can dramatically increase the amount of data stored.
type RealGauge3 struct {
	labelKeys [3]string

	mutex        sync.Mutex
	labelValues  []*[3]string
	labelSerials []string
	floatBits    []uint64
}

func (g *RealGauge1) getOrAdd(label string) *uint64 {
	g.mutex.Lock()

	for i, combo := range g.labelValues {
		if combo == label {
			g.mutex.Unlock()

			return &g.floatBits[i]
		}
	}

	g.labelValues = append(g.labelValues, label)
	g.labelSerials = append(g.labelSerials, formatLabel1(g.labelKey, label))
	i := len(g.floatBits)
	g.floatBits = append(g.floatBits, 0)

	g.mutex.Unlock()

	return &g.floatBits[i]
}

func (g *RealGauge2) getOrAdd(label1, label2 string) *uint64 {
	g.mutex.Lock()

	for i, combo := range g.labelValues {
		if combo[0] == label1 && combo[1] == label2 {
			g.mutex.Unlock()

			return &g.floatBits[i]
		}
	}

	combo := [2]string{label1, label2}
	g.labelValues = append(g.labelValues, &combo)
	g.labelSerials = append(g.labelSerials, formatLabel2(&g.labelKeys, &combo))
	i := len(g.floatBits)
	g.floatBits = append(g.floatBits, 0)

	g.mutex.Unlock()

	return &g.floatBits[i]
}

func (g *RealGauge3) getOrAdd(label1, label2, label3 string) *uint64 {
	g.mutex.Lock()

	for i, combo := range g.labelValues {
		if combo[0] == label1 && combo[1] == label2 && combo[2] == label3 {
			g.mutex.Unlock()

			return &g.floatBits[i]
		}
	}

	combo := [3]string{label1, label2, label3}
	g.labelValues = append(g.labelValues, &combo)
	g.labelSerials = append(g.labelSerials, formatLabel3(&g.labelKeys, &combo))
	i := len(g.floatBits)
	g.floatBits = append(g.floatBits, 0)

	g.mutex.Unlock()

	return &g.floatBits[i]
}

// Add is like RealGauge.Add, with the addition of a label value.
func (g *RealGauge1) Add(summand float64, label string) {
	add(g.getOrAdd(label), summand)
}

// Add is like RealGauge.Add, with the addition of 2 label values.
func (g *RealGauge2) Add(summand float64, label1, label2 string) {
	add(g.getOrAdd(label1, label2), summand)
}

// Add is like RealGauge.Add, with the addition of 3 label values.
func (g *RealGauge3) Add(summand float64, label1, label2, label3 string) {
	add(g.getOrAdd(label1, label2, label3), summand)
}

// Set is like RealGauge.Set, with the addition of a label value.
func (g *RealGauge1) Set(update float64, label string) {
	atomic.StoreUint64(g.getOrAdd(label), math.Float64bits(update))
}

// Set is like RealGauge.Set, with the addition of 2 label values.
func (g *RealGauge2) Set(update float64, label1, label2 string) {
	atomic.StoreUint64(g.getOrAdd(label1, label2), math.Float64bits(update))
}

// Set is like RealGauge.Set, with the addition of 3 label values.
func (g *RealGauge3) Set(update float64, label1, label2, label3 string) {
	atomic.StoreUint64(g.getOrAdd(label1, label2, label3), math.Float64bits(update))
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
		entry = &RealGauge1{labelKey: key}
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
		entry = &RealGauge2{
			labelKeys: [...]string{key1, key2},
		}
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
		entry = &RealGauge3{
			labelKeys: [...]string{key1, key2, key3},
		}
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

// LabelValue escapes the special characters.
var labelValue = strings.NewReplacer("\n", `\n`, `"`, `\"`, `\`, `\\`)

func formatLabel1(key, value string) string {
	var buf strings.Builder
	buf.Grow(6 + len(key) + len(value))

	buf.WriteByte('{')
	buf.WriteString(key)
	buf.WriteString(`="`)
	labelValue.WriteString(&buf, value)
	buf.WriteString(`"} `)

	return buf.String()
}

func formatLabel2(keys, values *[2]string) string {
	var buf strings.Builder
	buf.Grow(10 + len(keys[0]) + len(keys[1]) + len(values[0]) + len(values[1]))

	buf.WriteByte('{')
	buf.WriteString(keys[0])
	buf.WriteString(`="`)
	labelValue.WriteString(&buf, values[0])
	buf.WriteString(`",`)
	buf.WriteString(keys[1])
	buf.WriteString(`="`)
	labelValue.WriteString(&buf, values[1])
	buf.WriteString(`"} `)

	return buf.String()
}

func formatLabel3(keys, values *[3]string) string {
	var buf strings.Builder
	buf.Grow(14 + len(keys[0]) + len(keys[1]) + len(keys[2]) + len(values[0]) + len(values[1]) + len(values[2]))

	buf.WriteByte('{')
	buf.WriteString(keys[0])
	buf.WriteString(`="`)
	labelValue.WriteString(&buf, values[0])
	buf.WriteString(`",`)
	buf.WriteString(keys[1])
	buf.WriteString(`="`)
	labelValue.WriteString(&buf, values[1])
	buf.WriteString(`",`)
	buf.WriteString(keys[2])
	buf.WriteString(`="`)
	labelValue.WriteString(&buf, values[2])
	buf.WriteString(`"} `)

	return buf.String()
}
