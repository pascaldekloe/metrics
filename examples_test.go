package metrics_test

import (
	"fmt"
	"net/http/httptest"

	"github.com/pascaldekloe/metrics"
)

func Example_labeled() {
	Thermostat := metrics.MustPlaceGauge("thermostat_celcius")
	Thermostat.Set(20)

	Kitchen := metrics.Must1LabelGauge("thermostat_celcius", "room").Place("kitchen")
	Kitchen.Set(19)

	Station := metrics.Must2LabelGauge("station_celcius", "city", "source")
	Station.Place("Amsterdam (Schiphol)", "KNMI").Set(11.2)
	Station.Place("London", "BBC").Set(9.6)

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
