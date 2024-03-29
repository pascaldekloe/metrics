package metrics_test

import (
	"bytes"
	"io"
	"math"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pascaldekloe/metrics"
	"github.com/pascaldekloe/metrics/gostat"
)

func TestName(t *testing.T) {
	reg := metrics.NewRegister()

	if got := reg.MustCounter("c", "").Name(); got != "c" {
		t.Errorf(`counter got %q, want "c"`, got)
	}
	if got := reg.Must1LabelCounter("lc", "l")("v").Name(); got != "lc" {
		t.Errorf(`labeled counter got %q, want "lc"`, got)
	}

	if got := reg.MustInteger("i", "").Name(); got != "i" {
		t.Errorf(`integer got %q, want "i"`, got)
	}
	if got := reg.Must2LabelInteger("li", "l1", "l2")("v1", "v2").Name(); got != "li" {
		t.Errorf(`labeled integer got %q, want "li"`, got)
	}

	if got := reg.MustReal("r", "").Name(); got != "r" {
		t.Errorf(`real got %q, want "r"`, got)
	}
	if got := reg.Must3LabelReal("lr", "l1", "l2", "l3")("v1", "v2", "v3").Name(); got != "lr" {
		t.Errorf(`labeled real got %q, want "lr"`, got)
	}

	if got := reg.MustCounterSample("cs", "").Name(); got != "cs" {
		t.Errorf(`counter sample got %q, want "cs"`, got)
	}
	if got := reg.Must1LabelRealSample("lrs", "l")("v").Name(); got != "lrs" {
		t.Errorf(`labeled real sample got %q, want "lrs"`, got)
	}
}

func TestHelp(t *testing.T) {
	reg := metrics.NewRegister()

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
	reg.WriteTo(&buf)

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

var (
	LogSize = metrics.MustRealSample("log_bytes", "Size reported by the filesystem.")
	LogIdle = metrics.MustRealSample("log_idle_seconds", "Duration since last change.")
)

func ExampleSample_lazy() {
	// mount exposition point
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		// update standard samples
		gostat.Capture()

		// update custom samples
		log, err := os.Stat("./mission.log")
		if err == nil { // ⚠️ reverse error check
			now := time.Now()
			LogSize.Set(float64(log.Size()), now)
			LogIdle.SetSeconds(now.Sub(log.ModTime()), now)
		}

		// serve serialized
		metrics.ServeHTTP(w, r)
	})
	// Output:
}

func ExampleHistogram() {
	// setup
	demo := metrics.NewRegister()
	Duration := demo.Must2LabelHistogram("http_latency_seconds", "method", "status", 0.001, 0.005, 0.025, 0.125)
	demo.MustHelp("http_latency_seconds", "Time from request initiation until response body retrieval.")

	// measures
	Duration("GET", "2xx").Add(0.076875)
	Duration("GET", "3xx").Add(0.000141)
	Duration("OPTIONS", "2xx").Add(0.000009)
	Duration("GET", "2xx").Add(0.002277)
	Duration("GET", "2xx").Add(0.001871)
	Duration("GET", "2xx").Add(0.002378)

	// print
	metrics.SkipTimestamp = true
	demo.WriteTo(os.Stdout)
	// Output:
	// # Prometheus Samples
	//
	// # TYPE http_latency_seconds histogram
	// # HELP http_latency_seconds Time from request initiation until response body retrieval.
	// http_latency_seconds_count{method="GET",status="2xx"} 4
	// http_latency_seconds{le="0.001",method="GET",status="2xx"} 0
	// http_latency_seconds{le="0.005",method="GET",status="2xx"} 3
	// http_latency_seconds{le="0.025",method="GET",status="2xx"} 3
	// http_latency_seconds{le="0.125",method="GET",status="2xx"} 4
	// http_latency_seconds{le="+Inf",method="GET",status="2xx"} 4
	// http_latency_seconds_sum{method="GET",status="2xx"} 0.083401
	// http_latency_seconds_count{method="GET",status="3xx"} 1
	// http_latency_seconds{le="0.001",method="GET",status="3xx"} 1
	// http_latency_seconds{le="0.005",method="GET",status="3xx"} 1
	// http_latency_seconds{le="0.025",method="GET",status="3xx"} 1
	// http_latency_seconds{le="0.125",method="GET",status="3xx"} 1
	// http_latency_seconds{le="+Inf",method="GET",status="3xx"} 1
	// http_latency_seconds_sum{method="GET",status="3xx"} 0.000141
	// http_latency_seconds_count{method="OPTIONS",status="2xx"} 1
	// http_latency_seconds{le="0.001",method="OPTIONS",status="2xx"} 1
	// http_latency_seconds{le="0.005",method="OPTIONS",status="2xx"} 1
	// http_latency_seconds{le="0.025",method="OPTIONS",status="2xx"} 1
	// http_latency_seconds{le="0.125",method="OPTIONS",status="2xx"} 1
	// http_latency_seconds{le="+Inf",method="OPTIONS",status="2xx"} 1
	// http_latency_seconds_sum{method="OPTIONS",status="2xx"} 9e-06
}

func TestHistogramBuckets(t *testing.T) {
	reg := metrics.NewRegister()

	var golden = []struct {
		feed []float64
		want []float64
	}{
		{},
		{[]float64{4, 1, 2}, []float64{1, 2, 4}},
		{[]float64{8, math.Inf(1)}, []float64{8}},
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
		h := metrics.NewRegister().MustHistogram("bench_histogram_unit", "", .01, .02, .05, .1)

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
		m := metrics.NewRegister().MustReal("bench_real_unit", "")

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
		m := metrics.NewRegister().MustRealSample("bench_sample_unit", "")
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
		m := metrics.NewRegister().MustCounter("bench_counter_unit", "")

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
		m := metrics.NewRegister().MustInteger("bench_gauge_unit", "")

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
		h := metrics.NewRegister().MustHistogram("bench_histogram_unit", "", 1, 2, 5, 6)

		b.Run("sequential", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				h.Add(float64(i & 7))
			}
		})
		b.Run("2routines", func(b *testing.B) {
			done := make(chan struct{})
			f := func() {
				for i := b.N / 2; i >= 0; i-- {
					h.Add(float64(i & 7))
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
