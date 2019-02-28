package metrics

import (
	"strings"
	"sync"
)

// Metric is a named record.
type metric struct {
	typeComment string
	helpComment string

	counter   *Counter
	integer   *Integer
	real      *Real
	histogram *Histogram
	sample    *Sample

	counterL1s   []*map1LabelCounter
	counterL2s   []*map2LabelCounter
	counterL3s   []*map3LabelCounter
	integerL1s   []*map1LabelInteger
	integerL2s   []*map2LabelInteger
	integerL3s   []*map3LabelInteger
	realL1s      []*map1LabelReal
	realL2s      []*map2LabelReal
	realL3s      []*map3LabelReal
	histogramL1s []*map1LabelHistogram
	histogramL2s []*map2LabelHistogram
	sampleL1s    []*map1LabelSample
	sampleL2s    []*map2LabelSample
	sampleL3s    []*map3LabelSample
}

func (m *metric) typeID() byte {
	return m.typeComment[len(m.typeComment)-2]
}

var std = NewRegister()

// Register is a metric bundle.
type Register struct {
	mutex sync.RWMutex
	// mapping by name
	indices map[string]uint32
	// consistent order
	metrics []*metric
}

// NewRegister returns an empty registration instance for non-standard usecases.
func NewRegister() *Register {
	return &Register{indices: make(map[string]uint32)}
}

// MustCounter registers a new Counter. Registration panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Combinations with Sample and the various
// label options are allowed though. The Sample is ignored once a Counter is
// registered under the same name. This fallback allows warm starts.
func MustCounter(name string) *Counter {
	return std.MustCounter(name)
}

// MustCounter registers a new Counter. Registration panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Combinations with Sample and the various
// label options are allowed though. The Sample is ignored once a Counter is
// registered under the same name. This fallback allows warm starts.
func (r *Register) MustCounter(name string) *Counter {
	mustValidName(name)

	r.mutex.Lock()

	var m *metric
	if index, ok := r.indices[name]; ok {
		m = r.metrics[index]
		if m.typeID() != counterType || m.counter != nil {
			panic("metrics: name already in use")
		}
	} else {
		r.indices[name] = uint32(len(r.metrics))
		m = &metric{typeComment: typePrefix + name + counterTypeLineEnd}
		r.metrics = append(r.metrics, m)
	}

	m.counter = &Counter{prefix: name + " "}

	r.mutex.Unlock()

	return m.counter
}

// MustInteger registers a new gauge. Registration panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Combinations with Sample and the various
// label options are allowed though. The Sample is ignored once a gauge
// is registered under the same name. This fallback allows warm starts.
func MustInteger(name string) *Integer {
	return std.MustInteger(name)
}

// MustInteger registers a new gauge. Registration panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Combinations with Sample and the various
// label options are allowed though. The Sample is ignored once a gauge
// is registered under the same name. This fallback allows warm starts.
func (r *Register) MustInteger(name string) *Integer {
	mustValidName(name)

	r.mutex.Lock()

	var m *metric
	if index, ok := r.indices[name]; ok {
		m = r.metrics[index]
		if m.typeID() != gaugeType || m.integer != nil || m.real != nil {
			panic("metrics: name already in use")
		}
	} else {
		r.indices[name] = uint32(len(r.metrics))
		m = &metric{typeComment: typePrefix + name + gaugeTypeLineEnd}
		r.metrics = append(r.metrics, m)
	}

	m.integer = &Integer{prefix: name + " "}

	r.mutex.Unlock()

	return m.integer
}

// MustReal registers a new gauge. Registration panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Combinations with Sample and the various
// label options are allowed though. The Sample is ignored once a gauge
// is registered under the same name. This fallback allows warm starts.
func MustReal(name string) *Real {
	return std.MustReal(name)
}

// MustReal registers a new gauge. Registration panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Combinations with Sample and the various
// label options are allowed though. The Sample is ignored once a gauge
// is registered under the same name. This fallback allows warm starts.
func (r *Register) MustReal(name string) *Real {
	mustValidName(name)

	r.mutex.Lock()

	var m *metric
	if index, ok := r.indices[name]; ok {
		m = r.metrics[index]
		if m.typeID() != gaugeType || m.integer != nil || m.real != nil {
			panic("metrics: name already in use")
		}
	} else {
		r.indices[name] = uint32(len(r.metrics))
		m = &metric{typeComment: typePrefix + name + gaugeTypeLineEnd}
		r.metrics = append(r.metrics, m)
	}

	m.real = &Real{prefix: name + " "}

	r.mutex.Unlock()

	return m.real
}

// MustHistogram registers a new Histogram. Buckets define the upper
// boundaries, preferably in ascending order. Special cases not-a-number
// and both infinities are ignored.
// Registration panics when name was registered before, or when name
// doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*.
func MustHistogram(name string, buckets ...float64) *Histogram {
	return std.MustHistogram(name, buckets...)
}

// MustHistogram registers a new Histogram. Buckets define the upper
// boundaries, preferably in ascending order. Special cases not-a-number
// and both infinities are ignored.
// Registration panics when name was registered before, or when name
// doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*.
func (r *Register) MustHistogram(name string, buckets ...float64) *Histogram {
	mustValidName(name)

	h := newHistogram(name, buckets)
	h.prefix(name)

	r.mutex.Lock()

	var m *metric
	if index, ok := r.indices[name]; ok {
		m = r.metrics[index]
		if m.typeID() != histogramType || m.histogram != nil {
			panic("metrics: name already in use")
		}
	} else {
		r.indices[name] = uint32(len(r.metrics))
		m = &metric{typeComment: typePrefix + name + histogramTypeLineEnd}
		r.metrics = append(r.metrics, m)
	}

	m.histogram = h

	r.mutex.Unlock()

	return h
}

