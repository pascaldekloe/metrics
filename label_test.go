package metrics_test

import (
	"os"
	"testing"

	"github.com/pascaldekloe/metrics"
)

// TestLabels verifies the codec roundtrip.
func TestLabels(t *testing.T) {
	reg := metrics.NewRegister()

	switch got := reg.Must1LabelCounter("plain", "foo")("bar").Labels(); {
	case len(got) != 1, got["foo"] != "bar":
		t.Errorf(`got %q, want {"foo": "bar"}`, got)
	}

	switch got := reg.Must3LabelReal("escapes", "a", "b", "c")("\\", "\n", "\"").Labels(); {
	case len(got) != 3, got["a"] != "\\", got["b"] != "\n", got["c"] != "\"":
		t.Errorf(`got %q, want {"a": "\\", "b": "\n", "c": "\""}`, got)
	}

	// values may contain *any* byte sequence
	var all [256]byte
	for i := range all {
		all[i] = byte(i)
	}
	switch got := reg.Must1LabelCounter("raw", "foo")(string(all[:])).Labels(); {
	case len(got) != 1, got["foo"] != string(all[:]):
		t.Errorf(`got %q, want {"foo": %q}`, got, all)
	}
}

func Example_labels() {
	// setup
	demo := metrics.NewRegister()
	Building := demo.Must2LabelInteger("hitpoints_total", "ground", "building")
	Arsenal := demo.Must3LabelInteger("hitpoints_total", "ground", "arsenal", "side")
	demo.MustHelp("hitpoints_total", "Damage Capacity")

	// measures
	Building("Genesis Pit", "Civilian Hospital").Set(800)
	Arsenal("Genesis Pit", "Tech Center", "Nod").Set(500)
	Arsenal("Genesis Pit", "Cyborg", "Nod").Set(900)
	Arsenal("Genesis Pit", "Cyborg", "Nod").Add(-596)
	Building("Genesis Pit", "Civilian Hospital").Add(-490)
	Arsenal("Genesis Pit", "Cyborg", "Nod").Add(110)

	// print
	metrics.SkipTimestamp = true
	demo.WriteTo(os.Stdout)
	// Output:
	// # Prometheus Samples
	//
	// # TYPE hitpoints_total gauge
	// # HELP hitpoints_total Damage Capacity
	// hitpoints_total{building="Civilian Hospital",ground="Genesis Pit"} 310
	// hitpoints_total{arsenal="Tech Center",ground="Genesis Pit",side="Nod"} 500
	// hitpoints_total{arsenal="Cyborg",ground="Genesis Pit",side="Nod"} 414
}

func BenchmarkLabel(b *testing.B) {
	values := [...]string{"one", "two", "three", "four"}
	label1 := metrics.NewRegister().Must1LabelReal("bench_label_unit", "first")
	label2 := metrics.NewRegister().Must2LabelReal("bench_label_unit", "first", "second")
	label3 := metrics.NewRegister().Must3LabelReal("bench_label_unit", "first", "second", "third")

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
