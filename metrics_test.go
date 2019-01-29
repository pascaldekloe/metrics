package metrics

import (
	"mime"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func init() {
	// mockup
	nowMilli = func() int64 { return 1548759822954 }
}

func TestSerialize(t *testing.T) {
	// cleanup
	defer func() {
		registry = sync.Map{}
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

	const wantGauge = `# HELP 🆘
# TYPE gauge
g1 42 1548759822954
`
	const wantCounter = `# HELP escape\n… and \\
# TYPE counter
c1 9 1548759822954
`

	switch got := rec.Body.String(); got {
	case wantGauge + wantCounter, wantCounter + wantGauge:
		break
	default:
		t.Errorf("got body  %q", got)
		t.Errorf("want %q and %q", wantGauge, wantCounter)
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
