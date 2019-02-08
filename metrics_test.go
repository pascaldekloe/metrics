package metrics

import (
	"mime"
	"net/http/httptest"
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

	g1 := MustNewGauge("g1")
	g1.Help("🆘")
	g1.Set(42)
	c1 := MustNewCounter("c1")
	c1.Add(1)
	c1.Add(8)
	c1.Help("override first 1")
	c1.Help("escape\n… and \\")

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
