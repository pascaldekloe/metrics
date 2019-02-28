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

	reg.MustReal("g")
	reg.MustHelp("g", "set on gauge")
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
		h := reg.MustHistogram("h"+strconv.Itoa(i), gold.feed...)
		if !reflect.DeepEqual(h.BucketBounds, gold.want) {
			t.Errorf("%v: got buckets %v, want %v", gold.feed, h.BucketBounds, gold.want)
		}
	}
}

func BenchmarkGet(b *testing.B) {
	reg := NewRegister()

	b.Run("counter", func(b *testing.B) {
		c := reg.MustCounter("bench_integer_unit")

		b.Run("sequential", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				c.Get()
			}
		})
		b.Run("2routines", func(b *testing.B) {
			b.SetParallelism(2)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					c.Get()
				}
			})
		})
	})

	b.Run("gauge", func(b *testing.B) {
		g := reg.MustReal("bench_real_unit")

		b.Run("sequential", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				g.Get()
			}
		})
		b.Run("2routines", func(b *testing.B) {
			b.SetParallelism(2)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					g.Get()
				}
			})
		})
	})

	b.Run("sample", func(b *testing.B) {
		s := reg.MustGaugeSample("bench_sample_unit")

		b.Run("sequential", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				s.Get()
			}
		})
		b.Run("2routines", func(b *testing.B) {
			b.SetParallelism(2)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					s.Get()
				}
			})
		})
	})
}

func BenchmarkSet(b *testing.B) {
	reg := NewRegister()

	b.Run("gauge", func(b *testing.B) {
		g := reg.MustReal("bench_real_unit")

		b.Run("sequential", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				g.Set(42)
			}
		})
		b.Run("2routines", func(b *testing.B) {
			b.SetParallelism(2)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					g.Set(42)
				}
			})
		})
	})

	b.Run("sample", func(b *testing.B) {
		s := reg.MustGaugeSample("bench_sample_unit")
		timestamp := time.Now()

		b.Run("sequential", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				s.Set(42, timestamp)
			}
		})
		b.Run("2routines", func(b *testing.B) {
			b.SetParallelism(2)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					s.Set(42, timestamp)
				}
			})
		})
	})
}

func BenchmarkAdd(b *testing.B) {
	reg := NewRegister()

	b.Run("counter", func(b *testing.B) {
		c := reg.MustCounter("bench_integer_unit")

		b.Run("sequential", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				c.Add(1)
			}
		})
		b.Run("2routines", func(b *testing.B) {
			b.SetParallelism(2)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					c.Add(1)
				}
			})
		})
	})

	b.Run("gauge", func(b *testing.B) {
		g := reg.MustInteger("bench_real_unit")

		b.Run("sequential", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				g.Add(1)
			}
		})
		b.Run("2routines", func(b *testing.B) {
			b.SetParallelism(2)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					g.Add(1)
				}
			})
		})
	})

	b.Run("histogram5", func(b *testing.B) {
		g := reg.MustHistogram("bench_histogram_unit", .01, .02, .05, .1)

		b.Run("sequential", func(b *testing.B) {
			f := .001
			for i := 0; i < b.N; i++ {
				g.Add(f)
				f += .001
			}
		})
		b.Run("2routines", func(b *testing.B) {
			b.SetParallelism(2)
			b.RunParallel(func(pb *testing.PB) {
				f := .001
				for pb.Next() {
					g.Add(f)
					f += .001
				}
			})
		})
	})
}
