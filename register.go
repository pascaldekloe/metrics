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
	gauge     *Gauge
	histogram *Histogram
	sample    *Sample

	counterL1s   []*Map1LabelCounter
	counterL2s   []*Map2LabelCounter
	counterL3s   []*Map3LabelCounter
	gaugeL1s     []*Map1LabelGauge
	gaugeL2s     []*Map2LabelGauge
	gaugeL3s     []*Map3LabelGauge
	histogramL1s []*Map1LabelHistogram
	histogramL2s []*Map2LabelHistogram
	sampleL1s    []*Map1LabelSample
	sampleL2s    []*Map2LabelSample
	sampleL3s    []*Map3LabelSample
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

// MustNewCounter registers a new Counter. The function panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Combinations of Counter, Sample and the various
// label options are allowed though. The Sample is ignored once a Counter is
// registered under the same name. This fallback allows warm starts.
func MustNewCounter(name string) *Counter {
	return std.MustNewCounter(name)
}

// MustNewCounter registers a new Counter. The function panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Combinations of Counter, Sample and the various
// label options are allowed though. The Sample is ignored once a Counter is
// registered under the same name. This fallback allows warm starts.
func (r *Register) MustNewCounter(name string) *Counter {
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

	c := &Counter{prefix: name + " "}
	m.counter = c

	r.mutex.Unlock()

	return c
}

// MustNewGauge registers a new Gauge. The function panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Combinations of Gauge, Sample and the various
// label options are allowed though. The Sample is ignored once a Gauge is
// registered under the same name. This fallback allows warm starts.
func MustNewGauge(name string) *Gauge {
	return std.MustNewGauge(name)
}

// MustNewGauge registers a new Gauge. The function panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Combinations of Gauge, Sample and the various
// label options are allowed though. The Sample is ignored once a Gauge is
// registered under the same name. This fallback allows warm starts.
func (r *Register) MustNewGauge(name string) *Gauge {
	mustValidName(name)

	r.mutex.Lock()

	var m *metric
	if index, ok := r.indices[name]; ok {
		m = r.metrics[index]
		if m.typeID() != gaugeType || m.gauge != nil {
			panic("metrics: name already in use")
		}
	} else {
		r.indices[name] = uint32(len(r.metrics))
		m = &metric{typeComment: typePrefix + name + gaugeTypeLineEnd}
		r.metrics = append(r.metrics, m)
	}

	g := &Gauge{prefix: name + " "}
	m.gauge = g

	r.mutex.Unlock()

	return g
}

// MustNewHistogram registers a new Histogram. Buckets define the upper
// boundaries, preferably in ascending order. Special cases not-a-number
// and both infinities are ignored.
// The function panics when name was registered before, or when name
// doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*.
func MustNewHistogram(name string, buckets ...float64) *Histogram {
	return std.MustNewHistogram(name, buckets...)
}

