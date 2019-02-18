package metrics

import "testing"

func BenchmarkLabelWith(b *testing.B) {
	reg := NewRegister()

	values := [...]string{"first", "second", "third", "fourth"}
	g1 := reg.Must1LabelGauge("bench_label_unit", "first")
	g2 := reg.Must2LabelGauge("bench_label_unit", "first", "second")
	g3 := reg.Must3LabelGauge("bench_label_unit", "first", "second", "third")

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
