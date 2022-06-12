package metrics_test

import (
	"bytes"
	"math"
	"mime"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pascaldekloe/metrics"
)

func TestWriteTo(t *testing.T) {
	metrics.SkipTimestamp = true
	reg := metrics.NewRegister()

	var buf bytes.Buffer
	n, err := reg.WriteTo(&buf)
	if err != nil {
		t.Fatal("got error:", err)
	}
	if n != int64(buf.Len()) {
		t.Errorf("n = %d with %d bytes written", n, buf.Len())
	}
	if got, want := buf.String(), "# Prometheus Samples\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	reg.Must1LabelReal("v", "sign")("π").Set(math.Pi)

	buf.Reset()
	n, err = reg.WriteTo(&buf)
	if err != nil {
		t.Fatal("got error:", err)
	}
	if n != int64(buf.Len()) {
		t.Errorf("n = %d with %d bytes written", n, buf.Len())
	}
	const want = `# Prometheus Samples

# TYPE v gauge
v{sign="π"} 3.141592653589793
`
	if got := buf.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestServeHTTP(t *testing.T) {
	metrics.SkipTimestamp = true
	reg := metrics.NewRegister()

	m1 := reg.MustReal("m1", "🆘")
	m1.Set(42)
	m2 := reg.MustCounter("m2", "escape\n… and \\")
	m2.Add(1)
	m2.Add(8)

	rec := httptest.NewRecorder()
	reg.ServeHTTP(rec, httptest.NewRequest("GET", "/metrics", nil))

	contentType := rec.Result().Header.Get("Content-Type")
	if media, params, err := mime.ParseMediaType(contentType); err != nil {
		t.Errorf("malformed content type %q: %s", contentType, err)
	} else if media != "text/plain" {
		t.Errorf("got content type %q, want plain text", contentType)
	} else if params["version"] != "0.0.4" {
		t.Errorf("got content type %q, want version 0.0.4", contentType)
	}

	const want = `# Prometheus Samples

# TYPE m1 gauge
# HELP m1 🆘
m1 42

# TYPE m2 counter
# HELP m2 escape\n… and \\
m2 9
`
	if got := rec.Body.String(); got != want {
		t.Errorf("got %q", got)
		t.Errorf("want %q", want)
	}
}

func TestHTTPMethods(t *testing.T) {
	rec := httptest.NewRecorder()
	metrics.NewRegister().ServeHTTP(rec, httptest.NewRequest("POST", "/metrics", nil))
	got := rec.Result()

	if got.StatusCode != 405 {
		t.Errorf("got status code %d, want 405", got.StatusCode)
	}

	allow := got.Header.Get("Allow")
	if !strings.Contains(allow, "OPTIONS") || !strings.Contains(allow, "GET") || !strings.Contains(allow, "HEAD") {
		t.Errorf("got allow %q, want OPTIONS, GET and HEAD", allow)
	}
}

type voidResponseWriter int64

func (w *voidResponseWriter) Header() http.Header        { return http.Header{} }
func (w *voidResponseWriter) WriteHeader(statusCode int) {}
func (w *voidResponseWriter) Write(p []byte) (int, error) {
	*w += voidResponseWriter(len(p))
	return len(p), nil
}
func (w *voidResponseWriter) WriteString(s string) (int, error) {
	*w += voidResponseWriter(len(s))
	return len(s), nil
}

func BenchmarkServeHTTP(b *testing.B) {
	var reg *metrics.Register
	benchmarkHTTPHandler := func(b *testing.B) {
		req := httptest.NewRequest("GET", "/metrics", nil)
		var w voidResponseWriter
		for i := 0; i < b.N; i++ {
			reg.ServeHTTP(&w, req)
		}
		b.SetBytes(int64(w) / int64(b.N))
	}

	for _, n := range []int{32, 1024, 32768} {
		b.Run(strconv.Itoa(n), func(b *testing.B) {
			reg = metrics.NewRegister()
			for i := n; i > 0; i-- {
				reg.MustCounter("integer"+strconv.Itoa(i)+"_bench_unit", "").Add(uint64(i))
			}
			b.Run("counter", benchmarkHTTPHandler)

			reg = metrics.NewRegister()
			for i := n; i > 0; i-- {
				reg.MustReal("real"+strconv.Itoa(i)+"_bench_unit", "").Set(float64(i))
			}
			b.Run("real", benchmarkHTTPHandler)

			reg = metrics.NewRegister()
			for i := n; i > 0; i-- {
				reg.MustInteger("integer"+strconv.Itoa(i)+"_bench_unit", "").Set(int64(i))
			}
			b.Run("integer", benchmarkHTTPHandler)

			reg = metrics.NewRegister()
			for i := n; i > 0; i-- {
				reg.MustHistogram("histogram"+strconv.Itoa(i)+"_bench_unit", "", 1, 2, 3, 4).Add(3.14)
			}
			b.Run("histogram5", benchmarkHTTPHandler)

			reg = metrics.NewRegister()
			for i := n; i > 0; i-- {
				reg.Must1LabelReal("real"+strconv.Itoa(i)+"_label_bench_unit", "first")(strconv.Itoa(i % 5)).Set(float64(i))
			}
			b.Run("label5", benchmarkHTTPHandler)

			reg = metrics.NewRegister()
			for i := n; i > 0; i-- {
				reg.Must3LabelReal("real"+strconv.Itoa(i)+"_3label_bench_unit", "first", "second", "third")(strconv.Itoa(i%2), strconv.Itoa(i%3), strconv.Itoa(i%5)).Set(float64(i))
			}
			b.Run("label2x3x5", benchmarkHTTPHandler)

			reg = metrics.NewRegister()
			for i := n; i > 0; i-- {
				reg.MustRealSample("real"+strconv.Itoa(i)+"_bench_unit", "").Set(float64(i), time.Now())
			}
			b.Run("sample", benchmarkHTTPHandler)
		})
	}
}
