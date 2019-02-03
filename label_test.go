package metrics

import (
	"net/http/httptest"
	"testing"
)

func TestLabels(t *testing.T) {
	// mockup
	backup := appendTimeTail
	appendTimeTail = func(buf []byte) []byte {
		return append(buf, "1548759822954\n"...)
	}
	defer func() {
		appendTimeTail = backup

		reset()
	}()

	MustPlaceRealGauge("thermostat_celcius").Set(20)
	MustPlaceRealGauge1("thermostat_celcius", "room").Set(19, "kitchen")
	MustPlaceRealGauge2("weather_celcius", "city", "source").Set(11.2, "Amsterdam (Schiphol)", "KNMI")
	MustPlaceRealGauge2("weather_celcius", "city", "source").Set(9.6, "London", "BBC")
	MustPlaceRealGauge2("weather_celcius", "city", "source").Set(9.7, "London", "BBC")

	rec := httptest.NewRecorder()
	HTTPHandler(rec, httptest.NewRequest("GET", "/metrics", nil))

	const want = `# TYPE thermostat_celcius gauge
thermostat_celcius 20 1548759822954
thermostat_celcius{room="kitchen"} 19 1548759822954
# TYPE weather_celcius gauge
weather_celcius{city="Amsterdam (Schiphol)",source="KNMI"} 11.2 1548759822954
weather_celcius{city="London",source="BBC"} 9.7 1548759822954
`
	if got := rec.Body.String(); got != want {
		t.Errorf("got %q", got)
		t.Errorf("want %q", want)
	}
}
