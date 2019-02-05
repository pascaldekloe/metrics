package metrics

import (
	"strings"
	"sync"
)

// GaugeLabel1 is a gauge registration with a fixed label.
// Remember that every unique combination of key-value label pairs represents a
// new time series, which can dramatically increase the amount of data stored.
type GaugeLabel1 struct {
	name        string
	mutex       sync.Mutex
	labelKey    string
	labelValues []string
	gauges      []*Gauge
}

// GaugeLabel2 is a gauge registration with 2 fixed labels.
// Remember that every unique combination of key-value label pairs represents a
// new time series, which can dramatically increase the amount of data stored.
type GaugeLabel2 struct {
	name        string
	mutex       sync.Mutex
	labelKeys   [2]string
	labelValues []*[2]string
	gauges      []*Gauge
}

// GaugeLabel3 is a gauge registration with 3 fixed labels.
// Remember that every unique combination of key-value label pairs represents a
// new time series, which can dramatically increase the amount of data stored.
type GaugeLabel3 struct {
	name        string
	mutex       sync.Mutex
	labelKeys   [3]string
	labelValues []*[3]string
	gauges      []*Gauge
}

// ForLabel returns a dedicated Gauge for the respective label value.
// The value maps to the key as defined with MustGaugeLabel1.
func (g *GaugeLabel1) ForLabel(value string) *Gauge {
	g.mutex.Lock()

	for i, combo := range g.labelValues {
		if combo == value {
			g.mutex.Unlock()

			return g.gauges[i]
		}
	}

	g.labelValues = append(g.labelValues, value)
	entry := &Gauge{label: formatLabel1(g.name, g.labelKey, value)}
	g.gauges = append(g.gauges, entry)

	g.mutex.Unlock()

	return entry
}

// ForLabel returns a dedicated Gauge for the respective label values.
// The values map to the keys (in order) as defined with MustGaugeLabel2.
func (g *GaugeLabel2) ForLabels(value1, value2 string) *Gauge {
	g.mutex.Lock()

	for i, combo := range g.labelValues {
		if combo[0] == value1 && combo[1] == value2 {
			g.mutex.Unlock()

			return g.gauges[i]
		}
	}

	combo := [2]string{value1, value2}
	entry := &Gauge{label: formatLabel2(g.name, &g.labelKeys, &combo)}
	g.labelValues = append(g.labelValues, &combo)
	g.gauges = append(g.gauges, entry)

	g.mutex.Unlock()

	return entry
}

// ForLabel returns a dedicated Gauge for the respective label values.
// The values map to the keys (in order) as defined with MustGaugeLabel3.
func (g *GaugeLabel3) ForLabels(value1, value2, value3 string) *Gauge {
	g.mutex.Lock()

	for i, combo := range g.labelValues {
		if combo[0] == value1 && combo[1] == value2 && combo[2] == value3 {
			g.mutex.Unlock()

			return g.gauges[i]
		}
	}

	combo := [3]string{value1, value2, value3}
	entry := &Gauge{label: formatLabel3(g.name, &g.labelKeys, &combo)}
	g.labelValues = append(g.labelValues, &combo)
	g.gauges = append(g.gauges, entry)

	g.mutex.Unlock()

	return entry
}

// MustPlaceGaugeLabel1 registers a new GaugeLabel1 if the label key has not
// been used before on name.
//
// The function panics on any of the following:
//	* name in use as another metric type
//	* name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*
//	* label key does not match regular expression [a-zA-Z_][a-zA-Z0-9_]*
func MustPlaceGaugeLabel1(name, key string) *GaugeLabel1 {
	mustValidName(name)
	mustValidKey(key)

	mutex.Lock()

	var g *GaugeLabel1
	if index, ok := indices[name]; !ok {
		g = &GaugeLabel1{name: name, labelKey: key}

		indices[name] = uint32(len(metrics))
		metrics = append(metrics, &metric{
			typeComment: typePrefix + name + gaugeLineEnd,
			typeID:      gaugeType,
			gaugeL1s:    []*GaugeLabel1{g},
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
			g = &GaugeLabel1{name: name, labelKey: key}
			m.gaugeL1s = append(m.gaugeL1s, g)
		}
	}

	mutex.Unlock()

	return g
}

// MustPlaceGaugeLabel2 registers a new GaugeLabel2 if the label keys have
// not been used before on name.
//
// The function panics on any of the following:
//	* name in use as another metric type
//	* name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*
//	* label key does not match regular expression [a-zA-Z_][a-zA-Z0-9_]*
//	* label keys do not appear in ascending order
func MustPlaceGaugeLabel2(name, key1, key2 string) *GaugeLabel2 {
	mustValidName(name)
	mustValidKey(key1)
	mustValidKey(key2)
	if key1 > key2 {
		panic("metrics: label key arguments aren't sorted")
	}

	mutex.Lock()

	var g *GaugeLabel2
	if index, ok := indices[name]; !ok {
		g = &GaugeLabel2{name: name, labelKeys: [...]string{key1, key2}}

		indices[name] = uint32(len(metrics))
		metrics = append(metrics, &metric{
			typeComment: typePrefix + name + gaugeLineEnd,
			typeID:      gaugeType,
			gaugeL2s:    []*GaugeLabel2{g},
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
			g = &GaugeLabel2{name: name, labelKeys: [...]string{key1, key2}}
			m.gaugeL2s = append(m.gaugeL2s, g)
		}
	}

	mutex.Unlock()

	return g
}

// MustPlaceGaugeLabel3 registers a new GaugeLabel3 if the label keys have
// not been used before on name.
//
// The function panics on any of the following:
//	* name in use as another metric type
//	* name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*
//	* label key does not match regular expression [a-zA-Z_][a-zA-Z0-9_]*
//	* label keys do not appear in ascending order.
func MustPlaceGaugeLabel3(name, key1, key2, key3 string) *GaugeLabel3 {
	mustValidName(name)
	mustValidKey(key1)
	mustValidKey(key2)
	mustValidKey(key3)
	if key1 > key2 || key2 > key3 {
		panic("metrics: label key arguments aren't sorted")
	}

	mutex.Lock()

	var g *GaugeLabel3
	if index, ok := indices[name]; !ok {
		g = &GaugeLabel3{name: name, labelKeys: [...]string{key1, key2, key3}}

		indices[name] = uint32(len(metrics))
		metrics = append(metrics, &metric{
			typeComment: typePrefix + name + gaugeLineEnd,
			typeID:      gaugeType,
			gaugeL3s:    []*GaugeLabel3{g},
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
			g = &GaugeLabel3{name: name, labelKeys: [...]string{key1, key2, key3}}
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

func formatLabel1(name, key, value string) string {
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

func formatLabel2(name string, keys, values *[2]string) string {
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

func formatLabel3(name string, keys, values *[3]string) string {
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
