package metrics

import (
	"bytes"
	"io"
	"math"
	"mime"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func reset() {
	SkipTimestamp = false

	indices = make(map[string]uint32)
	metrics = nil
}

func TestHelp(t *testing.T) {
	defer reset()

	MustNewGauge("g").Help("set on gauge")
	Must1LabelGauge("l1g", "l").Help("set on map")
	Must2LabelGauge("l2g", "l1", "l2").With("v1", "v2").Help("set on labeled gauge")
	Must3LabelGauge("l3g", "l1", "l2", "l3").With("v1", "v2", "v3").Help("set on labeled gauge to override")
	Must3LabelGauge("l3g", "l4", "l5", "l6").With("v4", "v5", "v6").Help("override on labeled gauge")

	want := map[string]string{
		"g":   "set on gauge",
		"l1g": "set on map",
		"l2g": "set on labeled gauge",
		"l3g": "override on labeled gauge",
	}

	var buf bytes.Buffer
	WriteText(&buf)

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

func TestNewHistogramBuckets(t *testing.T) {
	defer reset()

	var golden = []struct {
		feed []float64
		want []float64
	}{
		{[]float64{4, 1, 2}, []float64{1, 2, 4}},
		{[]float64{8, math.Inf(1)}, []float64{8}},
		{[]float64{math.Inf(-1), 8}, []float64{8}},
		{[]float64{math.NaN(), 7, math.Inf(1), 3}, []float64{3, 7}},
	}

	for i, gold := range golden {
		h := MustNewHistogram("h"+strconv.Itoa(i), gold.feed...)
		if !reflect.DeepEqual(h.bucketBounds, gold.want) {
			t.Errorf("%v: got buckets %v, want %v", gold.feed, h.bucketBounds, gold.want)
		}
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