// MustGaugeSample registers a new Sample. Registration panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Combinations with Sample and the various
// label options are allowed though. The Sample is ignored once a Gauge is
// registered under the same name. This fallback allows warm starts.
func MustGaugeSample(name string) *Sample {
	return MustGaugeSample(name)
}

// MustGaugeSample registers a new Sample. Registration panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Combinations with Sample and the various
// label options are allowed though. The Sample is ignored once a Gauge is
// registered under the same name. This fallback allows warm starts.
func (r *Register) MustGaugeSample(name string) *Sample {
	mustValidName(name)

	r.mutex.Lock()

	var m *metric
	if index, ok := r.indices[name]; ok {
		m = r.metrics[index]
		if m.typeID() != gaugeType || m.sample != nil {
			panic("metrics: name already in use")
		}
	} else {
		r.indices[name] = uint32(len(r.metrics))
		m = &metric{typeComment: typePrefix + name + gaugeTypeLineEnd}
		r.metrics = append(r.metrics, m)
	}

	m.sample = &Sample{prefix: name + " "}

	r.mutex.Unlock()

	return m.sample
}

// MustCounterSample registers a new Sample. Registration panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Combinations with Sample and the various
// label options are allowed though. The Sample is ignored once a Counter is
// registered under the same name. This fallback allows warm starts.
func MustCounterSample(name string) *Sample {
	return std.MustCounterSample(name)
}

// MustCounterSample registers a new Sample. Registration panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Combinations with Sample and the various
// label options are allowed though. The Sample is ignored once a Counter is
// registered under the same name. This fallback allows warm starts.
func (r *Register) MustCounterSample(name string) *Sample {
	mustValidName(name)

	r.mutex.Lock()

	var m *metric
	if index, ok := r.indices[name]; ok {
		m = r.metrics[index]
		if m.typeID() != counterType || m.sample != nil {
			panic("metrics: name already in use")
		}
	} else {
		r.indices[name] = uint32(len(r.metrics))
		m = &metric{typeComment: typePrefix + name + counterTypeLineEnd}
		r.metrics = append(r.metrics, m)
	}

	m.sample = &Sample{prefix: name + " "}

	r.mutex.Unlock()

	return m.sample
}

// Must1LabelCounter returns a function which assigns dedicated Counter
// instances to each label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Counter represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) labelName does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) labelName is already in use.
func Must1LabelCounter(name, labelName string) func(labelValue string) *Counter {
	return std.Must1LabelCounter(name, labelName)
}

// Must1LabelCounter returns a function which assigns dedicated Counter
// instances to each label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Counter represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) labelName does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) labelName is already in use.
func (r *Register) Must1LabelCounter(name, labelName string) func(labelValue string) *Counter {
	mustValidName(name)
	mustValidLabelName(labelName)

	r.mutex.Lock()

	var l1 *map1LabelCounter
	if index, ok := r.indices[name]; !ok {
		l1 = &map1LabelCounter{map1Label: map1Label{
			name: name, labelName: labelName}}

		r.indices[name] = uint32(len(r.metrics))
		r.metrics = append(r.metrics, &metric{
			typeComment: typePrefix + name + counterTypeLineEnd,
			counterL1s:  []*map1LabelCounter{l1},
		})
	} else {
		m := r.metrics[index]
		if m.typeID() != counterType {
			panic("metrics: name in use as another type")
		}

		for _, o := range m.counterL1s {
			if o.labelName == labelName {
				panic("metrics: label already in use")
			}
		}
		l1 = &map1LabelCounter{map1Label: map1Label{
			name: name, labelName: labelName}}
		m.counterL1s = append(m.counterL1s, l1)
	}

	r.mutex.Unlock()
	return l1.with
}

// Must2LabelCounter returns a function which assigns dedicated Counter
// instances to each label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Counter represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) the label names do not match regular expression [a-zA-Z_][a-zA-Z0-9_]*,
// (4) the label names do not appear in ascending order or
// (5) the label names are already in use.
func Must2LabelCounter(name, label1Name, label2Name string) func(label1Value, label2Value string) *Counter {
	return std.Must2LabelCounter(name, label1Name, label2Name)
}

// Must2LabelCounter returns a function which assigns dedicated Counter
// instances to each label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Counter represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) the label names do not match regular expression [a-zA-Z_][a-zA-Z0-9_]*,
// (4) the label names do not appear in ascending order or
// (5) the label names are already in use.
func (r *Register) Must2LabelCounter(name, label1Name, label2Name string) func(label1Value, label2Value string) *Counter {
	mustValidName(name)
	mustValidLabelName(label1Name)
	mustValidLabelName(label2Name)
	if label1Name > label2Name {
		panic("metrics: label name arguments aren't sorted")
	}

	r.mutex.Lock()

	var l2 *map2LabelCounter
	if index, ok := r.indices[name]; !ok {
		l2 = &map2LabelCounter{map2Label: map2Label{
			name: name, labelNames: [...]string{label1Name, label2Name}}}

		r.indices[name] = uint32(len(r.metrics))
		r.metrics = append(r.metrics, &metric{
			typeComment: typePrefix + name + counterTypeLineEnd,
			counterL2s:  []*map2LabelCounter{l2},
		})
	} else {
		m := r.metrics[index]
		if m.typeID() != counterType {
			panic("metrics: name in use as another type")
		}

		for _, o := range m.counterL2s {
			if o.labelNames[0] == label1Name && o.labelNames[1] == label2Name {
				panic("metrics: labels already in use")
			}
		}
		l2 = &map2LabelCounter{map2Label: map2Label{
			name: name, labelNames: [...]string{label1Name, label2Name}}}
		m.counterL2s = append(m.counterL2s, l2)
	}

	r.mutex.Unlock()
	return l2.with
}

