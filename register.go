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
	realSampleID
	histogramID
)

// Help comments may have any [!] byte content, i.e., there is no illegal value.
var helpEscapes = strings.NewReplacer("\n", `\n`, `\`, `\\`)

// Metric is a named record.
type metric struct {
	typeID   uint
	comments string // TYPE + optional HELP

	counter   *Counter
	integer   *Integer
	real      *Real
	histogram *Histogram
	sample    *Sample

	labels []*labelMapping
}

func newMetric(name, help string, typeID uint) *metric {
	// compose comments
	var buf strings.Builder
	buf.Grow(len(name)*2 + len(help) + 27)
	buf.WriteString("\n# TYPE ")
	buf.WriteString(name)
	switch typeID {
	case counterID, counterSampleID:
		buf.WriteString(" counter")
	case integerID, realID, realSampleID:
		buf.WriteString(" gauge")
	case histogramID:
		buf.WriteString(" histogram")
	}
	if help != "" {
		buf.WriteString("\n# HELP ")
		buf.WriteString(name)
		buf.WriteByte(' ')
		helpEscapes.WriteString(&buf, help)
	}
	buf.WriteByte('\n')

	return &metric{typeID: typeID, comments: buf.String()}
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

// NewRegister returns an empty metric bundle. The corresponding functions
// of each Register method operate on the (hidden) default instance.
func NewRegister() *Register {
	return &Register{indices: make(map[string]uint32)}
}

func (reg *Register) mustGetOrSetMetric(name string, m *metric) *metric {
	if index, ok := reg.indices[name]; ok {
		// get it is
		got := reg.metrics[index]
		if got.typeID != m.typeID {
			panic("metrics: name already in use as another type")
		}
		return got
	}

	// set it is
	reg.indices[name] = uint32(len(reg.metrics))
	reg.metrics = append(reg.metrics, m)
	return m
}

func (reg *Register) mustGetOrCreateMetric(name string, typeID uint) *metric {
	if index, ok := reg.indices[name]; ok {
		// get it is
		got := reg.metrics[index]
		if got.typeID != typeID {
			panic("metrics: name already in use as another type")
		}
		return got
	}

	// create it is
	m := newMetric(name, "", typeID)
	reg.indices[name] = uint32(len(reg.metrics))
	reg.metrics = append(reg.metrics, m)
	return m
}

// MustCounter registers a new Counter. Registration panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Help is an optional comment text.
func MustCounter(name, help string) *Counter {
	return std.MustCounter(name, help)
}

// MustCounter registers a new Counter. Registration panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Help is an optional comment text.
func (reg *Register) MustCounter(name, help string) *Counter {
	mustValidMetricName(name)
	m := newMetric(name, help, counterID)

	reg.mutex.Lock()
	defer reg.mutex.Unlock()
	m = reg.mustGetOrSetMetric(name, m)
	if m.counter != nil {
		panic("metrics: name already in use")
	}
	m.counter = &Counter{prefix: name + " "}
	return m.counter
}

// MustInteger registers a new gauge. Registration panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Help is an optional comment text.
func MustInteger(name, help string) *Integer {
	return std.MustInteger(name, help)
}

// MustInteger registers a new gauge. Registration panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Help is an optional comment text.
func (reg *Register) MustInteger(name, help string) *Integer {
	mustValidMetricName(name)
	m := newMetric(name, help, integerID)

	reg.mutex.Lock()
	defer reg.mutex.Unlock()
	m = reg.mustGetOrSetMetric(name, m)
	if m.integer != nil {
		panic("metrics: name already in use")
	}
	m.integer = &Integer{prefix: name + " "}
	return m.integer
}

// MustReal registers a new gauge. Registration panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Help is an optional comment text.
func MustReal(name, help string) *Real {
	return std.MustReal(name, help)
}

// MustReal registers a new gauge. Registration panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Help is an optional comment text.
func (reg *Register) MustReal(name, help string) *Real {
	mustValidMetricName(name)
	m := newMetric(name, help, realID)

	reg.mutex.Lock()
	defer reg.mutex.Unlock()
	m = reg.mustGetOrSetMetric(name, m)
	if m.real != nil {
		panic("metrics: name already in use")
	}
	m.real = &Real{prefix: name + " "}
	return m.real
}

// MustHistogram registers a new Histogram. Registration panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Help is an optional comment text.
//
// Buckets are defined as upper boundary values, with positive infinity
// implied when absent. Any ∞ or not-a-number (NaN) value is ignored.
func MustHistogram(name, help string, buckets ...float64) *Histogram {
	return std.MustHistogram(name, help, buckets...)
}

// MustHistogram registers a new Histogram. Registration panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Help is an optional comment text.
//
// Buckets are defined as upper boundary values, with positive infinity
// implied when absent. Any ∞ or not-a-number (NaN) value is ignored.
func (reg *Register) MustHistogram(name, help string, buckets ...float64) *Histogram {
	mustValidMetricName(name)
	m := newMetric(name, help, histogramID)
	h := newHistogram(name, buckets)

	reg.mutex.Lock()
	defer reg.mutex.Unlock()
	m = reg.mustGetOrSetMetric(name, m)
	if m.histogram != nil {
		panic("metrics: name already in use")
	}
	m.histogram = h
	return h
}

// MustRealSample registers a new Sample. Registration panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Help is an optional comment text.
func MustRealSample(name, help string) *Sample {
	return std.MustRealSample(name, help)
}

// MustRealSample registers a new Sample. Registration panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Help is an optional comment text.
func (reg *Register) MustRealSample(name, help string) *Sample {
	mustValidMetricName(name)
	m := newMetric(name, help, realSampleID)

	reg.mutex.Lock()
	defer reg.mutex.Unlock()
	m = reg.mustGetOrSetMetric(name, m)
	if m.sample != nil {
		panic("metrics: name already in use")
	}
	m.sample = &Sample{prefix: name + " "}
	return m.sample
}

// MustCounterSample registers a new Sample. Registration panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Help is an optional comment text.
func MustCounterSample(name, help string) *Sample {
	return std.MustCounterSample(name, help)
}

// MustCounterSample registers a new Sample. Registration panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Help is an optional comment text.
func (reg *Register) MustCounterSample(name, help string) *Sample {
	mustValidMetricName(name)
	m := newMetric(name, help, counterSampleID)

	reg.mutex.Lock()
	defer reg.mutex.Unlock()
	m = reg.mustGetOrSetMetric(name, m)
	if m.sample != nil {
		panic("metrics: name already in use")
	}
	m.sample = &Sample{prefix: name + " "}
	return m.sample
}

// Must1LabelCounter returns a function which registers a dedicated Counter
// for each unique label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Counter represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) labelName does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) labelName is already in use.
func Must1LabelCounter(name, labelName string) func(labelValue string) *Counter {
	return std.Must1LabelCounter(name, labelName)
}

// Must1LabelCounter returns a function which registers a dedicated Counter
// for each unique label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Counter represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) labelName does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) labelName is already in use.
func (reg *Register) Must1LabelCounter(name, labelName string) func(labelValue string) *Counter {
	mustValidNames(name, labelName)

	reg.mutex.Lock()
	l := reg.mustGetOrCreateMetric(name, counterID).mustLabel(name, labelName, "", "")
	reg.mutex.Unlock()

	return l.counter1
}

// Must2LabelCounter returns a function which registers a dedicated Counter
// for each unique label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Counter represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) label names don't match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) label names are already in use.
func Must2LabelCounter(name, label1Name, label2Name string) func(label1Value, label2Value string) *Counter {
	return std.Must2LabelCounter(name, label1Name, label2Name)
}

// Must2LabelCounter returns a function which registers a dedicated Counter
// for each unique label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Counter represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) label names don't match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) label names are already in use.
func (reg *Register) Must2LabelCounter(name, label1Name, label2Name string) func(label1Value, label2Value string) *Counter {
	mustValidNames(name, label1Name, label2Name)

	var flip bool
	if label1Name > label2Name {
		label1Name, label2Name = label2Name, label1Name
		flip = true
	}

	reg.mutex.Lock()
	l := reg.mustGetOrCreateMetric(name, counterID).mustLabel(name, label1Name, label2Name, "")
	reg.mutex.Unlock()

	if flip {
		return l.counter21
	}
	return l.counter12
}

// Must3LabelCounter returns a function which registers a dedicated Counter
// for each unique label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Counter represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) label names don't match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) label names are already in use.
func Must3LabelCounter(name, label1Name, label2Name, label3Name string) func(label1Value, label2Value, label3Value string) *Counter {
	return std.Must3LabelCounter(name, label1Name, label2Name, label3Name)
}

// Must3LabelCounter returns a function which registers a dedicated Counter
// for each unique label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Counter represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) label names don't match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) label names are already in use.
func (reg *Register) Must3LabelCounter(name, label1Name, label2Name, label3Name string) func(label1Value, label2Value, label3Value string) *Counter {
	mustValidNames(name, label1Name, label2Name, label3Name)

	order := sort3(&label1Name, &label2Name, &label3Name)

	reg.mutex.Lock()
	l := reg.mustGetOrCreateMetric(name, counterID).mustLabel(name, label1Name, label2Name, label3Name)
	reg.mutex.Unlock()

	switch order {
	case order123:
		return l.counter123
	case order132:
		return l.counter132
	case order213:
		return l.counter213
	case order231:
		return l.counter231
	case order312:
		return l.counter312
	case order321:
		return l.counter321
	default:
		panic(order)
	}
}

// Must1LabelInteger returns a function which registers a dedicated Integer
// for each unique label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Integer represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) labelName does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) labelName is already in use.
func Must1LabelInteger(name, labelName string) func(labelValue string) *Integer {
	return std.Must1LabelInteger(name, labelName)
}

// Must1LabelInteger returns a function which registers a dedicated Integer
// for each unique label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Integer represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) labelName does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) labelName is already in use.
func (reg *Register) Must1LabelInteger(name, labelName string) func(labelValue string) *Integer {
	mustValidNames(name, labelName)

	reg.mutex.Lock()
	l := reg.mustGetOrCreateMetric(name, integerID).mustLabel(name, labelName, "", "")
	reg.mutex.Unlock()

	return l.integer1
}

// Must2LabelInteger returns a function which registers a dedicated Integer
// for each unique label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Integer represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) label names don't match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) label names are already in use.
func Must2LabelInteger(name, label1Name, label2Name string) func(label1Value, label2Value string) *Integer {
	return std.Must2LabelInteger(name, label1Name, label2Name)
}

// Must2LabelInteger returns a function which registers a dedicated Integer
// for each unique label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Integer represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) label names don't match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) label names are already in use.
func (reg *Register) Must2LabelInteger(name, label1Name, label2Name string) func(label1Value, label2Value string) *Integer {
	mustValidNames(name, label1Name, label2Name)

	var flip bool
	if label1Name > label2Name {
		label1Name, label2Name = label2Name, label1Name
		flip = true
	}

	reg.mutex.Lock()
	l := reg.mustGetOrCreateMetric(name, integerID).mustLabel(name, label1Name, label2Name, "")
	reg.mutex.Unlock()

	if flip {
		return l.integer21
	}
	return l.integer12
}

// Must3LabelInteger returns a function which registers a dedicated Integer
// for each unique label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Integer represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) label names don't match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) label names are already in use.
func Must3LabelInteger(name, label1Name, label2Name, label3Name string) func(label1Value, label2Value, label3Value string) *Integer {
	return std.Must3LabelInteger(name, label1Name, label2Name, label3Name)
}

// Must3LabelInteger returns a function which registers a dedicated Integer
// for each unique label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Integer represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) label names don't match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) label names are already in use.
func (reg *Register) Must3LabelInteger(name, label1Name, label2Name, label3Name string) func(label1Value, label2Value, label3Value string) *Integer {
	mustValidNames(name, label1Name, label2Name, label3Name)

	order := sort3(&label1Name, &label2Name, &label3Name)

	reg.mutex.Lock()
	l := reg.mustGetOrCreateMetric(name, integerID).mustLabel(name, label1Name, label2Name, label3Name)
	reg.mutex.Unlock()

	switch order {
	case order123:
		return l.integer123
	case order132:
		return l.integer132
	case order213:
		return l.integer213
	case order231:
		return l.integer231
	case order312:
		return l.integer312
	case order321:
		return l.integer321
	default:
		panic(order)
	}
}

// Must1LabelReal returns a function which registers a dedicated Real
// for each unique label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Real represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) labelName does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) labelName is already in use.
func Must1LabelReal(name, labelName string) func(labelValue string) *Real {
	return std.Must1LabelReal(name, labelName)
}

// Must1LabelReal returns a function which registers a dedicated Real
// for each unique label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Real represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) labelName does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) labelName is already in use.
func (reg *Register) Must1LabelReal(name, labelName string) func(labelValue string) *Real {
	mustValidNames(name, labelName)

	reg.mutex.Lock()
	l := reg.mustGetOrCreateMetric(name, realID).mustLabel(name, labelName, "", "")
	reg.mutex.Unlock()

	return l.real1
}

// Must2LabelReal returns a function which registers a dedicated Real
// for each unique label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Real represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) label names don't match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) label names are already in use.
func Must2LabelReal(name, label1Name, label2Name string) func(label1Value, label2Value string) *Real {
	return std.Must2LabelReal(name, label1Name, label2Name)
}

// Must2LabelReal returns a function which registers a dedicated Real
// for each unique label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Real represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) label names don't match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) label names are already in use.
func (reg *Register) Must2LabelReal(name, label1Name, label2Name string) func(label1Value, label2Value string) *Real {
	mustValidNames(name, label1Name, label2Name)

	var flip bool
	if label1Name > label2Name {
		label1Name, label2Name = label2Name, label1Name
		flip = true
	}

	reg.mutex.Lock()
	l := reg.mustGetOrCreateMetric(name, realID).mustLabel(name, label1Name, label2Name, "")
	reg.mutex.Unlock()

	if flip {
		return l.real21
	}
	return l.real12
}

// Must3LabelReal returns a function which registers a dedicated Real
// for each unique label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Real represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) label names don't match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) label names are already in use.
func Must3LabelReal(name, label1Name, label2Name, label3Name string) func(label1Value, label2Value, label3Value string) *Real {
	return std.Must3LabelReal(name, label1Name, label2Name, label3Name)
}

// Must3LabelReal returns a function which registers a dedicated Real
// for each unique label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Real represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) label names don't match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) label names are already in use.
func (reg *Register) Must3LabelReal(name, label1Name, label2Name, label3Name string) func(label1Value, label2Value, label3Value string) *Real {
	mustValidNames(name, label1Name, label2Name, label3Name)

	order := sort3(&label1Name, &label2Name, &label3Name)

	reg.mutex.Lock()
	l := reg.mustGetOrCreateMetric(name, realID).mustLabel(name, label1Name, label2Name, label3Name)
	reg.mutex.Unlock()

	switch order {
	case order123:
		return l.real123
	case order132:
		return l.real132
	case order213:
		return l.real213
	case order231:
		return l.real231
	case order312:
		return l.real312
	case order321:
		return l.real321
	default:
		panic(order)
	}
}

// Must1LabelCounterSample returns a function which registers a dedicated Sample
// for each unique label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Sample represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) labelName does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) labelName is already in use.
func Must1LabelCounterSample(name, labelName string) func(labelValue string) *Sample {
	return std.Must1LabelCounterSample(name, labelName)
}

// Must1LabelCounterSample returns a function which registers a dedicated Sample
// for each unique label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Sample represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) labelName does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) labelName is already in use.
func (reg *Register) Must1LabelCounterSample(name, labelName string) func(labelValue string) *Sample {
	mustValidNames(name, labelName)

	reg.mutex.Lock()
	l := reg.mustGetOrCreateMetric(name, counterSampleID).mustLabel(name, labelName, "", "")
	reg.mutex.Unlock()

	return l.sample1
}

// Must2LabelCounterSample returns a function which registers a dedicated Sample
// for each unique label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Sample represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) label names don't match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) label names are already in use.
func Must2LabelCounterSample(name, label1Name, label2Name string) func(label1Value, label2Value string) *Sample {
	return std.Must2LabelCounterSample(name, label1Name, label2Name)
}

// Must2LabelCounterSample returns a function which registers a dedicated Sample
// for each unique label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Sample represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) label names don't match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) label names are already in use.
func (reg *Register) Must2LabelCounterSample(name, label1Name, label2Name string) func(label1Value, label2Value string) *Sample {
	mustValidNames(name, label1Name, label2Name)

	var flip bool
	if label1Name > label2Name {
		label1Name, label2Name = label2Name, label1Name
		flip = true
	}

	reg.mutex.Lock()
	l := reg.mustGetOrCreateMetric(name, counterSampleID).mustLabel(name, label1Name, label2Name, "")
	reg.mutex.Unlock()

	if flip {
		return l.sample21
	}
	return l.sample12
}

// Must3LabelCounterSample returns a function which registers a dedicated Sample
// for each unique label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Sample represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) label names don't match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) label names are already in use.
func Must3LabelCounterSample(name, label1Name, label2Name, label3Name string) func(label1Value, label2Value, label3Value string) *Sample {
	return std.Must3LabelCounterSample(name, label1Name, label2Name, label3Name)
}

// Must3LabelCounterSample returns a function which registers a dedicated Sample
// for each unique label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Sample represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) label names don't match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) label names are already in use.
func (reg *Register) Must3LabelCounterSample(name, label1Name, label2Name, label3Name string) func(label1Value, label2Value, label3Value string) *Sample {
	mustValidNames(name, label1Name, label2Name, label3Name)

	order := sort3(&label1Name, &label2Name, &label3Name)

	reg.mutex.Lock()
	l := reg.mustGetOrCreateMetric(name, counterSampleID).mustLabel(name, label1Name, label2Name, label3Name)
	reg.mutex.Unlock()

	switch order {
	case order123:
		return l.sample123
	case order132:
		return l.sample132
	case order213:
		return l.sample213
	case order231:
		return l.sample231
	case order312:
		return l.sample312
	case order321:
		return l.sample321
	default:
		panic(order)
	}
}

// Must1LabelRealSample returns a function which registers a dedicated Sample
// for each unique label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Sample represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) labelName does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) labelName is already in use.
func Must1LabelRealSample(name, labelName string) func(labelValue string) *Sample {
	return std.Must1LabelRealSample(name, labelName)
}

// Must1LabelRealSample returns a function which registers a dedicated Sample
// for each unique label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Sample represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) labelName does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) labelName is already in use.
func (reg *Register) Must1LabelRealSample(name, labelName string) func(labelValue string) *Sample {
	mustValidNames(name, labelName)

	reg.mutex.Lock()
	l := reg.mustGetOrCreateMetric(name, realSampleID).mustLabel(name, labelName, "", "")
	reg.mutex.Unlock()

	return l.sample1
}

// Must2LabelRealSample returns a function which registers a dedicated Sample
// for each unique label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Sample represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) label names don't match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) label names are already in use.
func Must2LabelRealSample(name, label1Name, label2Name string) func(label1Value, label2Value string) *Sample {
	return std.Must2LabelRealSample(name, label1Name, label2Name)
}

// Must2LabelRealSample returns a function which registers a dedicated Sample
// for each unique label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Sample represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) label names don't match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) label names are already in use.
func (reg *Register) Must2LabelRealSample(name, label1Name, label2Name string) func(label1Value, label2Value string) *Sample {
	mustValidNames(name, label1Name, label2Name)

	var flip bool
	if label1Name > label2Name {
		label1Name, label2Name = label2Name, label1Name
		flip = true
	}

	reg.mutex.Lock()
	l := reg.mustGetOrCreateMetric(name, realSampleID).mustLabel(name, label1Name, label2Name, "")
	reg.mutex.Unlock()

	if flip {
		return l.sample21
	}
	return l.sample12
}

// Must3LabelRealSample returns a function which registers a dedicated Sample
// for each unique label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Sample represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) label names don't match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) label names are already in use.
func Must3LabelRealSample(name, label1Name, label2Name, label3Name string) func(label1Value, label2Value, label3Value string) *Sample {
	return std.Must3LabelRealSample(name, label1Name, label2Name, label3Name)
}

// Must3LabelRealSample returns a function which registers a dedicated Sample
// for each unique label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Sample represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) label names don't match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) label names are already in use.
func (reg *Register) Must3LabelRealSample(name, label1Name, label2Name, label3Name string) func(label1Value, label2Value, label3Value string) *Sample {
	mustValidNames(name, label1Name, label2Name, label3Name)

	order := sort3(&label1Name, &label2Name, &label3Name)

	reg.mutex.Lock()
	l := reg.mustGetOrCreateMetric(name, realSampleID).mustLabel(name, label1Name, label2Name, label3Name)
	reg.mutex.Unlock()

	switch order {
	case order123:
		return l.sample123
	case order132:
		return l.sample132
	case order213:
		return l.sample213
	case order231:
		return l.sample231
	case order312:
		return l.sample312
	case order321:
		return l.sample321
	default:
		panic(order)
	}
}

// Must1LabelHistogram returns a function which registers a dedicated Histogram
// for each unique label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Histogram represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) labelName does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) labelName is already in use.
//
// Buckets are defined as upper boundary values, with positive infinity
// implied when absent. Any ∞ or not-a-number (NaN) value is ignored.
func Must1LabelHistogram(name, labelName string, buckets ...float64) func(labelValue string) *Histogram {
	return std.Must1LabelHistogram(name, labelName, buckets...)
}

// Must1LabelHistogram returns a function which registers a dedicated Histogram
// for each unique label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Histogram represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) labelName does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) labelName is already in use.
//
// Buckets are defined as upper boundary values, with positive infinity
// implied when absent. Any ∞ or not-a-number (NaN) value is ignored.
func (reg *Register) Must1LabelHistogram(name, labelName string, buckets ...float64) func(labelValue string) *Histogram {
	mustValidNames(name, labelName)

	reg.mutex.Lock()
	l := reg.mustGetOrCreateMetric(name, histogramID).mustLabel(name, labelName, "", "")
	l.buckets = buckets
	reg.mutex.Unlock()

	return l.histogram1
}

// Must2LabelHistogram returns a function which registers a dedicated Histogram
// for each unique label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Histogram represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) label names don't match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) label names are already in use.
//
// Buckets are defined as upper boundary values, with positive infinity
// implied when absent. Any ∞ or not-a-number (NaN) value is ignored.
func Must2LabelHistogram(name, label1Name, label2Name string, buckets ...float64) func(label1Value, label2Value string) *Histogram {
	return std.Must2LabelHistogram(name, label1Name, label2Name, buckets...)
}

// Must2LabelHistogram returns a function which registers a dedicated Histogram
// for each unique label combination. Multiple goroutines may invoke the
// returned simultaneously. Remember that each Histogram represents a new time
// series, which can dramatically increase the amount of data stored.
//
// Must panics on any of the following:
// (1) name in use as another metric type,
// (2) name doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) label names don't match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) label names are already in use.
//
// Buckets are defined as upper boundary values, with positive infinity
// implied when absent. Any ∞ or not-a-number (NaN) value is ignored.
func (reg *Register) Must2LabelHistogram(name, label1Name, label2Name string, buckets ...float64) func(label1Value, label2Value string) *Histogram {
	mustValidNames(name, label1Name, label2Name)

	var flip bool
	if label1Name > label2Name {
		label1Name, label2Name = label2Name, label1Name
		flip = true
	}

	reg.mutex.Lock()
	l := reg.mustGetOrCreateMetric(name, histogramID).mustLabel(name, label1Name, label2Name, "")
	l.buckets = buckets
	reg.mutex.Unlock()

	if flip {
		return l.histogram21
	}
	return l.histogram12
}

func mustValidNames(metricName string, labelNames ...string) {
	mustValidMetricName(metricName)

	for _, name := range labelNames {
		mustValidLabelName(name)
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
	if s == "" {
		panic("metrics: empty label name")
	}

	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '_' {
			continue
		}
		if i == 0 || c < '0' || c > '9' {
			panic("metrics: label name doesn't match regular expression [a-zA-Z_][a-zA-Z0-9_]*")
		}
	}
}

// MustHelp sets the comment for the metric name. Any previous text is replaced.
// The function panics when name is not in use.
func MustHelp(name, text string) {
	std.MustHelp(name, text)
}

// MustHelp sets the comment for the metric name. Any previous text is replaced.
// The function panics when name is not in use.
func (reg *Register) MustHelp(name, text string) {
	reg.mutex.Lock()
	defer reg.mutex.Unlock()
	m := reg.metrics[reg.indices[name]]
	if m == nil {
		panic("metrics: name not in use")
	}

	// new-line characters are escaped in comments and label values
	i := strings.Index(m.comments, "\n# HELP ")
	if i >= 0 {
		// remove previous, but keep the new-line
		m.comments = m.comments[:i+1]
	}

	if text == "" {
		return // ommits HELP comment
	}

	var buf strings.Builder
	buf.Grow(len(m.comments) + len(name) + len(text) + 9)
	buf.WriteString(m.comments)
	buf.WriteString("# HELP ")
	buf.WriteString(name)
	buf.WriteByte(' ')
	helpEscapes.WriteString(&buf, text)
	buf.WriteByte('\n')
	m.comments = buf.String()
}

const (
	order123 = iota
	order132
	order213
	order231
	order312
	order321
)

func sort3(s1, s2, s3 *string) (order int) {
	if *s1 < *s2 {
		if *s2 < *s3 {
			order = order123
		} else if *s1 < *s3 {
			*s2, *s3 = *s3, *s2
			order = order132
		} else {
			*s1, *s2, *s3 = *s3, *s1, *s2
			order = order231
		}
	} else {
		if *s1 < *s3 {
			*s1, *s2 = *s2, *s1
			order = order213
		} else if *s2 < *s3 {
			*s1, *s2, *s3 = *s2, *s3, *s1
			order = order312
		} else {
			*s1, *s3 = *s3, *s1
			order = order321
		}
	}
	return
}
