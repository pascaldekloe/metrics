package metrics

import (
	"mime"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestServeHTTP(t *testing.T) {
	SkipTimestamp = true
	reg := NewRegister()

	g1 := reg.MustNewGauge("g1")
	reg.MustHelp("g1", "🆘")
	g1.Set(42)
	c1 := reg.MustNewCounter("c1")
	c1.Add(1)
	c1.Add(8)
	reg.MustHelp("c1", "escape\n… and \\")

	rec := httptest.NewRecorder()
	reg.ServeHTTP(rec, httptest.NewRequest("GET", "/metrics", nil))

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

	const want = `# Prometheus Samples

# TYPE g1 gauge
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
	NewRegister().ServeHTTP(rec, httptest.NewRequest("POST", "/metrics", nil))
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

func BenchmarkServeHTTP(b *testing.B) {
	var reg *Register
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
			reg = NewRegister()
			for i := n; i > 0; i-- {
				reg.MustNewCounter("integer" + strconv.Itoa(i) + "_bench_unit").Add(uint64(i))
			}
			b.Run("counter", benchmarkHTTPHandler)

			reg = NewRegister()
			for i := n; i > 0; i-- {
				reg.MustNewGauge("real" + strconv.Itoa(i) + "_bench_unit").Set(float64(i))
			}
			b.Run("gauge", benchmarkHTTPHandler)

			reg = NewRegister()
			for i := n; i > 0; i-- {
				reg.MustNewHistogram("histogram"+strconv.Itoa(i)+"_bench_unit", 1, 2, 3, 4).Add(3.14)
			}
			b.Run("histogram5", benchmarkHTTPHandler)

			reg = NewRegister()
			for i := n; i > 0; i-- {
				reg.Must1LabelGauge("real"+strconv.Itoa(i)+"_label_bench_unit", "first").With(strconv.Itoa(i % 5)).Set(float64(i))
			}
			b.Run("label5", benchmarkHTTPHandler)

			reg = NewRegister()
			for i := n; i > 0; i-- {
				reg.Must3LabelGauge("real"+strconv.Itoa(i)+"_3label_bench_unit", "first", "second", "third").With(strconv.Itoa(i%2), strconv.Itoa(i%3), strconv.Itoa(i%5)).Set(float64(i))
			}
			b.Run("label2x3x5", benchmarkHTTPHandler)

			reg = NewRegister()
			for i := n; i > 0; i-- {
				reg.MustNewGaugeSample("sample"+strconv.Itoa(i)+"_bench_unit").Set(float64(i), time.Now())
			}
			b.Run("sample", benchmarkHTTPHandler)
		})
	}
}