// Must3LabelCounter returns a function which assigns dedicated Counter
// instances to each label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Counter represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) the label names do not match regular expression [a-zA-Z_][a-zA-Z0-9_]*,
// (4) the label names do not appear in ascending order or
// (5) the label names are already in use.
func Must3LabelCounter(name, label1Name, label2Name, label3Name string) func(label1Value, label2Value, label3Value string) *Counter {
	return std.Must3LabelCounter(name, label1Name, label2Name, label3Name)
}

// Must3LabelCounter returns a function which assigns dedicated Counter
// instances to each label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Counter represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) the label names do not match regular expression [a-zA-Z_][a-zA-Z0-9_]*,
// (4) the label names do not appear in ascending order or
// (5) the label names are already in use.
func (r *Register) Must3LabelCounter(name, label1Name, label2Name, label3Name string) func(label1Value, label2Value, label3Value string) *Counter {
	mustValidName(name)
	mustValidLabelName(label1Name)
	mustValidLabelName(label2Name)
	mustValidLabelName(label3Name)
	if label1Name > label2Name || label2Name > label3Name {
		panic("metrics: label name arguments aren't sorted")
	}

	r.mutex.Lock()

	var l3 *map3LabelCounter
	if index, ok := r.indices[name]; !ok {
		l3 = &map3LabelCounter{map3Label: map3Label{
			name: name, labelNames: [...]string{label1Name, label2Name, label3Name}}}

		r.indices[name] = uint32(len(r.metrics))
		r.metrics = append(r.metrics, &metric{
			typeComment: typePrefix + name + counterTypeLineEnd,
			counterL3s:  []*map3LabelCounter{l3},
		})
	} else {
		m := r.metrics[index]
		if m.typeID() != counterType {
			panic("metrics: name in use as another type")
		}

		for _, o := range m.counterL3s {
			if o.labelNames[0] == label1Name && o.labelNames[1] == label2Name && o.labelNames[2] == label3Name {
				panic("metrics: labels already in use")
			}
		}
		l3 = &map3LabelCounter{map3Label: map3Label{
			name: name, labelNames: [...]string{label1Name, label2Name, label3Name}}}
		m.counterL3s = append(m.counterL3s, l3)
	}

	r.mutex.Unlock()
	return l3.with
}

// Must1LabelInteger returns a function which assigns dedicated Integer
// instances to each label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Integer represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) labelName does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) labelName is already in use.
func Must1LabelInteger(name, labelName string) func(labelValue string) *Integer {
	return std.Must1LabelInteger(name, labelName)
}

// Must1LabelInteger returns a function which assigns dedicated Integer
// instances to each label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Integer represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) labelName does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) labelName is already in use.
func (r *Register) Must1LabelInteger(name, labelName string) func(labelValue string) *Integer {
	mustValidName(name)
	mustValidLabelName(labelName)

	r.mutex.Lock()

	var l1 *map1LabelInteger
	if index, ok := r.indices[name]; !ok {
		l1 = &map1LabelInteger{map1Label: map1Label{
			name: name, labelName: labelName}}

		r.indices[name] = uint32(len(r.metrics))
		r.metrics = append(r.metrics, &metric{
			typeComment: typePrefix + name + gaugeTypeLineEnd,
			integerL1s:  []*map1LabelInteger{l1},
		})
	} else {
		m := r.metrics[index]
		if m.typeID() != gaugeType || m.real != nil || len(m.realL1s) != 0 || len(m.realL2s) != 0 || len(m.realL3s) != 0 {
			panic("metrics: name in use as another type")
		}

		for _, o := range m.integerL1s {
			if o.labelName == labelName {
				panic("metrics: label already in use")
			}
		}
		l1 = &map1LabelInteger{map1Label: map1Label{
			name: name, labelName: labelName}}
		m.integerL1s = append(m.integerL1s, l1)
	}

	r.mutex.Unlock()
	return l1.with
}

// Must2LabelInteger returns a function which assigns dedicated Integer
// instances to each label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Integer represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) the label names do not match regular expression [a-zA-Z_][a-zA-Z0-9_]*,
// (4) the label names do not appear in ascending order or
// (5) the label names are already in use.
func Must2LabelInteger(name, label1Name, label2Name string) func(label1Value, label2Value string) *Integer {
	return std.Must2LabelInteger(name, label1Name, label2Name)
}

