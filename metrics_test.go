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
	gauges = nil
	gaugeIndices = make(map[string]int)
	counters = nil
	counterIndices = make(map[string]int)
}

func TestSerialize(t *testing.T) {
	// mockup
	backup := appendTimeTail
	appendTimeTail = func(buf []byte) []byte {
		return append(buf, "1548759822954\n"...)
	}
	defer func() {
		appendTimeTail = backup

		reset()
	}()

	MustPlaceGauge("g1").Help("🆘")
	MustPlaceGauge("g1").Set(42)
	MustPlaceCounter("c1").Help("override first 1").Add(1)
	MustPlaceCounter("c1").Help("escape\n… and \\").Add(8)

	rec := httptest.NewRecorder()
	HTTPHandler(rec, httptest.NewRequest("GET", "/metrics", nil))

	contentType := rec.Result().Header.Get("Content-Type")
	if media, params, err := mime.ParseMediaType(contentType); err != nil {
		t.Errorf("malformed content type %q: %s", contentType, err)
	} else if media != "text/plain" {
		t.Errorf("got content type %q, want plain text", contentType)
	} else if params["charset"] != "UTF-8" {
		t.Errorf("got content type %q, want UTF-8 charset", contentType)
	}

	const want = `# HELP g1 🆘
# TYPE g1 gauge
g1 42 1548759822954
# HELP c1 escape\n… and \\
# TYPE c1 counter
c1 9 1548759822954
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

	g := MustPlaceGauge("some_arbitrary_test_unit")
	for i := 0; i < b.N; i++ {
		g.Help("some arbitrary test text")
	}
}

type voidResponseWriter http.Header

func (void voidResponseWriter) Header() http.Header    { return http.Header(void) }
func (voidResponseWriter) Write(p []byte) (int, error) { return len(p), nil }
func (voidResponseWriter) WriteHeader(statusCode int)  {}

func BenchmarkHTTPHandler(b *testing.B) {
	defer reset()

	seedN := func(n int) {
		for i := n / 2; i > 0; i-- {
			MustPlaceGauge("some_arbitrary_test_unit" + strconv.Itoa(i)).Set(int64(i))
		}
		for i := n / 2; i > 0; i-- {
			MustPlaceCounter("some_arbitrary_count" + strconv.Itoa(i)).Add(uint64(i))
		}
	}

	req := httptest.NewRequest("GET", "/metrics", nil)
	resp := voidResponseWriter{}

	for _, n := range []int{0, 32, 1024, 32768} {
		seedN(n)

		b.Run(strconv.Itoa(n)+"-metrics", func(b *testing.B) {
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				HTTPHandler(resp, req)
			}
		})
	}
}
