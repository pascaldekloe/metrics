package metrics

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func BenchmarkWithLabels(b *testing.B) {
	values := [...]string{"first", "second", "third", "fourth"}
	g1 := Must1LabelGauge("bench_label_unit", "label1")
	g2 := Must2LabelGauge("bench_label_unit", "label1", "label2")
	g3 := Must3LabelGauge("bench_label_unit", "label1", "label2", "label3")

	b.Run("sequential", func(b *testing.B) {
		b.Run("4", func(b *testing.B) {
			for i := 0; i < b.N; i += 4 {
				for _, v := range values {
					g1.With(v)
				}
			}
		})
		b.Run("4x4", func(b *testing.B) {
			for i := 0; i < b.N; i += 4 * 4 {
				for _, v1 := range values {
					for _, v2 := range values {
						g2.With(v1, v2)
					}
				}
			}
		})
		b.Run("4x4x4", func(b *testing.B) {
			for i := 0; i < b.N; i += 4 * 4 * 4 {
				for _, v1 := range values {
					for _, v2 := range values {
						for _, v3 := range values {
							g3.With(v1, v2, v3)
						}
					}
				}
			}
		})
	})

	b.Run("parallel", func(b *testing.B) {
		b.Run("4", func(b *testing.B) {
			b.RunParallel(func(pb *testing.PB) {
				for {
					for _, v := range values {
						if !pb.Next() {
							return
						}
						g1.With(v)
					}
				}
			})
		})
		b.Run("4x4", func(b *testing.B) {
			b.RunParallel(func(pb *testing.PB) {
				for {
					for _, v1 := range values {
						for _, v2 := range values {
							if !pb.Next() {
								return
							}
							g2.With(v1, v2)
						}
					}
				}
			})
		})
		b.Run("4x4x4", func(b *testing.B) {
			b.RunParallel(func(pb *testing.PB) {
				for {
					for _, v1 := range values {
						for _, v2 := range values {
							for _, v3 := range values {
								if !pb.Next() {
									return
								}
								g3.With(v1, v2, v3)
							}
						}
					}
				}
			})
		})
	})
}

func BenchmarkGet(b *testing.B) {
	defer reset()

	b.Run("counter", func(b *testing.B) {
		c := MustNewCounter("bench_integer_unit")

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
		g := MustNewGauge("bench_real_unit")

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
}

func BenchmarkSet(b *testing.B) {
	defer reset()

	b.Run("gauge", func(b *testing.B) {
		g := MustNewGauge("bench_real_unit")

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
}

func BenchmarkAdd(b *testing.B) {
	defer reset()

	b.Run("counter", func(b *testing.B) {
		c := MustNewCounter("bench_integer_unit")

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
		g := MustNewGauge("bench_real_unit")

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
		g := MustNewHistogram("bench_histogram_unit", .01, .02, .05, .1)

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

type voidResponseWriter http.Header

func (void voidResponseWriter) Header() http.Header    { return http.Header(void) }
func (voidResponseWriter) Write(p []byte) (int, error) { return len(p), nil }
func (voidResponseWriter) WriteHeader(statusCode int)  {}

func benchmarkHTTPHandler(b *testing.B) {
	req := httptest.NewRequest("GET", "/metrics", nil)
	resp := voidResponseWriter{}
	for i := 0; i < b.N; i++ {
		HTTPHandler(resp, req)
	}
}

func BenchmarkHTTPHandler(b *testing.B) {
	defer reset()

	for _, n := range []int{32, 1024, 32768} {
		b.Run(strconv.Itoa(n), func(b *testing.B) {
			reset()
			for i := n; i > 0; i-- {
				MustNewCounter("integer" + strconv.Itoa(i) + "_bench_unit").Add(uint64(i))
			}
			b.Run("counter", benchmarkHTTPHandler)

			reset()
			for i := n; i > 0; i-- {
				MustNewGauge("real" + strconv.Itoa(i) + "_bench_unit").Set(float64(i))
			}
			b.Run("gauge", benchmarkHTTPHandler)

			reset()
			for i := n; i > 0; i-- {
				MustNewHistogram("histogram"+strconv.Itoa(i)+"_bench_unit", 1, 2, 3, 4).Add(3.14)
			}
			b.Run("histogram5", benchmarkHTTPHandler)

			reset()
			for i := n; i > 0; i-- {
				Must1LabelGauge("real"+strconv.Itoa(i)+"_label_bench_unit", "first").With(strconv.Itoa(i % 5)).Set(float64(i))
			}
			b.Run("label5", benchmarkHTTPHandler)

			reset()
			for i := n; i > 0; i-- {
				Must3LabelGauge("real"+strconv.Itoa(i)+"_3label_bench_unit", "first", "second", "third").With(strconv.Itoa(i%2), strconv.Itoa(i%3), strconv.Itoa(i%5)).Set(float64(i))
			}
			b.Run("label2x3x5", benchmarkHTTPHandler)
		})
	}
}