// Must2LabelInteger returns a function which assigns dedicated Integer
// instances to each label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Integer represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) the label names do not match regular expression [a-zA-Z_][a-zA-Z0-9_]*,
// (4) the label names do not appear in ascending order or
// (5) the label names are already in use.
func (r *Register) Must2LabelInteger(name, label1Name, label2Name string) func(label1Value, label2Value string) *Integer {
	mustValidName(name)
	mustValidLabelName(label1Name)
	mustValidLabelName(label2Name)
	if label1Name > label2Name {
		panic("metrics: label name arguments aren't sorted")
	}

	r.mutex.Lock()

	var l2 *map2LabelInteger
	if index, ok := r.indices[name]; !ok {
		l2 = &map2LabelInteger{map2Label: map2Label{
			name: name, labelNames: [...]string{label1Name, label2Name}}}

		r.indices[name] = uint32(len(r.metrics))
		r.metrics = append(r.metrics, &metric{
			typeComment: typePrefix + name + gaugeTypeLineEnd,
			integerL2s:  []*map2LabelInteger{l2},
		})
	} else {
		m := r.metrics[index]
		if m.typeID() != gaugeType || m.real != nil || len(m.realL1s) != 0 || len(m.realL2s) != 0 || len(m.realL3s) != 0 {
			panic("metrics: name in use as another type")
		}

		for _, o := range m.integerL2s {
			if o.labelNames[0] == label1Name && o.labelNames[1] == label2Name {
				panic("metrics: labels already in use")
			}
		}
		l2 = &map2LabelInteger{map2Label: map2Label{
			name: name, labelNames: [...]string{label1Name, label2Name}}}
		m.integerL2s = append(m.integerL2s, l2)
	}

	r.mutex.Unlock()
	return l2.with
}

// Must3LabelInteger returns a function which assigns dedicated Integer
// instances to each label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Integer represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) the label names do not match regular expression [a-zA-Z_][a-zA-Z0-9_]*,
// (4) the label names do not appear in ascending order or
// (5) the label names are already in use.
func Must3LabelInteger(name, label1Name, label2Name, label3Name string) func(label1Value, label2Value, label3Value string) *Integer {
	return std.Must3LabelInteger(name, label1Name, label2Name, label3Name)
}

// Must3LabelInteger returns a function which assigns dedicated Integer
// instances to each label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Integer represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) the label names do not match regular expression [a-zA-Z_][a-zA-Z0-9_]*,
// (4) the label names do not appear in ascending order or
// (5) the label names are already in use.
func (r *Register) Must3LabelInteger(name, label1Name, label2Name, label3Name string) func(label1Value, label2Value, label3Value string) *Integer {
	mustValidName(name)
	mustValidLabelName(label1Name)
	mustValidLabelName(label2Name)
	mustValidLabelName(label3Name)
	if label1Name > label2Name || label2Name > label3Name {
		panic("metrics: label name arguments aren't sorted")
	}

	r.mutex.Lock()

	var l3 *map3LabelInteger
	if index, ok := r.indices[name]; !ok {
		l3 = &map3LabelInteger{map3Label: map3Label{
			name: name, labelNames: [...]string{label1Name, label2Name, label3Name}}}

		r.indices[name] = uint32(len(r.metrics))
		r.metrics = append(r.metrics, &metric{
			typeComment: typePrefix + name + gaugeTypeLineEnd,
			integerL3s:  []*map3LabelInteger{l3},
		})
	} else {
		m := r.metrics[index]
		if m.typeID() != gaugeType || m.real != nil || len(m.realL1s) != 0 || len(m.realL2s) != 0 || len(m.realL3s) != 0 {
			panic("metrics: name in use as another type")
		}

		for _, o := range m.integerL3s {
			if o.labelNames[0] == label1Name && o.labelNames[1] == label2Name && o.labelNames[2] == label3Name {
				panic("metrics: labels already in use")
			}
		}
		l3 = &map3LabelInteger{map3Label: map3Label{
			name: name, labelNames: [...]string{label1Name, label2Name, label3Name}}}
		m.integerL3s = append(m.integerL3s, l3)
	}

	r.mutex.Unlock()
	return l3.with
}

// Must1LabelReal returns a function which assigns dedicated Real
// instances to each label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Real represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) labelName does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) labelName is already in use.
func Must1LabelReal(name, labelName string) func(labelValue string) *Real {
	return std.Must1LabelReal(name, labelName)
}

// Must1LabelReal returns a function which assigns dedicated Real
// instances to each label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Real represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) labelName does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) labelName is already in use.
func (r *Register) Must1LabelReal(name, labelName string) func(labelValue string) *Real {
	mustValidName(name)
	mustValidLabelName(labelName)

	r.mutex.Lock()

	var l1 *map1LabelReal
	if index, ok := r.indices[name]; !ok {
		l1 = &map1LabelReal{map1Label: map1Label{
			name: name, labelName: labelName}}

		r.indices[name] = uint32(len(r.metrics))
		r.metrics = append(r.metrics, &metric{
			typeComment: typePrefix + name + gaugeTypeLineEnd,
			realL1s:     []*map1LabelReal{l1},
		})
	} else {
		m := r.metrics[index]
		if m.typeID() != gaugeType || m.integer != nil || len(m.integerL1s) != 0 || len(m.integerL2s) != 0 || len(m.integerL3s) != 0 {
			panic("metrics: name in use as another type")
		}

		for _, o := range m.realL1s {
			if o.labelName == labelName {
				panic("metrics: label already in use")
			}
		}
		l1 = &map1LabelReal{map1Label: map1Label{
			name: name, labelName: labelName}}
		m.realL1s = append(m.realL1s, l1)
	}

	r.mutex.Unlock()
	return l1.with
}

