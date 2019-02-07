package metrics_test

import (
	"fmt"
	"net/http/httptest"

	"github.com/pascaldekloe/metrics"
)

func Example() {
	Thermostat := metrics.MustPlaceGauge("thermostat_celcius")
	Thermostat.Set(20)

	Kitchen := metrics.Must1LabelGauge("thermostat_celcius", "room").Place("kitchen")
	Kitchen.Set(19)

	Station := metrics.Must2LabelGauge("station_celcius", "city", "source")
	Station.Place("Amsterdam (Schiphol)", "KNMI").Set(11.2)
	Station.Place("London", "BBC").Set(9.6)

	Delay := metrics.MustPlaceHistogram("db_delay_seconds", 1e-6, 2e-6, 5e-6)
	Delay.Add(0.00000391)
	Delay.Add(0.00000024054)
	Delay.Add(0.000002198)
	Delay.Add(0.000573708)

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
	// # TYPE db_delay_seconds histogram
	// db_delay_seconds{le="1e-06"} 1
	// db_delay_seconds{le="2e-06"} 1
	// db_delay_seconds{le="5e-06"} 3
	// db_delay_seconds{le="+Inf"} 4
	// db_delay_seconds_sum 0.00058005654
	// db_delay_seconds_count 4
}