// MustNewHistogram registers a new Histogram. Buckets define the upper
// boundaries, preferably in ascending order. Special cases not-a-number
// and both infinities are ignored.
// The function panics when name was registered before, or when name
// doesn't match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*.
func (r *Register) MustNewHistogram(name string, buckets ...float64) *Histogram {
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

// MustNewGaugeSample registers a new Sample. The function panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Combinations of Gauge, Sample and the various
// label options are allowed though. The Sample is ignored once a Gauge is
// registered under the same name. This fallback allows warm starts.
func MustNewGaugeSample(name string) *Sample {
	return MustNewGaugeSample(name)
}

// MustNewGaugeSample registers a new Sample. The function panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Combinations of Gauge, Sample and the various
// label options are allowed though. The Sample is ignored once a Gauge is
// registered under the same name. This fallback allows warm starts.
func (r *Register) MustNewGaugeSample(name string) *Sample {
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

	s := &Sample{prefix: name + " "}
	m.sample = s

	r.mutex.Unlock()

	return s
}

// MustNewCounterSample registers a new Sample. The function panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Combinations of Counter, Sample and the various
// label options are allowed though. The Sample is ignored once a Counter is
// registered under the same name. This fallback allows warm starts.
func MustNewCounterSample(name string) *Sample {
	return std.MustNewCounterSample(name)
}

// MustNewCounterSample registers a new Sample. The function panics when name
// was registered before, or when name doesn't match regular expression
// [a-zA-Z_:][a-zA-Z0-9_:]*. Combinations of Counter, Sample and the various
// label options are allowed though. The Sample is ignored once a Counter is
// registered under the same name. This fallback allows warm starts.
func (r *Register) MustNewCounterSample(name string) *Sample {
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

	s := &Sample{prefix: name + " "}
	m.sample = s

	r.mutex.Unlock()

	return s
}

// MustNew1LabelCounter returns a composition with one fixed label.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) the label is already in use.
func MustNew1LabelCounter(name, labelName string) *Map1LabelCounter {
	return std.MustNew1LabelCounter(name, labelName)
}

// MustNew1LabelCounter returns a composition with one fixed label.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) the label is already in use.
func (r *Register) MustNew1LabelCounter(name, labelName string) *Map1LabelCounter {
	mustValidName(name)
	mustValidLabelName(labelName)

	r.mutex.Lock()

	var l1 *Map1LabelCounter
	if index, ok := r.indices[name]; !ok {
		l1 = &Map1LabelCounter{map1Label: map1Label{
			name: name, labelName: labelName}}

		r.indices[name] = uint32(len(r.metrics))
		r.metrics = append(r.metrics, &metric{
			typeComment: typePrefix + name + counterTypeLineEnd,
			counterL1s:  []*Map1LabelCounter{l1},
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
		l1 = &Map1LabelCounter{map1Label: map1Label{
			name: name, labelName: labelName}}
		m.counterL1s = append(m.counterL1s, l1)
	}

	r.mutex.Unlock()
	return l1
}

// MustNew2LabelCounter returns a composition with two fixed labels.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]*,
// (4) the label names do not appear in ascending order or
// (5) the labels are already in use.
func MustNew2LabelCounter(name, label1Name, label2Name string) *Map2LabelCounter {
	return std.MustNew2LabelCounter(name, label1Name, label2Name)
}

// MustNew2LabelCounter returns a composition with two fixed labels.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]*,
// (4) the label names do not appear in ascending order or
// (5) the labels are already in use.
func (r *Register) MustNew2LabelCounter(name, label1Name, label2Name string) *Map2LabelCounter {
	mustValidName(name)
	mustValidLabelName(label1Name)
	mustValidLabelName(label2Name)
	if label1Name > label2Name {
		panic("metrics: label name arguments aren't sorted")
	}

	r.mutex.Lock()

	var l2 *Map2LabelCounter
	if index, ok := r.indices[name]; !ok {
		l2 = &Map2LabelCounter{map2Label: map2Label{
			name: name, labelNames: [...]string{label1Name, label2Name}}}

		r.indices[name] = uint32(len(r.metrics))
		r.metrics = append(r.metrics, &metric{
			typeComment: typePrefix + name + counterTypeLineEnd,
			counterL2s:  []*Map2LabelCounter{l2},
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
		l2 = &Map2LabelCounter{map2Label: map2Label{
			name: name, labelNames: [...]string{label1Name, label2Name}}}
		m.counterL2s = append(m.counterL2s, l2)
	}

	r.mutex.Unlock()
	return l2
}

// MustNew3LabelCounter returns a composition with three fixed labels.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]*,
// (4) the label names do not appear in ascending order or
// (5) the labels are already in use.
func MustNew3LabelCounter(name, label1Name, label2Name, label3Name string) *Map3LabelCounter {
	return std.MustNew3LabelCounter(name, label1Name, label2Name, label3Name)
}

// MustNew3LabelCounter returns a composition with three fixed labels.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]*,
// (4) the label names do not appear in ascending order or
// (5) the labels are already in use.
func (r *Register) MustNew3LabelCounter(name, label1Name, label2Name, label3Name string) *Map3LabelCounter {
	mustValidName(name)
	mustValidLabelName(label1Name)
	mustValidLabelName(label2Name)
	mustValidLabelName(label3Name)
	if label1Name > label2Name || label2Name > label3Name {
		panic("metrics: label name arguments aren't sorted")
	}

	r.mutex.Lock()

	var l3 *Map3LabelCounter
	if index, ok := r.indices[name]; !ok {
		l3 = &Map3LabelCounter{map3Label: map3Label{
			name: name, labelNames: [...]string{label1Name, label2Name, label3Name}}}

		r.indices[name] = uint32(len(r.metrics))
		r.metrics = append(r.metrics, &metric{
			typeComment: typePrefix + name + counterTypeLineEnd,
			counterL3s:  []*Map3LabelCounter{l3},
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
		l3 = &Map3LabelCounter{map3Label: map3Label{
			name: name, labelNames: [...]string{label1Name, label2Name, label3Name}}}
		m.counterL3s = append(m.counterL3s, l3)
	}

	r.mutex.Unlock()
	return l3
}

// MustNew1LabelGauge returns a composition with one fixed label.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) the label is already in use.
func MustNew1LabelGauge(name, labelName string) *Map1LabelGauge {
	return std.MustNew1LabelGauge(name, labelName)
}

// MustNew1LabelGauge returns a composition with one fixed label.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) the label is already in use.
func (r *Register) MustNew1LabelGauge(name, labelName string) *Map1LabelGauge {
	mustValidName(name)
	mustValidLabelName(labelName)

	r.mutex.Lock()

	var l1 *Map1LabelGauge
	if index, ok := r.indices[name]; !ok {
		l1 = &Map1LabelGauge{map1Label: map1Label{
			name: name, labelName: labelName}}

		r.indices[name] = uint32(len(r.metrics))
		r.metrics = append(r.metrics, &metric{
			typeComment: typePrefix + name + gaugeTypeLineEnd,
			gaugeL1s:    []*Map1LabelGauge{l1},
		})
	} else {
		m := r.metrics[index]
		if m.typeID() != gaugeType {
			panic("metrics: name in use as another type")
		}

		for _, o := range m.gaugeL1s {
			if o.labelName == labelName {
				panic("metrics: label already in use")
			}
		}
		l1 = &Map1LabelGauge{map1Label: map1Label{
			name: name, labelName: labelName}}
		m.gaugeL1s = append(m.gaugeL1s, l1)
	}

	r.mutex.Unlock()
	return l1
}

// MustNew2LabelGauge returns a composition with two fixed labels.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]*,
// (4) the label names do not appear in ascending order or
// (5) the labels are already in use.
func MustNew2LabelGauge(name, label1Name, label2Name string) *Map2LabelGauge {
	return std.MustNew2LabelGauge(name, label1Name, label2Name)
}

// MustNew2LabelGauge returns a composition with two fixed labels.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]*,
// (4) the label names do not appear in ascending order or
// (5) the labels are already in use.
func (r *Register) MustNew2LabelGauge(name, label1Name, label2Name string) *Map2LabelGauge {
	mustValidName(name)
	mustValidLabelName(label1Name)
	mustValidLabelName(label2Name)
	if label1Name > label2Name {
		panic("metrics: label name arguments aren't sorted")
	}

	r.mutex.Lock()

	var l2 *Map2LabelGauge
	if index, ok := r.indices[name]; !ok {
		l2 = &Map2LabelGauge{map2Label: map2Label{
			name: name, labelNames: [...]string{label1Name, label2Name}}}

		r.indices[name] = uint32(len(r.metrics))
		r.metrics = append(r.metrics, &metric{
			typeComment: typePrefix + name + gaugeTypeLineEnd,
			gaugeL2s:    []*Map2LabelGauge{l2},
		})
	} else {
		m := r.metrics[index]
		if m.typeID() != gaugeType {
			panic("metrics: name in use as another type")
		}

		for _, o := range m.gaugeL2s {
			if o.labelNames[0] == label1Name && o.labelNames[1] == label2Name {
				panic("metrics: labels already in use")
			}
		}
		l2 = &Map2LabelGauge{map2Label: map2Label{
			name: name, labelNames: [...]string{label1Name, label2Name}}}
		m.gaugeL2s = append(m.gaugeL2s, l2)
	}

	r.mutex.Unlock()
	return l2
}

// MustNew3LabelGauge returns a composition with three fixed labels.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]*,
// (4) the label names do not appear in ascending order or
// (5) the labels are already in use.
func MustNew3LabelGauge(name, label1Name, label2Name, label3Name string) *Map3LabelGauge {
	return std.MustNew3LabelGauge(name, label1Name, label2Name, label3Name)
}

// MustNew3LabelGauge returns a composition with three fixed labels.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]*,
// (4) the label names do not appear in ascending order or
// (5) the labels are already in use.
func (r *Register) MustNew3LabelGauge(name, label1Name, label2Name, label3Name string) *Map3LabelGauge {
	mustValidName(name)
	mustValidLabelName(label1Name)
	mustValidLabelName(label2Name)
	mustValidLabelName(label3Name)
	if label1Name > label2Name || label2Name > label3Name {
		panic("metrics: label name arguments aren't sorted")
	}

	r.mutex.Lock()

	var l3 *Map3LabelGauge
	if index, ok := r.indices[name]; !ok {
		l3 = &Map3LabelGauge{map3Label: map3Label{
			name: name, labelNames: [...]string{label1Name, label2Name, label3Name}}}

		r.indices[name] = uint32(len(r.metrics))
		r.metrics = append(r.metrics, &metric{
			typeComment: typePrefix + name + gaugeTypeLineEnd,
			gaugeL3s:    []*Map3LabelGauge{l3},
		})
	} else {
		m := r.metrics[index]
		if m.typeID() != gaugeType {
			panic("metrics: name in use as another type")
		}

		for _, o := range m.gaugeL3s {
			if o.labelNames[0] == label1Name && o.labelNames[1] == label2Name && o.labelNames[2] == label3Name {
				panic("metrics: labels already in use")
			}
		}
		l3 = &Map3LabelGauge{map3Label: map3Label{
			name: name, labelNames: [...]string{label1Name, label2Name, label3Name}}}
		m.gaugeL3s = append(m.gaugeL3s, l3)
	}

	r.mutex.Unlock()
	return l3
}

// MustNew1LabelHistogram returns a composition with one fixed label.
// Buckets define the upper boundaries, preferably in ascending order.
// Special cases not-a-number and both infinities are ignored.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) the label is already in use.
func MustNew1LabelHistogram(name, labelName string, buckets ...float64) *Map1LabelHistogram {
	return std.MustNew1LabelHistogram(name, labelName, buckets...)
}

// MustNew1LabelHistogram returns a composition with one fixed label.
// Buckets define the upper boundaries, preferably in ascending order.
// Special cases not-a-number and both infinities are ignored.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) the label is already in use.
func (r *Register) MustNew1LabelHistogram(name, labelName string, buckets ...float64) *Map1LabelHistogram {
	mustValidName(name)
	mustValidLabelName(labelName)

	r.mutex.Lock()

	var l1 *Map1LabelHistogram
	if index, ok := r.indices[name]; !ok {
		l1 = &Map1LabelHistogram{buckets: buckets, map1Label: map1Label{
			name: name, labelName: labelName}}

		r.indices[name] = uint32(len(r.metrics))
		r.metrics = append(r.metrics, &metric{
			typeComment:  typePrefix + name + histogramTypeLineEnd,
			histogramL1s: []*Map1LabelHistogram{l1},
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
		l1 = &Map1LabelHistogram{buckets: buckets, map1Label: map1Label{
			name: name, labelName: labelName}}
		m.histogramL1s = append(m.histogramL1s, l1)
	}

	r.mutex.Unlock()
	return l1
}

// MustNew2LabelHistogram returns a composition with two fixed labels.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]*,
// (4) the label names do not appear in ascending order or
// (5) the labels are already in use.
func MustNew2LabelHistogram(name, label1Name, label2Name string, buckets ...float64) *Map2LabelHistogram {
	return std.MustNew2LabelHistogram(name, label1Name, label2Name, buckets...)
}

// MustNew2LabelHistogram returns a composition with two fixed labels.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]*,
// (4) the label names do not appear in ascending order or
// (5) the labels are already in use.
func (r *Register) MustNew2LabelHistogram(name, label1Name, label2Name string, buckets ...float64) *Map2LabelHistogram {
	mustValidName(name)
	mustValidLabelName(label1Name)
	mustValidLabelName(label2Name)
	if label1Name > label2Name {
		panic("metrics: label name arguments aren't sorted")
	}

	r.mutex.Lock()

	var l2 *Map2LabelHistogram
	if index, ok := r.indices[name]; !ok {
		l2 = &Map2LabelHistogram{buckets: buckets, map2Label: map2Label{
			name: name, labelNames: [...]string{label1Name, label2Name}}}

		r.indices[name] = uint32(len(r.metrics))
		r.metrics = append(r.metrics, &metric{
			typeComment:  typePrefix + name + histogramTypeLineEnd,
			histogramL2s: []*Map2LabelHistogram{l2},
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
		l2 = &Map2LabelHistogram{buckets: buckets, map2Label: map2Label{
			name: name, labelNames: [...]string{label1Name, label2Name}}}
		m.histogramL2s = append(m.histogramL2s, l2)
	}

	r.mutex.Unlock()
	return l2
}

// MustNew1LabelCounterSample returns a composition with one fixed label.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) the label is already in use.
func MustNew1LabelCounterSample(name, labelName string) *Map1LabelSample {
	return std.MustNew1LabelCounterSample(name, labelName)
}

// MustNew1LabelCounterSample returns a composition with one fixed label.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) the label is already in use.
func (r *Register) MustNew1LabelCounterSample(name, labelName string) *Map1LabelSample {
	mustValidName(name)
	mustValidLabelName(labelName)

	r.mutex.Lock()

	var l1 *Map1LabelSample
	if index, ok := r.indices[name]; !ok {
		l1 = &Map1LabelSample{map1Label: map1Label{
			name: name, labelName: labelName}}

		r.indices[name] = uint32(len(r.metrics))
		r.metrics = append(r.metrics, &metric{
			typeComment: typePrefix + name + counterTypeLineEnd,
			sampleL1s:   []*Map1LabelSample{l1},
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
		l1 = &Map1LabelSample{map1Label: map1Label{
			name: name, labelName: labelName}}
		m.sampleL1s = append(m.sampleL1s, l1)
	}

	r.mutex.Unlock()
	return l1
}

// MustNew2LabelCounterSample returns a composition with two fixed labels.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]*,
// (4) the label names do not appear in ascending order or
// (5) the labels are already in use.
func MustNew2LabelCounterSample(name, label1Name, label2Name string) *Map2LabelSample {
	return std.MustNew2LabelCounterSample(name, label1Name, label2Name)
}

// MustNew2LabelCounterSample returns a composition with two fixed labels.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]*,
// (4) the label names do not appear in ascending order or
// (5) the labels are already in use.
func (r *Register) MustNew2LabelCounterSample(name, label1Name, label2Name string) *Map2LabelSample {
	mustValidName(name)
	mustValidLabelName(label1Name)
	mustValidLabelName(label2Name)
	if label1Name > label2Name {
		panic("metrics: label name arguments aren't sorted")
	}

	r.mutex.Lock()

	var l2 *Map2LabelSample
	if index, ok := r.indices[name]; !ok {
		l2 = &Map2LabelSample{map2Label: map2Label{
			name: name, labelNames: [...]string{label1Name, label2Name}}}

		r.indices[name] = uint32(len(r.metrics))
		r.metrics = append(r.metrics, &metric{
			typeComment: typePrefix + name + counterTypeLineEnd,
			sampleL2s:   []*Map2LabelSample{l2},
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
		l2 = &Map2LabelSample{map2Label: map2Label{
			name: name, labelNames: [...]string{label1Name, label2Name}}}
		m.sampleL2s = append(m.sampleL2s, l2)
	}

	r.mutex.Unlock()
	return l2
}

// MustNew3LabelCounterSample returns a composition with three fixed labels.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]*,
// (4) the label names do not appear in ascending order or
// (5) the labels are already in use.
func MustNew3LabelCounterSample(name, label1Name, label2Name, label3Name string) *Map3LabelSample {
	return std.MustNew3LabelCounterSample(name, label1Name, label2Name, label3Name)
}

// MustNew3LabelCounterSample returns a composition with three fixed labels.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]*,
// (4) the label names do not appear in ascending order or
// (5) the labels are already in use.
func (r *Register) MustNew3LabelCounterSample(name, label1Name, label2Name, label3Name string) *Map3LabelSample {
	mustValidName(name)
	mustValidLabelName(label1Name)
	mustValidLabelName(label2Name)
	mustValidLabelName(label3Name)
	if label1Name > label2Name || label2Name > label3Name {
		panic("metrics: label name arguments aren't sorted")
	}

	r.mutex.Lock()

	var l3 *Map3LabelSample
	if index, ok := r.indices[name]; !ok {
		l3 = &Map3LabelSample{map3Label: map3Label{
			name: name, labelNames: [...]string{label1Name, label2Name, label3Name}}}

		r.indices[name] = uint32(len(r.metrics))
		r.metrics = append(r.metrics, &metric{
			typeComment: typePrefix + name + counterTypeLineEnd,
			sampleL3s:   []*Map3LabelSample{l3},
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
		l3 = &Map3LabelSample{map3Label: map3Label{
			name: name, labelNames: [...]string{label1Name, label2Name, label3Name}}}
		m.sampleL3s = append(m.sampleL3s, l3)
	}

	r.mutex.Unlock()
	return l3
}

// MustNew1LabelGaugeSample returns a composition with one fixed label.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) the label is already in use.
func MustNew1LabelGaugeSample(name, labelName string) *Map1LabelSample {
	return std.MustNew1LabelGaugeSample(name, labelName)
}

// MustNew1LabelGaugeSample returns a composition with one fixed label.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]* or
// (4) the label is already in use.
func (r *Register) MustNew1LabelGaugeSample(name, labelName string) *Map1LabelSample {
	mustValidName(name)
	mustValidLabelName(labelName)

	r.mutex.Lock()

	var l1 *Map1LabelSample
	if index, ok := r.indices[name]; !ok {
		l1 = &Map1LabelSample{map1Label: map1Label{
			name: name, labelName: labelName}}

		r.indices[name] = uint32(len(r.metrics))
		r.metrics = append(r.metrics, &metric{
			typeComment: typePrefix + name + gaugeTypeLineEnd,
			sampleL1s:   []*Map1LabelSample{l1},
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
		l1 = &Map1LabelSample{map1Label: map1Label{
			name: name, labelName: labelName}}
		m.sampleL1s = append(m.sampleL1s, l1)
	}

	r.mutex.Unlock()
	return l1
}

// MustNew2LabelGaugeSample returns a composition with two fixed labels.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]*,
// (4) the label names do not appear in ascending order or
// (5) the labels are already in use.
func MustNew2LabelGaugeSample(name, label1Name, label2Name string) *Map2LabelSample {
	return std.MustNew2LabelGaugeSample(name, label1Name, label2Name)
}

// MustNew2LabelGaugeSample returns a composition with two fixed labels.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]*,
// (4) the label names do not appear in ascending order or
// (5) the labels are already in use.
func (r *Register) MustNew2LabelGaugeSample(name, label1Name, label2Name string) *Map2LabelSample {
	mustValidName(name)
	mustValidLabelName(label1Name)
	mustValidLabelName(label2Name)
	if label1Name > label2Name {
		panic("metrics: label name arguments aren't sorted")
	}

	r.mutex.Lock()

	var l2 *Map2LabelSample
	if index, ok := r.indices[name]; !ok {
		l2 = &Map2LabelSample{map2Label: map2Label{
			name: name, labelNames: [...]string{label1Name, label2Name}}}

		r.indices[name] = uint32(len(r.metrics))
		r.metrics = append(r.metrics, &metric{
			typeComment: typePrefix + name + gaugeTypeLineEnd,
			sampleL2s:   []*Map2LabelSample{l2},
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
		l2 = &Map2LabelSample{map2Label: map2Label{
			name: name, labelNames: [...]string{label1Name, label2Name}}}
		m.sampleL2s = append(m.sampleL2s, l2)
	}

	r.mutex.Unlock()
	return l2
}

// MustNew3LabelGaugeSample returns a composition with three fixed labels.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]*,
// (4) the label names do not appear in ascending order or
// (5) the labels are already in use.
func MustNew3LabelGaugeSample(name, label1Name, label2Name, label3Name string) *Map3LabelSample {
	return std.MustNew3LabelGaugeSample(name, label1Name, label2Name, label3Name)
}

// MustNew3LabelGaugeSample returns a composition with three fixed labels.
// The function panics on any of the following:
// (1) name in use as another metric type,
// (2) name does not match regular expression [a-zA-Z_:][a-zA-Z0-9_:]*,
// (3) a label name does not match regular expression [a-zA-Z_][a-zA-Z0-9_]*,
// (4) the label names do not appear in ascending order or
// (5) the labels are already in use.
func (r *Register) MustNew3LabelGaugeSample(name, label1Name, label2Name, label3Name string) *Map3LabelSample {
	mustValidName(name)
	mustValidLabelName(label1Name)
	mustValidLabelName(label2Name)
	mustValidLabelName(label3Name)
	if label1Name > label2Name || label2Name > label3Name {
		panic("metrics: label name arguments aren't sorted")
	}

	r.mutex.Lock()

	var l3 *Map3LabelSample
	if index, ok := r.indices[name]; !ok {
		l3 = &Map3LabelSample{map3Label: map3Label{
			name: name, labelNames: [...]string{label1Name, label2Name, label3Name}}}

		r.indices[name] = uint32(len(r.metrics))
		r.metrics = append(r.metrics, &metric{
			typeComment: typePrefix + name + gaugeTypeLineEnd,
			sampleL3s:   []*Map3LabelSample{l3},
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
		l3 = &Map3LabelSample{map3Label: map3Label{
			name: name, labelNames: [...]string{label1Name, label2Name, label3Name}}}
		m.sampleL3s = append(m.sampleL3s, l3)
	}

	r.mutex.Unlock()
	return l3
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
