package metrics_test

import (
	"os"
	"time"

	"github.com/pascaldekloe/metrics"
)

// Basic Types
func Example() {
	Uptime := metrics.MustCounterSample("db_uptime_seconds", "")
	Uptime.Set(0, time.Now()) // set on ready

	Disk := metrics.MustRealSample("disk_usage_ratio", "Sectors of the total capacity.")
	Disk.Set(.47, time.Now()) // periodic OS call

	Size := metrics.MustCounter("db_response_bytes_total", "Raw size of the lookup.")
	Size.Add(812) // written amount

	Cache := metrics.MustInteger("db_cache_queries", "Number of query answers in cache.")
	Cache.Set(1000) // warm start
	Cache.Add(1)    // new entry
	Cache.Add(-900) // expiry sweep

	Backup := metrics.MustReal("db_backup_seconds", "Duration of the last backup.")
	Backup.Set(4.665)

	Delay := metrics.MustHistogram("db_delay_seconds", "Duration until response available.", 1e-6, 2e-6, 5e-6)
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
	// # TYPE db_uptime_seconds counter
	// db_uptime_seconds 0
	//
	// # TYPE disk_usage_ratio gauge
	// # HELP disk_usage_ratio Sectors of the total capacity.
	// disk_usage_ratio 0.47
	//
	// # TYPE db_response_bytes_total counter
	// # HELP db_response_bytes_total Raw size of the lookup.
	// db_response_bytes_total 812
	//
	// # TYPE db_cache_queries gauge
	// # HELP db_cache_queries Number of query answers in cache.
	// db_cache_queries 101
	//
	// # TYPE db_backup_seconds gauge
	// # HELP db_backup_seconds Duration of the last backup.
	// db_backup_seconds 4.665
	//
	// # TYPE db_delay_seconds histogram
	// # HELP db_delay_seconds Duration until response available.
	// db_delay_seconds_count 4
	// db_delay_seconds{le="1e-06"} 1
	// db_delay_seconds{le="2e-06"} 1
	// db_delay_seconds{le="5e-06"} 3
	// db_delay_seconds{le="+Inf"} 4
	// db_delay_seconds_sum 0.00058005654
}

// Label Combination
func Example_labels() {
	demo := metrics.NewRegister()
	Building := demo.Must2LabelInteger("health_hitpoints_total", "building", "ground")
	Unit := demo.Must3LabelInteger("health_hitpoints_total", "ground", "side", "unit")
	demo.MustHelp("health_hitpoints_total", "Damage Capacity")

	// launch
	Unit("Genisis Pit", "Nod", "Artilery").Set(300)
	Unit("Genisis Pit", "Nod", "Cyborg").Set(900)
	Building("Civilian Hospital", "Genisis Pit").Add(800)
	// attack
	Unit("Genisis Pit", "Nod", "Cyborg").Add(-596)
	Building("Civilian Hospital", "Genisis Pit").Add(-490)
	// tiberium
	Unit("Genisis Pit", "Nod", "Artilery").Add(-24)
	Unit("Genisis Pit", "Nod", "Cyborg").Add(110)

	metrics.SkipTimestamp = true
	demo.WriteText(os.Stdout)
	// Output:
	// # Prometheus Samples
	//
	// # TYPE health_hitpoints_total gauge
	// # HELP health_hitpoints_total Damage Capacity
	// health_hitpoints_total{building="Civilian Hospital",ground="Genisis Pit"} 310
	// health_hitpoints_total{ground="Genisis Pit",side="Nod",unit="Artilery"} 276
	// health_hitpoints_total{ground="Genisis Pit",side="Nod",unit="Cyborg"} 414
}

// Fixed Label Assignment
func Example_fixedLabels() {
	demo := metrics.NewRegister()
	measured := demo.Must2LabelRealSample("measured_celcius", "room", "source")
	setpoint := demo.Must1LabelReal("setpoint_celcius", "room")
	heating := demo.Must1LabelInteger("heating_watts", "room")
	heated := demo.Must1LabelCounterSample("radiator_joules_total", "room")
	cycles := demo.Must1LabelCounter("cycles_total", "room")

	rooms := [2]struct {
		Measured *metrics.Sample
		Setpoint *metrics.Real
		Heating  *metrics.Integer
		Heated   *metrics.Sample
		Cycles   *metrics.Counter
	}{}
	for i, label := range []string{"bedroom", "kitchen"} {
		rooms[i].Measured = measured(label, "thermostat")
		rooms[i].Setpoint = setpoint(label)
		rooms[i].Heating = heating(label)
		rooms[i].Heated = heated(label)
		rooms[i].Cycles = cycles(label)
	}

	rooms[0].Measured.Set(16.3, time.Date(2019, 2, 20, 17, 59, 46, 0, time.UTC))
	rooms[0].Setpoint.Set(19)
	rooms[0].Cycles.Add(1)
	rooms[0].Heated.Set(.27, time.Now())
	rooms[0].Heating.Set(1105)

	metrics.SkipTimestamp = true
	demo.WriteText(os.Stdout)
	// Output:
	// # Prometheus Samples
	//
	// # TYPE measured_celcius gauge
	// measured_celcius{room="bedroom",source="thermostat"} 16.3
	//
	// # TYPE setpoint_celcius gauge
	// setpoint_celcius{room="bedroom"} 19
	// setpoint_celcius{room="kitchen"} 0
	//
	// # TYPE heating_watts gauge
	// heating_watts{room="bedroom"} 1105
	// heating_watts{room="kitchen"} 0
	//
	// # TYPE radiator_joules_total counter
	// radiator_joules_total{room="bedroom"} 0.27
	//
	// # TYPE cycles_total counter
	// cycles_total{room="bedroom"} 1
	// cycles_total{room="kitchen"} 0
}

func ExampleMust1LabelHistogram() {
	demo := metrics.NewRegister()

	Duration := demo.Must1LabelHistogram("http_latency_seconds", "method", 0.001, 0.005, 0.01, 0.01)
	demo.MustHelp("http_latency_seconds", "Time from request initiation until response body retrieval.")

	Duration("GET").Add(0.0768753)
	Duration("OPTIONS").Add(0.0001414)
	Duration("GET").Add(0.0022779)
	Duration("GET").Add(0.0018714)
	Duration("GET").Add(0.0023789)

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

func ExampleMust2LabelHistogram() {
	demo := metrics.NewRegister()

	Duration := demo.Must2LabelHistogram("http_latency_seconds", "method", "status", 0.001, 0.005, 0.01, 0.01)
	demo.MustHelp("http_latency_seconds", "Time from request initiation until response body retrieval.")

	Duration("GET", "2xx").Add(0.0768753)
	Duration("GET", "3xx").Add(0.0001414)
	Duration("GET", "2xx").Add(0.0022779)
	Duration("GET", "2xx").Add(0.0018714)
	Duration("GET", "2xx").Add(0.0023789)

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
