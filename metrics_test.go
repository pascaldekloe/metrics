package metrics

import (
	"bytes"
	"io"
	"math"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestHelp(t *testing.T) {
	reg := NewRegister()

	reg.MustReal("g", "set on gauge")
	reg.Must1LabelReal("lm", "l")
	reg.MustHelp("lm", "set on map to override")
	reg.Must2LabelReal("lm", "l1", "l2")
	reg.MustHelp("lm", "override on map")
	reg.Must3LabelReal("lg", "l1", "l2", "l3")("v1", "v2", "v3")
	reg.MustHelp("lg", "set on labeled gauge to override")
	reg.Must3LabelReal("lg", "l4", "l5", "l6")("v4", "v5", "v6")
	reg.MustHelp("lg", "override on labeled gauge")

	want := map[string]string{
		"g":  "set on gauge",
		"lm": "override on map",
		"lg": "override on labeled gauge",
	}

	var buf bytes.Buffer
	reg.WriteText(&buf)

	got := make(map[string]string)
	for {
		line, err := buf.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}

		if !strings.HasPrefix(line, "# HELP ") {
			continue
		}

		split := strings.IndexByte(line[7:], ' ')
		if split < 0 {
			t.Errorf("malformed help comment %q", line)
			continue
		}
		got[line[7:7+split]] = line[8+split : len(line)-1]
	}

	for name, help := range want {
		if s := got[name]; s != help {
			t.Errorf("got %q for %q, want %q", s, name, help)
		}
	}
}

func TestHistogramBuckets(t *testing.T) {
	reg := NewRegister()

	var golden = []struct {
		feed []float64
		want []float64
	}{
		{[]float64{4, 1, 2}, []float64{1, 2, 4}},
		{[]float64{8, math.Inf(1)}, []float64{8}},
		{[]float64{math.Inf(-1), 8}, []float64{8}},
		{[]float64{math.NaN(), 7, math.Inf(1), 3}, []float64{3, 7}},
	}

	for i, gold := range golden {
		h := reg.MustHistogram("h"+strconv.Itoa(i), "", gold.feed...)
		if !reflect.DeepEqual(h.BucketBounds, gold.want) {
			t.Errorf("%v: got buckets %v, want %v", gold.feed, h.BucketBounds, gold.want)
		}
	}
}

func BenchmarkGet(b *testing.B) {
	b.Run("histogram5", func(b *testing.B) {
		h := NewRegister().MustHistogram("bench_histogram_unit", "", .01, .02, .05, .1)

		b.Run("sequential", func(b *testing.B) {
			var buckets []uint64
			for i := 0; i < b.N; i++ {
				buckets, _, _ = h.Get(buckets[:0])
			}
		})

		b.Run("2routines", func(b *testing.B) {
			done := make(chan []uint64)
			f := func() {
				var buckets []uint64
				for i := b.N / 2; i >= 0; i-- {
					buckets, _, _ = h.Get(buckets[:0])
				}
				done <- buckets
			}
			go f()
			go f()
			<-done
			<-done
		})
	})
}

func BenchmarkSet(b *testing.B) {
	b.Run("real", func(b *testing.B) {
		m := NewRegister().MustReal("bench_real_unit", "")

		b.Run("sequential", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				m.Set(42)
			}
		})
		b.Run("2routines", func(b *testing.B) {
			done := make(chan struct{})
			f := func() {
				for i := b.N / 2; i >= 0; i-- {
					m.Set(42)
				}
				done <- struct{}{}
			}
			go f()
			go f()
			<-done
			<-done
		})
	})

	b.Run("sample", func(b *testing.B) {
		m := NewRegister().MustRealSample("bench_sample_unit", "")
		timestamp := time.Now()

		b.Run("sequential", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				m.Set(42, timestamp)
			}
		})
		b.Run("2routines", func(b *testing.B) {
			done := make(chan struct{})
			f := func() {
				for i := b.N / 2; i >= 0; i-- {
					m.Set(42, timestamp)
				}
				done <- struct{}{}
			}
			go f()
			go f()
			<-done
			<-done
		})
	})
}

func BenchmarkAdd(b *testing.B) {
	b.Run("counter", func(b *testing.B) {
		m := NewRegister().MustCounter("bench_counter_unit", "")

		b.Run("sequential", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				m.Add(1)
			}
		})
		b.Run("2routines", func(b *testing.B) {
			done := make(chan struct{})
			f := func() {
				for i := b.N / 2; i >= 0; i-- {
					m.Add(1)
				}
				done <- struct{}{}
			}
			go f()
			go f()
			<-done
			<-done
		})
	})

	b.Run("integer", func(b *testing.B) {
		m := NewRegister().MustInteger("bench_gauge_unit", "")

		b.Run("sequential", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				m.Add(1)
			}
		})
		b.Run("2routines", func(b *testing.B) {
			done := make(chan struct{})
			f := func() {
				for i := b.N / 2; i >= 0; i-- {
					m.Add(1)
				}
				done <- struct{}{}
			}
			go f()
			go f()
			<-done
			<-done
		})
	})

	b.Run("histogram5", func(b *testing.B) {
		h := NewRegister().MustHistogram("bench_histogram_unit", "", .01, .02, .05, .1)

		b.Run("sequential", func(b *testing.B) {
			f := .001
			for i := 0; i < b.N; i++ {
				h.Add(f)
				f += .001
			}
		})
		b.Run("2routines", func(b *testing.B) {
			done := make(chan struct{})
			f := func() {
				f := .001
				for i := b.N / 2; i >= 0; i-- {
					h.Add(f)
					f += .001
				}
				done <- struct{}{}
			}
			go f()
			go f()
			<-done
			<-done
		})
	})
}