// Must2LabelReal returns a function which assigns dedicated Real
// instances to each label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Real represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) the label names do not match regular expression [a-zA-Z_][a-zA-Z0-9_]*,
// (4) the label names do not appear in ascending order or
// (5) the label names are already in use.
func Must2LabelReal(name, label1Name, label2Name string) func(label1Value, label2Value string) *Real {
	return std.Must2LabelReal(name, label1Name, label2Name)
}

// Must2LabelReal returns a function which assigns dedicated Real
// instances to each label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Real represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) the label names do not match regular expression [a-zA-Z_][a-zA-Z0-9_]*,
// (4) the label names do not appear in ascending order or
// (5) the label names are already in use.
func (r *Register) Must2LabelReal(name, label1Name, label2Name string) func(label1Value, label2Value string) *Real {
	mustValidName(name)
	mustValidLabelName(label1Name)
	mustValidLabelName(label2Name)
	if label1Name > label2Name {
		panic("metrics: label name arguments aren't sorted")
	}

	r.mutex.Lock()

	var l2 *map2LabelReal
	if index, ok := r.indices[name]; !ok {
		l2 = &map2LabelReal{map2Label: map2Label{
			name: name, labelNames: [...]string{label1Name, label2Name}}}

		r.indices[name] = uint32(len(r.metrics))
		r.metrics = append(r.metrics, &metric{
			typeComment: typePrefix + name + gaugeTypeLineEnd,
			realL2s:     []*map2LabelReal{l2},
		})
	} else {
		m := r.metrics[index]
		if m.typeID() != gaugeType || m.integer != nil || len(m.integerL1s) != 0 || len(m.integerL2s) != 0 || len(m.integerL3s) != 0 {
			panic("metrics: name in use as another type")
		}

		for _, o := range m.realL2s {
			if o.labelNames[0] == label1Name && o.labelNames[1] == label2Name {
				panic("metrics: labels already in use")
			}
		}
		l2 = &map2LabelReal{map2Label: map2Label{
			name: name, labelNames: [...]string{label1Name, label2Name}}}
		m.realL2s = append(m.realL2s, l2)
	}

	r.mutex.Unlock()
	return l2.with
}

// Must3LabelReal returns a function which assigns dedicated Real
// instances to each label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Real represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) the label names do not match regular expression [a-zA-Z_][a-zA-Z0-9_]*,
// (4) the label names do not appear in ascending order or
// (5) the label names are already in use.
func Must3LabelReal(name, label1Name, label2Name, label3Name string) func(label1Value, label2Value, label3Value string) *Real {
	return std.Must3LabelReal(name, label1Name, label2Name, label3Name)
}

// Must3LabelReal returns a function which assigns dedicated Real
// instances to each label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Real represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) the label names do not match regular expression [a-zA-Z_][a-zA-Z0-9_]*,
// (4) the label names do not appear in ascending order or
// (5) the label names are already in use.
func (r *Register) Must3LabelReal(name, label1Name, label2Name, label3Name string) func(label1Value, label2Value, label3Value string) *Real {
	mustValidName(name)
	mustValidLabelName(label1Name)
	mustValidLabelName(label2Name)
	mustValidLabelName(label3Name)
	if label1Name > label2Name || label2Name > label3Name {
		panic("metrics: label name arguments aren't sorted")
	}

	r.mutex.Lock()

	var l3 *map3LabelReal
	if index, ok := r.indices[name]; !ok {
		l3 = &map3LabelReal{map3Label: map3Label{
			name: name, labelNames: [...]string{label1Name, label2Name, label3Name}}}

		r.indices[name] = uint32(len(r.metrics))
		r.metrics = append(r.metrics, &metric{
			typeComment: typePrefix + name + gaugeTypeLineEnd,
			realL3s:     []*map3LabelReal{l3},
		})
	} else {
		m := r.metrics[index]
		if m.typeID() != gaugeType || m.integer != nil || len(m.integerL1s) != 0 || len(m.integerL2s) != 0 || len(m.integerL3s) != 0 {
			panic("metrics: name in use as another type")
		}

		for _, o := range m.realL3s {
			if o.labelNames[0] == label1Name && o.labelNames[1] == label2Name && o.labelNames[2] == label3Name {
				panic("metrics: labels already in use")
			}
		}
		l3 = &map3LabelReal{map3Label: map3Label{
			name: name, labelNames: [...]string{label1Name, label2Name, label3Name}}}
		m.realL3s = append(m.realL3s, l3)
	}

	r.mutex.Unlock()
	return l3.with
}

// Must1LabelHistogram returns a function which assigns dedicated Histogram
// instances to each label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Histogram represents a new time
// series, which can dramatically increase the amount of data stored.
// Buckets define the upper boundaries, preferably in ascending order.
// Special cases not-a-number and both infinities are ignored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) labelName does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) labelName is already in use.
func Must1LabelHistogram(name, labelName string, buckets ...float64) func(labelValue string) *Histogram {
	return std.Must1LabelHistogram(name, labelName, buckets...)
}

