package metrics

import (
	"strings"
	"sync"
)

// Map1LabelGauge is a gauge composition with a fixed label.
// Remember that every unique combination of key-value label pairs represents a
// new time series, which can dramatically increase the amount of data stored.
type Map1LabelGauge struct {
	name        string
	mutex       sync.Mutex
	labelKey    string
	labelValues []string
	gauges      []*Gauge
}

// Map2LabelGauge is a gauge composition with 2 fixed labels.
// Remember that every unique combination of key-value label pairs represents a
// new time series, which can dramatically increase the amount of data stored.
type Map2LabelGauge struct {
	name        string
	mutex       sync.Mutex
	labelKeys   [2]string
	labelValues []*[2]string
	gauges      []*Gauge
}

// Map3LabelGauge is a gauge composition with 3 fixed labels.
// Remember that every unique combination of key-value label pairs represents a
// new time series, which can dramatically increase the amount of data stored.
type Map3LabelGauge struct {
	name        string
	mutex       sync.Mutex
	labelKeys   [3]string
	labelValues []*[3]string
	gauges      []*Gauge
}

// With registers a new Gauge if the label hasn't been used before.
// The value maps to the key as defined with Must1LabelGauge.
func (g *Map1LabelGauge) With(value string) *Gauge {
	g.mutex.Lock()

	for i, combo := range g.labelValues {
		if combo == value {
			g.mutex.Unlock()

			return g.gauges[i]
		}
	}

	g.labelValues = append(g.labelValues, value)
	entry := &Gauge{prefix: format1Prefix(g.name, g.labelKey, value)}
	g.gauges = append(g.gauges, entry)

	g.mutex.Unlock()

	return entry
}

// With registers a new Gauge if the labels haven't been used before.
// The values map to the keys (in order) as defined with Must2LabelGauge.
func (g *Map2LabelGauge) With(value1, value2 string) *Gauge {
	g.mutex.Lock()

	for i, combo := range g.labelValues {
		if combo[0] == value1 && combo[1] == value2 {
			g.mutex.Unlock()

			return g.gauges[i]
		}
	}

	combo := [2]string{value1, value2}
	entry := &Gauge{prefix: format2Prefix(g.name, &g.labelKeys, &combo)}
	g.labelValues = append(g.labelValues, &combo)
	g.gauges = append(g.gauges, entry)

	g.mutex.Unlock()

	return entry
}

// With registers a new Gauge if the labels haven't been used before.
// The values map to the keys (in order) as defined with Must3LabelGauge.
func (g *Map3LabelGauge) With(value1, value2, value3 string) *Gauge {
	g.mutex.Lock()

	for i, combo := range g.labelValues {
		if combo[0] == value1 && combo[1] == value2 && combo[2] == value3 {
			g.mutex.Unlock()

			return g.gauges[i]
		}
	}

	combo := [3]string{value1, value2, value3}
	entry := &Gauge{prefix: format3Prefix(g.name, &g.labelKeys, &combo)}
	g.labelValues = append(g.labelValues, &combo)
	g.gauges = append(g.gauges, entry)

	g.mutex.Unlock()

	return entry
}

// Must1LabelGauge returns a composition with one fixed label key.
// The function panics on any of the following:
//	* name in use as another metric type
//	* name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*
//	* key does not match regular expression [a-zA-Z_][a-zA-Z0-9_]*
func Must1LabelGauge(name, key string) *Map1LabelGauge {
	mustValidName(name)
	mustValidKey(key)

	mutex.Lock()

	var g *Map1LabelGauge
	if index, ok := indices[name]; !ok {
		g = &Map1LabelGauge{name: name, labelKey: key}

		indices[name] = uint32(len(metrics))
		metrics = append(metrics, &metric{
			typeComment: typePrefix + name + gaugeTypeLineEnd,
			typeID:      gaugeType,
			gaugeL1s:    []*Map1LabelGauge{g},
		})
	} else {
		m := metrics[index]
		if m.typeID != gaugeType {
			panic("metrics: name in use as another type")
		}

		for _, l1 := range m.gaugeL1s {
			if l1.labelKey == key {
				g = l1
				break
			}
		}
		if g == nil {
			g = &Map1LabelGauge{name: name, labelKey: key}
			m.gaugeL1s = append(m.gaugeL1s, g)
		}
	}

	mutex.Unlock()

	return g
}

