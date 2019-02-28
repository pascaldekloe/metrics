package metrics

import "testing"

func BenchmarkLabel(b *testing.B) {
	reg := NewRegister()

	values := [...]string{"one", "two", "three", "four"}
	label1 := reg.Must1LabelReal("bench_label_unit", "first")
	label2 := reg.Must2LabelReal("bench_label_unit", "first", "second")
	label3 := reg.Must3LabelReal("bench_label_unit", "first", "second", "third")

	b.Run("sequential", func(b *testing.B) {
		b.Run("4", func(b *testing.B) {
			for i := 0; i < b.N; i += 4 {
				for _, v := range values {
					label1(v)
				}
			}
		})
		b.Run("4x4", func(b *testing.B) {
			for i := 0; i < b.N; i += 4 * 4 {
				for _, v1 := range values {
					for _, v2 := range values {
						label2(v1, v2)
					}
				}
			}
		})
		b.Run("4x4x4", func(b *testing.B) {
			for i := 0; i < b.N; i += 4 * 4 * 4 {
				for _, v1 := range values {
					for _, v2 := range values {
						for _, v3 := range values {
							label3(v1, v2, v3)
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
						label1(v)
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
							label2(v1, v2)
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
								label3(v1, v2, v3)
							}
						}
					}
				}
			})
		})
	})
}