// Must1LabelHistogram returns a function which assigns dedicated Histogram
// instances to each label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Histogram represents a new time
// series, which can dramatically increase the amount of data stored.
// Buckets define the upper boundaries, preferably in ascending order.
// Special cases not-a-number and both infinities are ignored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) labelName does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) labelName is already in use.
func (r *Register) Must1LabelHistogram(name, labelName string, buckets ...float64) func(labelValue string) *Histogram {
	mustValidName(name)
	mustValidLabelName(labelName)

	r.mutex.Lock()

	var l1 *map1LabelHistogram
	if index, ok := r.indices[name]; !ok {
		l1 = &map1LabelHistogram{buckets: buckets, map1Label: map1Label{
			name: name, labelName: labelName}}

		r.indices[name] = uint32(len(r.metrics))
		r.metrics = append(r.metrics, &metric{
			typeComment:  typePrefix + name + histogramTypeLineEnd,
			histogramL1s: []*map1LabelHistogram{l1},
		})
	} else {
		m := r.metrics[index]
		if m.typeID() != histogramType {
			panic("metrics: name in use as another type")
		}

		for _, o := range m.histogramL1s {
			if o.labelName == labelName {
				panic("metrics: label already in use")
			}
		}
		l1 = &map1LabelHistogram{buckets: buckets, map1Label: map1Label{
			name: name, labelName: labelName}}
		m.histogramL1s = append(m.histogramL1s, l1)
	}

	r.mutex.Unlock()
	return l1.with
}

// Must2LabelHistogram returns a function which assigns dedicated Histogram
// instances to each label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Histogram represents a new time
// series, which can dramatically increase the amount of data stored.
// Buckets define the upper boundaries, preferably in ascending order.
// Special cases not-a-number and both infinities are ignored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) the label names do not match regular expression [a-zA-Z_][a-zA-Z0-9_]*,
// (4) the label names do not appear in ascending order or
// (5) the label names are already in use.
func Must2LabelHistogram(name, label1Name, label2Name string, buckets ...float64) func(label1Value, label2Value string) *Histogram {
	return std.Must2LabelHistogram(name, label1Name, label2Name, buckets...)
}

// Must2LabelHistogram returns a function which assigns dedicated Histogram
// instances to each label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Histogram represents a new time
// series, which can dramatically increase the amount of data stored.
// Buckets define the upper boundaries, preferably in ascending order.
// Special cases not-a-number and both infinities are ignored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) the label names do not match regular expression [a-zA-Z_][a-zA-Z0-9_]*,
// (4) the label names do not appear in ascending order or
// (5) the label names are already in use.
func (r *Register) Must2LabelHistogram(name, label1Name, label2Name string, buckets ...float64) func(label1Value, label2Value string) *Histogram {
	mustValidName(name)
	mustValidLabelName(label1Name)
	mustValidLabelName(label2Name)
	if label1Name > label2Name {
		panic("metrics: label name arguments aren't sorted")
	}

	r.mutex.Lock()

	var l2 *map2LabelHistogram
	if index, ok := r.indices[name]; !ok {
		l2 = &map2LabelHistogram{buckets: buckets, map2Label: map2Label{
			name: name, labelNames: [...]string{label1Name, label2Name}}}

		r.indices[name] = uint32(len(r.metrics))
		r.metrics = append(r.metrics, &metric{
			typeComment:  typePrefix + name + histogramTypeLineEnd,
			histogramL2s: []*map2LabelHistogram{l2},
		})
	} else {
		m := r.metrics[index]
		if m.typeID() != histogramType {
			panic("metrics: name in use as another type")
		}

		for _, o := range m.histogramL2s {
			if o.labelNames[0] == label1Name && o.labelNames[1] == label2Name {
				panic("metrics: labels already in use")
			}
		}
		l2 = &map2LabelHistogram{buckets: buckets, map2Label: map2Label{
			name: name, labelNames: [...]string{label1Name, label2Name}}}
		m.histogramL2s = append(m.histogramL2s, l2)
	}

	r.mutex.Unlock()
	return l2.with
}

// Must1LabelCounterSample returns a function which assigns dedicated Sample
// instances to each label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Sample represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) labelName does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) labelName is already in use.
func Must1LabelCounterSample(name, labelName string) func(labelValue string) *Sample {
	return std.Must1LabelCounterSample(name, labelName)
}

// Must1LabelCounterSample returns a function which assigns dedicated Sample
// instances to each label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Sample represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) labelName does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) labelName is already in use.
func (r *Register) Must1LabelCounterSample(name, labelName string) func(labelValue string) *Sample {
	mustValidName(name)
	mustValidLabelName(labelName)

	r.mutex.Lock()

	var l1 *map1LabelSample
	if index, ok := r.indices[name]; !ok {
		l1 = &map1LabelSample{map1Label: map1Label{
			name: name, labelName: labelName}}

		r.indices[name] = uint32(len(r.metrics))
		r.metrics = append(r.metrics, &metric{
			typeComment: typePrefix + name + counterTypeLineEnd,
			sampleL1s:   []*map1LabelSample{l1},
		})
	} else {
		m := r.metrics[index]
		if m.typeID() != counterType {
			panic("metrics: name in use as another type")
		}

		for _, o := range m.sampleL1s {
			if o.labelName == labelName {
				panic("metrics: label already in use")
			}
		}
		l1 = &map1LabelSample{map1Label: map1Label{
			name: name, labelName: labelName}}
		m.sampleL1s = append(m.sampleL1s, l1)
	}

	r.mutex.Unlock()
	return l1.with
}

// Must2LabelCounterSample returns a function which assigns dedicated Sample
// instances to each label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Sample represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) the label names do not match regular expression [a-zA-Z_][a-zA-Z0-9_]*,
// (4) the label names do not appear in ascending order or
// (5) the label names are already in use.
func Must2LabelCounterSample(name, label1Name, label2Name string) func(label1Value, label2Value string) *Sample {
	return std.Must2LabelCounterSample(name, label1Name, label2Name)
}

