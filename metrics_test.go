package metrics

import (
	"mime"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func reset() {
	SkipTimestamp = false

	indices = make(map[string]uint32)
	metrics = nil
}

func TestSerialize(t *testing.T) {
	defer reset()
	SkipTimestamp = true

	MustPlaceGauge("g1").Set(42)
	MustPlaceCounter("c1").Add(1)
	MustPlaceCounter("c1").Add(8)
	MustHelp("g1", "🆘")
	MustHelp("c1", "override first 1")
	MustHelp("c1", "escape\n… and \\")

	rec := httptest.NewRecorder()
	HTTPHandler(rec, httptest.NewRequest("GET", "/metrics", nil))

	contentType := rec.Result().Header.Get("Content-Type")
	if media, params, err := mime.ParseMediaType(contentType); err != nil {
		t.Errorf("malformed content type %q: %s", contentType, err)
	} else if media != "text/plain" {
		t.Errorf("got content type %q, want plain text", contentType)
	} else if params["version"] != "0.0.4" {
		t.Errorf("got content type %q, want version 0.0.4", contentType)
	} else if params["charset"] != "UTF-8" {
		t.Errorf("got content type %q, want UTF-8 charset", contentType)
	}

	const want = `# TYPE g1 gauge
# HELP g1 🆘
g1 42
# TYPE c1 counter
# HELP c1 escape\n… and \\
c1 9
`
	if got := rec.Body.String(); got != want {
		t.Errorf("got %q", got)
		t.Errorf("want %q", want)
	}
}

func TestHTTPMethods(t *testing.T) {
	rec := httptest.NewRecorder()
	HTTPHandler(rec, httptest.NewRequest("POST", "/metrics", nil))
	got := rec.Result()

	if got.StatusCode != 405 {
		t.Errorf("got status code %d, want 405", got.StatusCode)
	}

	allow := got.Header.Get("Allow")
	if !strings.Contains(allow, "OPTIONS") || !strings.Contains(allow, "GET") || !strings.Contains(allow, "HEAD") {
		t.Errorf("got allow %q, want OPTIONS, GET and HEAD", allow)
	}
}

func BenchmarkMustPlace(b *testing.B) {
	defer reset()

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		MustPlaceGauge("some_arbitrary_test_unit")
	}
}

func BenchmarkHelp(b *testing.B) {
	defer reset()

	b.ReportAllocs()

	MustPlaceGauge("some_arbitrary_test_unit")
	for i := 0; i < b.N; i++ {
		MustHelp("some_arbitrary_test_unit", "some arbitrary test text")
	}
}

func BenchmarkParallelAdd(b *testing.B) {
	defer reset()

	b.Run("integer", func(b *testing.B) {
		b.ReportAllocs()

		c := MustPlaceCounter("bench_integer_unit")
		b.RunParallel(func(pb *testing.PB) {
			for i := uint64(0); pb.Next(); i++ {
				c.Add(i)
			}
		})
	})

	b.Run("real", func(b *testing.B) {
		b.ReportAllocs()

		c := MustPlaceGauge("bench_real_unit")
		b.RunParallel(func(pb *testing.PB) {
			for i := 0; pb.Next(); i++ {
				c.Add(float64(i))
			}
		})
	})

	values := [...]string{"first", "second", "third", "fourth", "fifth"}

	b.Run("label2x5", func(b *testing.B) {
		b.ReportAllocs()

		g := MustPlaceGaugeLabel2("bench_label_unit", "label", "more")
		b.RunParallel(func(pb *testing.PB) {
			for i := 0; pb.Next(); i++ {
				g.Add(float64(i), values[i%2], values[i%5])
			}
		})
	})
}

type voidResponseWriter http.Header

func (void voidResponseWriter) Header() http.Header    { return http.Header(void) }
func (voidResponseWriter) Write(p []byte) (int, error) { return len(p), nil }
func (voidResponseWriter) WriteHeader(statusCode int)  {}

func benchmarkHTTPHandler(b *testing.B) {
	b.ReportAllocs()

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
				MustPlaceCounter("integer" + strconv.Itoa(i) + "_bench_unit").Add(uint64(i))
			}
			b.Run("integer", benchmarkHTTPHandler)

			reset()
			for i := n; i > 0; i-- {
				MustPlaceGauge("real" + strconv.Itoa(i) + "_bench_unit").Set(float64(i))
			}
			b.Run("real", benchmarkHTTPHandler)

			reset()
			for i := n; i > 0; i-- {
				MustPlaceGaugeLabel1("real"+strconv.Itoa(i)+"_label_bench_unit", "first").Set(float64(i), strconv.Itoa(i%5))
			}
			b.Run("label5", benchmarkHTTPHandler)

			reset()
			for i := n; i > 0; i-- {
				MustPlaceGaugeLabel3("real"+strconv.Itoa(i)+"_3label_bench_unit", "first", "second", "third").Set(float64(i), strconv.Itoa(i%2), strconv.Itoa(i%3), strconv.Itoa(i%5))
			}
			b.Run("label2x3x5", benchmarkHTTPHandler)
		})
	}
}
