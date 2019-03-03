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
	mustValidMetricName(name)

	r.mutex.Lock()
	defer r.mutex.Unlock()

	m := r.mustMetric(name, counterTypeLineEnd)
	if m.counter != nil {
		panic("metrics: name already in use")
	}

	m.counter = &Counter{prefix: name + " "}
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
	mustValidMetricName(name)

	r.mutex.Lock()
	defer r.mutex.Unlock()

	m := r.mustMetric(name, gaugeTypeLineEnd)
	if m.integer != nil || m.real != nil {
		panic("metrics: name already in use")
	}

	m.integer = &Integer{prefix: name + " "}
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
	mustValidMetricName(name)

	r.mutex.Lock()
	defer r.mutex.Unlock()

	m := r.mustMetric(name, gaugeTypeLineEnd)
	if m.integer != nil || m.real != nil {
		panic("metrics: name already in use")
	}

	m.real = &Real{prefix: name + " "}
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
	mustValidMetricName(name)

	h := newHistogram(name, buckets)
	h.prefix(name)

	r.mutex.Lock()
	defer r.mutex.Unlock()

	m := r.mustMetric(name, histogramTypeLineEnd)
	if m.histogram != nil {
		panic("metrics: name already in use")
	}

	m.histogram = h
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
	mustValidMetricName(name)

	r.mutex.Lock()
	defer r.mutex.Unlock()

	m := r.mustMetric(name, gaugeTypeLineEnd)
	if m.sample != nil {
		panic("metrics: name already in use")
	}

	m.sample = &Sample{prefix: name + " "}
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
	mustValidMetricName(name)

	r.mutex.Lock()
	defer r.mutex.Unlock()

	m := r.mustMetric(name, counterTypeLineEnd)
	if m.sample != nil {
		panic("metrics: name already in use")
	}

	m.sample = &Sample{prefix: name + " "}

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
	mustValidNames(name, labelName)

	m1 := new(map1LabelCounter)
	m1.name = name
	m1.labelName = labelName

	r.mutex.Lock()
	defer r.mutex.Unlock()

	m := r.mustMetric(name, counterTypeLineEnd)
	for _, o := range m.counterL1s {
		if o.labelName == m1.labelName {
			panic("metrics: labels already in use")
		}
	}
	m.counterL1s = append(m.counterL1s, m1)

	return m1.with
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
	mustValidNames(name, label1Name, label2Name)

	m2 := new(map2LabelCounter)
	m2.name = name
	m2.labelNames = [...]string{label1Name, label2Name}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	m := r.mustMetric(name, counterTypeLineEnd)
	for _, o := range m.counterL2s {
		if o.labelNames == m2.labelNames {
			panic("metrics: labels already in use")
		}
	}
	m.counterL2s = append(m.counterL2s, m2)

	return m2.with
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
	mustValidNames(name, label1Name, label2Name, label3Name)

	m3 := new(map3LabelCounter)
	m3.name = name
	m3.labelNames = [...]string{label1Name, label2Name, label3Name}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	m := r.mustMetric(name, counterTypeLineEnd)
	for _, o := range m.counterL3s {
		if o.labelNames == m3.labelNames {
			panic("metrics: labels already in use")
		}
	}
	m.counterL3s = append(m.counterL3s, m3)

	return m3.with
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
	mustValidNames(name, labelName)

	m1 := new(map1LabelInteger)
	m1.name = name
	m1.labelName = labelName

	r.mutex.Lock()
	defer r.mutex.Unlock()

	m := r.mustMetric(name, gaugeTypeLineEnd)
	for _, o := range m.integerL1s {
		if o.labelName == m1.labelName {
			panic("metrics: labels already in use")
		}
	}
	m.integerL1s = append(m.integerL1s, m1)

	return m1.with
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
	mustValidNames(name, label1Name, label2Name)

	m2 := new(map2LabelInteger)
	m2.name = name
	m2.labelNames = [...]string{label1Name, label2Name}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	m := r.mustMetric(name, gaugeTypeLineEnd)
	for _, o := range m.integerL2s {
		if o.labelNames == m2.labelNames {
			panic("metrics: labels already in use")
		}
	}
	m.integerL2s = append(m.integerL2s, m2)

	return m2.with
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
	mustValidNames(name, label1Name, label2Name, label3Name)

	m3 := new(map3LabelInteger)
	m3.name = name
	m3.labelNames = [...]string{label1Name, label2Name, label3Name}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	m := r.mustMetric(name, gaugeTypeLineEnd)
	for _, o := range m.integerL3s {
		if o.labelNames == m3.labelNames {
			panic("metrics: labels already in use")
		}
	}
	m.integerL3s = append(m.integerL3s, m3)

	return m3.with
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
	mustValidNames(name, labelName)

	m1 := new(map1LabelReal)
	m1.name = name
	m1.labelName = labelName

	r.mutex.Lock()
	defer r.mutex.Unlock()

	m := r.mustMetric(name, gaugeTypeLineEnd)
	for _, o := range m.realL1s {
		if o.labelName == m1.labelName {
			panic("metrics: labels already in use")
		}
	}
	m.realL1s = append(m.realL1s, m1)

	return m1.with
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
	mustValidNames(name, label1Name, label2Name)

	m2 := new(map2LabelReal)
	m2.name = name
	m2.labelNames = [...]string{label1Name, label2Name}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	m := r.mustMetric(name, gaugeTypeLineEnd)
	for _, o := range m.realL2s {
		if o.labelNames == m2.labelNames {
			panic("metrics: labels already in use")
		}
	}
	m.realL2s = append(m.realL2s, m2)

	return m2.with
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
	mustValidNames(name, label1Name, label2Name, label3Name)

	m3 := new(map3LabelReal)
	m3.name = name
	m3.labelNames = [...]string{label1Name, label2Name, label3Name}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	m := r.mustMetric(name, gaugeTypeLineEnd)
	for _, o := range m.realL3s {
		if o.labelNames == m3.labelNames {
			panic("metrics: labels already in use")
		}
	}
	m.realL3s = append(m.realL3s, m3)

	return m3.with
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
	mustValidNames(name, labelName)

	m1 := &map1LabelHistogram{buckets: buckets}
	m1.name = name
	m1.labelName = labelName

	r.mutex.Lock()
	defer r.mutex.Unlock()

	m := r.mustMetric(name, histogramTypeLineEnd)
	for _, o := range m.histogramL1s {
		if o.labelName == m1.labelName {
			panic("metrics: labels already in use")
		}
	}
	m.histogramL1s = append(m.histogramL1s, m1)

	return m1.with
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
	mustValidNames(name, label1Name, label2Name)

	m2 := &map2LabelHistogram{buckets: buckets}
	m2.name = name
	m2.labelNames = [...]string{label1Name, label2Name}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	m := r.mustMetric(name, histogramTypeLineEnd)
	for _, o := range m.histogramL2s {
		if o.labelNames == m2.labelNames {
			panic("metrics: labels already in use")
		}
	}
	m.histogramL2s = append(m.histogramL2s, m2)

	return m2.with
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
	mustValidNames(name, labelName)

	m1 := new(map1LabelSample)
	m1.name = name
	m1.labelName = labelName

	r.mutex.Lock()
	defer r.mutex.Unlock()

	m := r.mustMetric(name, counterTypeLineEnd)
	for _, o := range m.sampleL1s {
		if o.labelName == m1.labelName {
			panic("metrics: labels already in use")
		}
	}
	m.sampleL1s = append(m.sampleL1s, m1)

	return m1.with
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
	mustValidNames(name, label1Name, label2Name)

	m2 := new(map2LabelSample)
	m2.name = name
	m2.labelNames = [...]string{label1Name, label2Name}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	m := r.mustMetric(name, counterTypeLineEnd)
	for _, o := range m.sampleL2s {
		if o.labelNames == m2.labelNames {
			panic("metrics: labels already in use")
		}
	}
	m.sampleL2s = append(m.sampleL2s, m2)

	return m2.with
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
	mustValidNames(name, label1Name, label2Name, label3Name)

	m3 := new(map3LabelSample)
	m3.name = name
	m3.labelNames = [...]string{label1Name, label2Name, label3Name}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	m := r.mustMetric(name, counterTypeLineEnd)
	for _, o := range m.sampleL3s {
		if o.labelNames == m3.labelNames {
			panic("metrics: labels already in use")
		}
	}
	m.sampleL3s = append(m.sampleL3s, m3)

	return m3.with
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
	mustValidNames(name, labelName)

	m1 := new(map1LabelSample)
	m1.name = name
	m1.labelName = labelName

	r.mutex.Lock()
	defer r.mutex.Unlock()

	m := r.mustMetric(name, gaugeTypeLineEnd)
	for _, o := range m.sampleL1s {
		if o.labelName == m1.labelName {
			panic("metrics: labels already in use")
		}
	}
	m.sampleL1s = append(m.sampleL1s, m1)

	return m1.with
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
	mustValidNames(name, label1Name, label2Name)

	m2 := new(map2LabelSample)
	m2.name = name
	m2.labelNames = [...]string{label1Name, label2Name}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	m := r.mustMetric(name, gaugeTypeLineEnd)
	for _, o := range m.sampleL2s {
		if o.labelNames == m2.labelNames {
			panic("metrics: labels already in use")
		}
	}
	m.sampleL2s = append(m.sampleL2s, m2)

	return m2.with
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
	mustValidNames(name, label1Name, label2Name, label3Name)

	m3 := new(map3LabelSample)
	m3.name = name
	m3.labelNames = [...]string{label1Name, label2Name, label3Name}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	m := r.mustMetric(name, gaugeTypeLineEnd)
	for _, o := range m.sampleL3s {
		if o.labelNames == m3.labelNames {
			panic("metrics: labels already in use")
		}
	}
	m.sampleL3s = append(m.sampleL3s, m3)

	return m3.with
}

func (r *Register) mustMetric(name, typeLineEnd string) *metric {
	if index, ok := r.indices[name]; ok {
		m := r.metrics[index]
		if m.typeID() != typeLineEnd[len(typeLineEnd)-2] {
			panic("metrics: name in use as another type")
		}

		return m
	}

	m := &metric{typeComment: typePrefix + name + typeLineEnd}

	r.indices[name] = uint32(len(r.metrics))
	r.metrics = append(r.metrics, m)

	return m
}

func mustValidNames(metricName string, labelNames ...string) {
	mustValidMetricName(metricName)

	var last string
	for _, name := range labelNames {
		if name <= last {
			if name == "" {
				panic("metrics: empty label name")
			}
			panic("metrics: label name arguments aren't sorted")
		}
		mustValidLabelName(name)
		last = name
	}
}

func mustValidMetricName(s string) {
	if s == "" {
		panic("metrics: empty name")
	}

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
