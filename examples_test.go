package metrics_test

import (
	"os"
	"time"

	"github.com/pascaldekloe/metrics"
)

func Example() {
	Thermostat := metrics.MustNewGauge("thermostat_celcius")
	Thermostat.Set(20)

	Kitchen := metrics.MustNew1LabelGauge("thermostat_celcius", "room").With("kitchen")
	Kitchen.Set(19)

	Station := metrics.MustNew2LabelGauge("station_celcius", "city", "source")
	Station.With("Amsterdam (Schiphol)", "KNMI").Set(11.2)
	Station.With("London", "BBC").Set(9.6)

	Delay := metrics.MustNewHistogram("db_delay_seconds", 1e-6, 2e-6, 5e-6)
	Delay.Add(0.00000391)
	Delay.Add(0.00000024054)
	Delay.Add(0.000002198)
	Delay.Add(0.000573708)

	// print samples
	metrics.SkipTimestamp = true
	metrics.WriteText(os.Stdout)
	// Output:
	// # Prometheus Samples
	//
	// # TYPE thermostat_celcius gauge
	// thermostat_celcius 20
	// thermostat_celcius{room="kitchen"} 19
	//
	// # TYPE station_celcius gauge
	// station_celcius{city="Amsterdam (Schiphol)",source="KNMI"} 11.2
	// station_celcius{city="London",source="BBC"} 9.6
	//
	// # TYPE db_delay_seconds histogram
	// db_delay_seconds_count 4
	// db_delay_seconds{le="1e-06"} 1
	// db_delay_seconds{le="2e-06"} 1
	// db_delay_seconds{le="5e-06"} 3
	// db_delay_seconds{le="+Inf"} 4
	// db_delay_seconds_sum 0.00058005654
}

// Label Assignment And Default Value Initiation
func Example_labels() {
	demo := metrics.NewRegister()
	measured := demo.MustNew1LabelGaugeSample("measured_celcius", "room")
	setpoint := demo.MustNew1LabelGauge("setpoint_celcius", "room")
	cycles := demo.MustNew1LabelCounter("cycles_total", "room")
	heating := demo.MustNew1LabelCounterSample("heating_joules_total", "room")

	rooms := []*struct {
		Measured *metrics.Sample
		Setpoint *metrics.Gauge
		Cycles   *metrics.Counter
		Heating  *metrics.Sample
	}{
		{
			Measured: measured.With("bedroom"),
			Setpoint: setpoint.With("bedroom"),
			Cycles:   cycles.With("bedroom"),
			Heating:  heating.With("bedroom"),
		}, {
			Measured: measured.With("kitchen"),
			Setpoint: setpoint.With("kitchen"),
			Cycles:   cycles.With("kitchen"),
			Heating:  heating.With("kitchen"),
		},
	}

	rooms[0].Measured.Set(16.3, time.Date(2019, 2, 20, 17, 59, 46, 0, time.UTC))
	rooms[0].Setpoint.Set(19)
	rooms[0].Cycles.Add(1)
	rooms[0].Heating.Set(1105, time.Date(2019, 2, 20, 17, 59, 47, 0, time.UTC))

	metrics.SkipTimestamp = true
	demo.WriteText(os.Stdout)
	// Output:
	// # Prometheus Samples
	//
	// # TYPE measured_celcius gauge
	// measured_celcius{room="bedroom"} 16.3
	//
	// # TYPE setpoint_celcius gauge
	// setpoint_celcius{room="bedroom"} 19
	// setpoint_celcius{room="kitchen"} 0
	//
	// # TYPE cycles_total counter
	// cycles_total{room="bedroom"} 1
	// cycles_total{room="kitchen"} 0
	//
	// # TYPE heating_joules_total counter
	// heating_joules_total{room="bedroom"} 1105
}

func ExampleMap1LabelHistogram() {
	demo := metrics.NewRegister()

	Duration := demo.MustNew1LabelHistogram("http_latency_seconds", "method", 0.001, 0.005, 0.01, 0.01)
	demo.MustHelp("http_latency_seconds", "Time from request initiation until response body retrieval.")

	Duration.With("GET").Add(0.0768753)
	Duration.With("OPTIONS").Add(0.0001414)
	Duration.With("GET").Add(0.0022779)
	Duration.With("GET").Add(0.0018714)
	Duration.With("GET").Add(0.0023789)

	metrics.SkipTimestamp = true
	demo.WriteText(os.Stdout)
	// Output:
	// # Prometheus Samples
	//
	// # TYPE http_latency_seconds histogram
	// # HELP http_latency_seconds Time from request initiation until response body retrieval.
	// http_latency_seconds_count{method="GET"} 4
	// http_latency_seconds{le="0.001",method="GET"} 0
	// http_latency_seconds{le="0.005",method="GET"} 3
	// http_latency_seconds{le="0.01",method="GET"} 3
	// http_latency_seconds{le="+Inf",method="GET"} 4
	// http_latency_seconds_sum{method="GET"} 0.08340349999999999
	// http_latency_seconds_count{method="OPTIONS"} 1
	// http_latency_seconds{le="0.001",method="OPTIONS"} 1
	// http_latency_seconds{le="0.005",method="OPTIONS"} 1
	// http_latency_seconds{le="0.01",method="OPTIONS"} 1
	// http_latency_seconds{le="+Inf",method="OPTIONS"} 1
	// http_latency_seconds_sum{method="OPTIONS"} 0.0001414
}

func ExampleMap2LabelHistogram() {
	demo := metrics.NewRegister()

	Duration := demo.MustNew2LabelHistogram("http_latency_seconds", "method", "status", 0.001, 0.005, 0.01, 0.01)
	demo.MustHelp("http_latency_seconds", "Time from request initiation until response body retrieval.")

	Duration.With("GET", "2xx").Add(0.0768753)
	Duration.With("GET", "3xx").Add(0.0001414)
	Duration.With("GET", "2xx").Add(0.0022779)
	Duration.With("GET", "2xx").Add(0.0018714)
	Duration.With("GET", "2xx").Add(0.0023789)

	metrics.SkipTimestamp = true
	demo.WriteText(os.Stdout)
	// Output:
	// # Prometheus Samples
	//
	// # TYPE http_latency_seconds histogram
	// # HELP http_latency_seconds Time from request initiation until response body retrieval.
	// http_latency_seconds_count{method="GET",status="2xx"} 4
	// http_latency_seconds{le="0.001",method="GET",status="2xx"} 0
	// http_latency_seconds{le="0.005",method="GET",status="2xx"} 3
	// http_latency_seconds{le="0.01",method="GET",status="2xx"} 3
	// http_latency_seconds{le="+Inf",method="GET",status="2xx"} 4
	// http_latency_seconds_sum{method="GET",status="2xx"} 0.08340349999999999
	// http_latency_seconds_count{method="GET",status="3xx"} 1
	// http_latency_seconds{le="0.001",method="GET",status="3xx"} 1
	// http_latency_seconds{le="0.005",method="GET",status="3xx"} 1
	// http_latency_seconds{le="0.01",method="GET",status="3xx"} 1
	// http_latency_seconds{le="+Inf",method="GET",status="3xx"} 1
	// http_latency_seconds_sum{method="GET",status="3xx"} 0.0001414
}