// Must2LabelGauge returns a composition with two fixed label keys.
// The function panics on any of the following:
//	* name in use as another metric type
//	* name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*
//	* a key does not match regular expression [a-zA-Z_][a-zA-Z0-9_]*
//	* keys do not appear in ascending order.
func Must2LabelGauge(name, key1, key2 string) *Map2LabelGauge {
	mustValidName(name)
	mustValidKey(key1)
	mustValidKey(key2)
	if key1 > key2 {
		panic("metrics: label key arguments aren't sorted")
	}

	mutex.Lock()

	var g *Map2LabelGauge
	if index, ok := indices[name]; !ok {
		g = &Map2LabelGauge{name: name, labelKeys: [...]string{key1, key2}}

		indices[name] = uint32(len(metrics))
		metrics = append(metrics, &metric{
			typeComment: typePrefix + name + gaugeTypeLineEnd,
			typeID:      gaugeType,
			gaugeL2s:    []*Map2LabelGauge{g},
		})
	} else {
		m := metrics[index]
		if m.typeID != gaugeType {
			panic("metrics: name in use as another type")
		}

		for _, l2 := range m.gaugeL2s {
			if l2.labelKeys[0] == key1 && l2.labelKeys[1] == key2 {
				g = l2
				break
			}
		}
		if g == nil {
			g = &Map2LabelGauge{name: name, labelKeys: [...]string{key1, key2}}
			m.gaugeL2s = append(m.gaugeL2s, g)
		}
	}

	mutex.Unlock()

	return g
}

// Must3LabelGauge returns a composition with three fixed label keys.
// The function panics on any of the following:
//	* name in use as another metric type
//	* name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*
//	* a key does not match regular expression [a-zA-Z_][a-zA-Z0-9_]*
//	* keys do not appear in ascending order.
func Must3LabelGauge(name, key1, key2, key3 string) *Map3LabelGauge {
	mustValidName(name)
	mustValidKey(key1)
	mustValidKey(key2)
	mustValidKey(key3)
	if key1 > key2 || key2 > key3 {
		panic("metrics: label key arguments aren't sorted")
	}

	mutex.Lock()

	var g *Map3LabelGauge
	if index, ok := indices[name]; !ok {
		g = &Map3LabelGauge{name: name, labelKeys: [...]string{key1, key2, key3}}

		indices[name] = uint32(len(metrics))
		metrics = append(metrics, &metric{
			typeComment: typePrefix + name + gaugeTypeLineEnd,
			typeID:      gaugeType,
			gaugeL3s:    []*Map3LabelGauge{g},
		})
	} else {
		m := metrics[index]
		if m.typeID != gaugeType {
			panic("metrics: name in use as another type")
		}

		for _, l3 := range m.gaugeL3s {
			if l3.labelKeys[0] == key1 && l3.labelKeys[1] == key2 && l3.labelKeys[2] == key3 {
				g = l3
				break
			}
		}
		if g == nil {
			g = &Map3LabelGauge{name: name, labelKeys: [...]string{key1, key2, key3}}
			m.gaugeL3s = append(m.gaugeL3s, g)
		}
	}

	mutex.Unlock()

	return g
}

func mustValidKey(s string) {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '_' {
			continue
		}
		if i == 0 || c < '0' || c > '9' {
			panic("metrics: label key doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_]*")
		}
	}
}

var valueEscapes = strings.NewReplacer("\n", `\n`, `"`, `\"`, `\`, `\\`)

func format1Prefix(name, key, value string) string {
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

func format2Prefix(name string, keys, values *[2]string) string {
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

func format3Prefix(name string, keys, values *[3]string) string {
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
