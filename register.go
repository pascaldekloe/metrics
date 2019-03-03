package metrics

import (
	"strings"
	"sync"
)

const (
	counterID = iota
	counterSampleID
	integerID
	realID
	gaugeSampleID
	histogramID
)

// Metric is a named record.
type metric struct {
	typeID      uint
	typeComment string
	helpComment string

	counter   *Counter
	integer   *Integer
	real      *Real
	histogram *Histogram
	sample    *Sample

	labels []*labelMapping
}

func (m *metric) mustLabel(name, labelName1, labelName2, labelName3 string) *labelMapping {
	entry := &labelMapping{
		name:       name,
		labelNames: [...]string{labelName1, labelName2, labelName3},
	}

	for _, o := range m.labels {
		if o.labelNames == entry.labelNames {
			panic("metrics: labels already in use")
		}
	}

	m.labels = append(m.labels, entry)

	return entry
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
// [a-zA-Z_:][a-zA-Z0-9_:]*. Label combinations are allowed though.
func MustCounter(name string) *Counter {
	return std.MustCounter(name)
}

// MustCounter registers a new Counter. Registration panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Label combinations are allowed though.
func (r *Register) MustCounter(name string) *Counter {
	mustValidMetricName(name)

	r.mutex.Lock()
	defer r.mutex.Unlock()

	m := r.mustMetric(name, counterTypeLineEnd, counterID)
	if m.counter != nil {
		panic("metrics: name already in use")
	}

	m.counter = &Counter{prefix: name + " "}
	return m.counter
}

// MustInteger registers a new gauge. Registration panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Label combinations are allowed though.
func MustInteger(name string) *Integer {
	return std.MustInteger(name)
}

// MustInteger registers a new gauge. Registration panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Label combinations are allowed though.
func (r *Register) MustInteger(name string) *Integer {
	mustValidMetricName(name)

	r.mutex.Lock()
	defer r.mutex.Unlock()

	m := r.mustMetric(name, gaugeTypeLineEnd, integerID)
	if m.integer != nil || m.real != nil {
		panic("metrics: name already in use")
	}

	m.integer = &Integer{prefix: name + " "}
	return m.integer
}

// MustReal registers a new gauge. Registration panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Label combinations are allowed though.
func MustReal(name string) *Real {
	return std.MustReal(name)
}

// MustReal registers a new gauge. Registration panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Label combinations are allowed though.
func (r *Register) MustReal(name string) *Real {
	mustValidMetricName(name)

	r.mutex.Lock()
	defer r.mutex.Unlock()

	m := r.mustMetric(name, gaugeTypeLineEnd, realID)
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

	m := r.mustMetric(name, histogramTypeLineEnd, histogramID)
	if m.histogram != nil {
		panic("metrics: name already in use")
	}

	m.histogram = h
	return h
}

// MustGaugeSample registers a new Sample. Registration panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Label combinations are allowed though.
func MustGaugeSample(name string) *Sample {
	return MustGaugeSample(name)
}

// MustGaugeSample registers a new Sample. Registration panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Label combinations are allowed though.
func (r *Register) MustGaugeSample(name string) *Sample {
	mustValidMetricName(name)

	r.mutex.Lock()
	defer r.mutex.Unlock()

	m := r.mustMetric(name, gaugeTypeLineEnd, gaugeSampleID)
	if m.sample != nil {
		panic("metrics: name already in use")
	}

	m.sample = &Sample{prefix: name + " "}
	return m.sample
}

// MustCounterSample registers a new Sample. Registration panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Label combinations are allowed though.
func MustCounterSample(name string) *Sample {
	return std.MustCounterSample(name)
}

// MustCounterSample registers a new Sample. Registration panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Label combinations are allowed though.
func (r *Register) MustCounterSample(name string) *Sample {
	mustValidMetricName(name)

	r.mutex.Lock()
	defer r.mutex.Unlock()

	m := r.mustMetric(name, counterTypeLineEnd, counterSampleID)
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

	r.mutex.Lock()
	l := r.mustMetric(name, counterTypeLineEnd, counterID).mustLabel(name, labelName, "", "")
	r.mutex.Unlock()

	return l.counter1
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

	r.mutex.Lock()
	l := r.mustMetric(name, counterTypeLineEnd, counterID).mustLabel(name, label1Name, label2Name, "")
	r.mutex.Unlock()

	return l.counter2
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

	r.mutex.Lock()
	l := r.mustMetric(name, counterTypeLineEnd, counterID).mustLabel(name, label1Name, label2Name, label3Name)
	r.mutex.Unlock()

	return l.counter3
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

	r.mutex.Lock()
	l := r.mustMetric(name, gaugeTypeLineEnd, integerID).mustLabel(name, labelName, "", "")
	r.mutex.Unlock()

	return l.integer1
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

	r.mutex.Lock()
	l := r.mustMetric(name, gaugeTypeLineEnd, integerID).mustLabel(name, label1Name, label2Name, "")
	r.mutex.Unlock()

	return l.integer2
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

	r.mutex.Lock()
	l := r.mustMetric(name, gaugeTypeLineEnd, integerID).mustLabel(name, label1Name, label2Name, label3Name)
	r.mutex.Unlock()

	return l.integer3
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

	r.mutex.Lock()
	l := r.mustMetric(name, gaugeTypeLineEnd, realID).mustLabel(name, labelName, "", "")
	r.mutex.Unlock()

	return l.real1
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

	r.mutex.Lock()
	l := r.mustMetric(name, gaugeTypeLineEnd, realID).mustLabel(name, label1Name, label2Name, "")
	r.mutex.Unlock()

	return l.real2
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

	r.mutex.Lock()
	l := r.mustMetric(name, gaugeTypeLineEnd, realID).mustLabel(name, label1Name, label2Name, label3Name)
	r.mutex.Unlock()

	return l.real3
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

	r.mutex.Lock()
	l := r.mustMetric(name, histogramTypeLineEnd, histogramID).mustLabel(name, labelName, "", "")
	l.buckets = buckets
	r.mutex.Unlock()

	return l.histogram1
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

	r.mutex.Lock()
	l := r.mustMetric(name, histogramTypeLineEnd, histogramID).mustLabel(name, label1Name, label2Name, "")
	l.buckets = buckets
	r.mutex.Unlock()

	return l.histogram2
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

	r.mutex.Lock()
	l := r.mustMetric(name, counterTypeLineEnd, counterSampleID).mustLabel(name, labelName, "", "")
	r.mutex.Unlock()

	return l.sample1
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

	r.mutex.Lock()
	l := r.mustMetric(name, counterTypeLineEnd, counterSampleID).mustLabel(name, label1Name, label2Name, "")
	r.mutex.Unlock()

	return l.sample2
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

	r.mutex.Lock()
	l := r.mustMetric(name, counterTypeLineEnd, counterSampleID).mustLabel(name, label1Name, label2Name, label3Name)
	r.mutex.Unlock()

	return l.sample3
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

	r.mutex.Lock()
	l := r.mustMetric(name, gaugeTypeLineEnd, gaugeSampleID).mustLabel(name, labelName, "", "")
	r.mutex.Unlock()

	return l.sample1
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

	r.mutex.Lock()
	l := r.mustMetric(name, gaugeTypeLineEnd, gaugeSampleID).mustLabel(name, label1Name, label2Name, "")
	r.mutex.Unlock()

	return l.sample2
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

	r.mutex.Lock()
	l := r.mustMetric(name, gaugeTypeLineEnd, gaugeSampleID).mustLabel(name, label1Name, label2Name, label3Name)
	r.mutex.Unlock()

	return l.sample3
}

func (r *Register) mustMetric(name, typeLineEnd string, typeID uint) *metric {
	if index, ok := r.indices[name]; ok {
		m := r.metrics[index]
		if m.typeID != typeID {
			panic("metrics: name in use as another type")
		}

		return m
	}

	m := &metric{typeComment: typePrefix + name + typeLineEnd, typeID: typeID}

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