// Must2LabelCounterSample returns a function which assigns dedicated Sample
// instances to each label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Sample represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) the label names do not match regular expression [a-zA-Z_][a-zA-Z0-9_]*,
// (4) the label names do not appear in ascending order or
// (5) the label names are already in use.
func (r *Register) Must2LabelCounterSample(name, label1Name, label2Name string) func(label1Value, label2Value string) *Sample {
	mustValidName(name)
	mustValidLabelName(label1Name)
	mustValidLabelName(label2Name)
	if label1Name > label2Name {
		panic("metrics: label name arguments aren't sorted")
	}

	r.mutex.Lock()

	var l2 *map2LabelSample
	if index, ok := r.indices[name]; !ok {
		l2 = &map2LabelSample{map2Label: map2Label{
			name: name, labelNames: [...]string{label1Name, label2Name}}}

		r.indices[name] = uint32(len(r.metrics))
		r.metrics = append(r.metrics, &metric{
			typeComment: typePrefix + name + counterTypeLineEnd,
			sampleL2s:   []*map2LabelSample{l2},
		})
	} else {
		m := r.metrics[index]
		if m.typeID() != counterType {
			panic("metrics: name in use as another type")
		}

		for _, o := range m.sampleL2s {
			if o.labelNames[0] == label1Name && o.labelNames[1] == label2Name {
				panic("metrics: labels already in use")
			}
		}
		l2 = &map2LabelSample{map2Label: map2Label{
			name: name, labelNames: [...]string{label1Name, label2Name}}}
		m.sampleL2s = append(m.sampleL2s, l2)
	}

	r.mutex.Unlock()
	return l2.with
}

// Must3LabelCounterSample returns a function which assigns dedicated Sample
// instances to each label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Sample represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) the label names do not match regular expression [a-zA-Z_][a-zA-Z0-9_]*,
// (4) the label names do not appear in ascending order or
// (5) the label names are already in use.
func Must3LabelCounterSample(name, label1Name, label2Name, label3Name string) func(label1Value, label2Value, label3Value string) *Sample {
	return std.Must3LabelCounterSample(name, label1Name, label2Name, label3Name)
}

// Must3LabelCounterSample returns a function which assigns dedicated Sample
// instances to each label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Sample represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) the label names do not match regular expression [a-zA-Z_][a-zA-Z0-9_]*,
// (4) the label names do not appear in ascending order or
// (5) the label names are already in use.
func (r *Register) Must3LabelCounterSample(name, label1Name, label2Name, label3Name string) func(label1Value, label2Value, label3Value string) *Sample {
	mustValidName(name)
	mustValidLabelName(label1Name)
	mustValidLabelName(label2Name)
	mustValidLabelName(label3Name)
	if label1Name > label2Name || label2Name > label3Name {
		panic("metrics: label name arguments aren't sorted")
	}

	r.mutex.Lock()

	var l3 *map3LabelSample
	if index, ok := r.indices[name]; !ok {
		l3 = &map3LabelSample{map3Label: map3Label{
			name: name, labelNames: [...]string{label1Name, label2Name, label3Name}}}

		r.indices[name] = uint32(len(r.metrics))
		r.metrics = append(r.metrics, &metric{
			typeComment: typePrefix + name + counterTypeLineEnd,
			sampleL3s:   []*map3LabelSample{l3},
		})
	} else {
		m := r.metrics[index]
		if m.typeID() != counterType {
			panic("metrics: name in use as another type")
		}

		for _, o := range m.sampleL3s {
			if o.labelNames[0] == label1Name && o.labelNames[1] == label2Name && o.labelNames[2] == label3Name {
				panic("metrics: labels already in use")
			}
		}
		l3 = &map3LabelSample{map3Label: map3Label{
			name: name, labelNames: [...]string{label1Name, label2Name, label3Name}}}
		m.sampleL3s = append(m.sampleL3s, l3)
	}

	r.mutex.Unlock()
	return l3.with
}

// Must1LabelGaugeSample returns a function which assigns dedicated Sample
// instances to each label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Sample represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) labelName does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) labelName is already in use.
func Must1LabelGaugeSample(name, labelName string) func(labelValue string) *Sample {
	return std.Must1LabelGaugeSample(name, labelName)
}

// Must1LabelGaugeSample returns a function which assigns dedicated Sample
// instances to each label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Sample represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) labelName does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) labelName is already in use.
func (r *Register) Must1LabelGaugeSample(name, labelName string) func(labelValue string) *Sample {
	mustValidName(name)
	mustValidLabelName(labelName)

	r.mutex.Lock()

	var l1 *map1LabelSample
	if index, ok := r.indices[name]; !ok {
		l1 = &map1LabelSample{map1Label: map1Label{
			name: name, labelName: labelName}}

		r.indices[name] = uint32(len(r.metrics))
		r.metrics = append(r.metrics, &metric{
			typeComment: typePrefix + name + gaugeTypeLineEnd,
			sampleL1s:   []*map1LabelSample{l1},
		})
	} else {
		m := r.metrics[index]
		if m.typeID() != gaugeType {
			panic("metrics: name in use as another type")
		}

		for _, o := range m.sampleL1s {
			if o.labelName == labelName {
				panic("metrics: label already in use")
			}
		}
		l1 = &map1LabelSample{map1Label: map1Label{
			name: name, labelName: labelName}}
		m.sampleL1s = append(m.sampleL1s, l1)
	}

	r.mutex.Unlock()
	return l1.with
}

