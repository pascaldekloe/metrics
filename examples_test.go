package metrics_test

import (
	"fmt"
	"net/http/httptest"

	"github.com/pascaldekloe/metrics"
)

func Example_labeled() {
	Thermostat := metrics.MustPlaceRealGauge("thermostat_celcius")
	Thermostat.Set(20)

	PerRoom := metrics.MustPlaceRealGauge1("thermostat_celcius", "room")
	PerRoom.Set(19, "kitchen")

	Station := metrics.MustPlaceRealGauge2("station_celcius", "city", "source")
	Station.Set(11.2, "Amsterdam (Schiphol)", "KNMI")
	Station.Set(9.6, "London", "BBC")

	// print samples
	metrics.SkipTimestamp = true
	rec := httptest.NewRecorder()
	metrics.HTTPHandler(rec, httptest.NewRequest("GET", "/metrics", nil))
	fmt.Print(rec.Body.String())

	// Output:
	// # TYPE thermostat_celcius gauge
	// thermostat_celcius 20
	// thermostat_celcius{room="kitchen"} 19
	// # TYPE station_celcius gauge
	// station_celcius{city="Amsterdam (Schiphol)",source="KNMI"} 11.2
	// station_celcius{city="London",source="BBC"} 9.6
}