// Must2LabelGaugeSample returns a function which assigns dedicated Sample
// instances to each label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Sample represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) the label names do not match regular expression [a-zA-Z_][a-zA-Z0-9_]*,
// (4) the label names do not appear in ascending order or
// (5) the label names are already in use.
func Must2LabelGaugeSample(name, label1Name, label2Name string) func(label1Value, label2Value string) *Sample {
	return std.Must2LabelGaugeSample(name, label1Name, label2Name)
}

// Must2LabelGaugeSample returns a function which assigns dedicated Sample
// instances to each label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Sample represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) the label names do not match regular expression [a-zA-Z_][a-zA-Z0-9_]*,
// (4) the label names do not appear in ascending order or
// (5) the label names are already in use.
func (r *Register) Must2LabelGaugeSample(name, label1Name, label2Name string) func(label1Value, label2Value string) *Sample {
	mustValidName(name)
	mustValidLabelName(label1Name)
	mustValidLabelName(label2Name)
	if label1Name > label2Name {
		panic("metrics: label name arguments aren't sorted")
	}

	r.mutex.Lock()

	var l2 *map2LabelSample
	if index, ok := r.indices[name]; !ok {
		l2 = &map2LabelSample{map2Label: map2Label{
			name: name, labelNames: [...]string{label1Name, label2Name}}}

		r.indices[name] = uint32(len(r.metrics))
		r.metrics = append(r.metrics, &metric{
			typeComment: typePrefix + name + gaugeTypeLineEnd,
			sampleL2s:   []*map2LabelSample{l2},
		})
	} else {
		m := r.metrics[index]
		if m.typeID() != gaugeType {
			panic("metrics: name in use as another type")
		}

		for _, o := range m.sampleL2s {
			if o.labelNames[0] == label1Name && o.labelNames[1] == label2Name {
				panic("metrics: labels already in use")
			}
		}
		l2 = &map2LabelSample{map2Label: map2Label{
			name: name, labelNames: [...]string{label1Name, label2Name}}}
		m.sampleL2s = append(m.sampleL2s, l2)
	}

	r.mutex.Unlock()
	return l2.with
}

// Must3LabelGaugeSample returns a function which assigns dedicated Sample
// instances to each label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Sample represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) the label names do not match regular expression [a-zA-Z_][a-zA-Z0-9_]*,
// (4) the label names do not appear in ascending order or
// (5) the label names are already in use.
func Must3LabelGaugeSample(name, label1Name, label2Name, label3Name string) func(label1Value, label2Value, label3Value string) *Sample {
	return std.Must3LabelGaugeSample(name, label1Name, label2Name, label3Name)
}

// Must3LabelGaugeSample returns a function which assigns dedicated Sample
// instances to each label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Sample represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) the label names do not match regular expression [a-zA-Z_][a-zA-Z0-9_]*,
// (4) the label names do not appear in ascending order or
// (5) the label names are already in use.
func (r *Register) Must3LabelGaugeSample(name, label1Name, label2Name, label3Name string) func(label1Value, label2Value, label3Value string) *Sample {
	mustValidName(name)
	mustValidLabelName(label1Name)
	mustValidLabelName(label2Name)
	mustValidLabelName(label3Name)
	if label1Name > label2Name || label2Name > label3Name {
		panic("metrics: label name arguments aren't sorted")
	}

	r.mutex.Lock()

	var l3 *map3LabelSample
	if index, ok := r.indices[name]; !ok {
		l3 = &map3LabelSample{map3Label: map3Label{
			name: name, labelNames: [...]string{label1Name, label2Name, label3Name}}}

		r.indices[name] = uint32(len(r.metrics))
		r.metrics = append(r.metrics, &metric{
			typeComment: typePrefix + name + gaugeTypeLineEnd,
			sampleL3s:   []*map3LabelSample{l3},
		})
	} else {
		m := r.metrics[index]
		if m.typeID() != gaugeType {
			panic("metrics: name in use as another type")
		}

		for _, o := range m.sampleL3s {
			if o.labelNames[0] == label1Name && o.labelNames[1] == label2Name && o.labelNames[2] == label3Name {
				panic("metrics: labels already in use")
			}
		}
		l3 = &map3LabelSample{map3Label: map3Label{
			name: name, labelNames: [...]string{label1Name, label2Name, label3Name}}}
		m.sampleL3s = append(m.sampleL3s, l3)
	}

	r.mutex.Unlock()
	return l3.with
}

func mustValidName(s string) {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '_' || c == ':' {
			continue
		}
		if i == 0 || c < '0' || c > '9' {
			panic("metrics: name doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*")
		}
	}
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

var helpEscapes = strings.NewReplacer("\n", `\n`, `\`, `\\`)

// Help sets the comment for the metric name. Any previous text is replaced.
// The function panics when name is not in use.
func MustHelp(name, text string) {
	std.MustHelp(name, text)
}

// Help sets the comment for the metric name. Any previous text is replaced.
// The function panics when name is not in use.
func (r *Register) MustHelp(name, text string) {
	var buf strings.Builder
	buf.Grow(len(helpPrefix) + len(name) + len(text) + 2)
	buf.WriteString(helpPrefix)
	buf.WriteString(name)
	buf.WriteByte(' ')
	helpEscapes.WriteString(&buf, text)
	buf.WriteByte('\n')

	r.mutex.Lock()
	m := r.metrics[r.indices[name]]
	if m == nil {
		panic("metrics: name not in use")
	}
	m.helpComment = buf.String()
	r.mutex.Unlock()
}
