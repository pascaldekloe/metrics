package metrics_test

import (
	"os"
	"time"

	"github.com/pascaldekloe/metrics"
)

// Metric Types
func Example() {
	// totals with natural numbers
	RespBytes := metrics.MustCounter("db_response_bytes_total", "Raw size of the lookup.")
	// gauge with integer numbers
	CacheCount := metrics.MustInteger("db_cache_queries", "Number of query answers in cache.")
	// double precision floating points
	BackupPriority := metrics.MustReal("db_backup_priority", "Sentiment for data redundancy.")
	// count in steps of ≤ 1 µs, ≤ 2 µs, ≤ 5 µs and > 5 µs
	DelaySeconds := metrics.MustHistogram("db_delay_seconds", "Duration until response available.", 1e-6, 2e-6, 5e-6)
	// samples for periodic updates
	UptimeSeconds := metrics.MustCounterSample("db_uptime_seconds", "Duration since launch.")
	DiskUsage := metrics.MustRealSample("db_disk_usage_ratio", "Sectors of the total capacity.")

	// measures
	BackupPriority.Set(7.3)
	CacheCount.Set(1000)
	UptimeSeconds.Set(0, time.Now())
	DelaySeconds.Add(0.00000391)
	DelaySeconds.Add(0.00000024054)
	DelaySeconds.Add(0.000002198)
	DelaySeconds.Add(0.000573708)
	CacheCount.Add(1)
	RespBytes.Add(812)
	DiskUsage.Set(.47, time.Now())
	CacheCount.Add(-997)
	UptimeSeconds.Set(5, time.Now())

	// print
	metrics.SkipTimestamp = true
	metrics.WriteTo(os.Stdout)
	// Output:
	// # Prometheus Samples
	//
	// # TYPE db_response_bytes_total counter
	// # HELP db_response_bytes_total Raw size of the lookup.
	// db_response_bytes_total 812
	//
	// # TYPE db_cache_queries gauge
	// # HELP db_cache_queries Number of query answers in cache.
	// db_cache_queries 4
	//
	// # TYPE db_backup_priority gauge
	// # HELP db_backup_priority Sentiment for data redundancy.
	// db_backup_priority 7.3
	//
	// # TYPE db_delay_seconds histogram
	// # HELP db_delay_seconds Duration until response available.
	// db_delay_seconds_count 4
	// db_delay_seconds{le="1e-06"} 1
	// db_delay_seconds{le="2e-06"} 1
	// db_delay_seconds{le="5e-06"} 3
	// db_delay_seconds{le="+Inf"} 4
	// db_delay_seconds_sum 0.00058005654
	//
	// # TYPE db_uptime_seconds counter
	// # HELP db_uptime_seconds Duration since launch.
	// db_uptime_seconds 5
	//
	// # TYPE db_disk_usage_ratio gauge
	// # HELP db_disk_usage_ratio Sectors of the total capacity.
	// db_disk_usage_ratio 0.47
}

// Label Combination
func Example_labels() {
	// setup
	demo := metrics.NewRegister()
	Building := demo.Must2LabelInteger("hitpoints_total", "ground", "building")
	Arsenal := demo.Must3LabelInteger("hitpoints_total", "ground", "arsenal", "side")
	demo.MustHelp("hitpoints_total", "Damage Capacity")

	// measures
	Building("Genesis Pit", "Civilian Hospital").Set(800)
	Arsenal("Genesis Pit", "Tech Center", "Nod").Set(500)
	Arsenal("Genesis Pit", "Cyborg", "Nod").Set(900)
	Arsenal("Genesis Pit", "Cyborg", "Nod").Add(-596)
	Building("Genesis Pit", "Civilian Hospital").Add(-490)
	Arsenal("Genesis Pit", "Cyborg", "Nod").Add(110)

	// print
	metrics.SkipTimestamp = true
	demo.WriteTo(os.Stdout)
	// Output:
	// # Prometheus Samples
	//
	// # TYPE hitpoints_total gauge
	// # HELP hitpoints_total Damage Capacity
	// hitpoints_total{building="Civilian Hospital",ground="Genesis Pit"} 310
	// hitpoints_total{arsenal="Tech Center",ground="Genesis Pit",side="Nod"} 500
	// hitpoints_total{arsenal="Cyborg",ground="Genesis Pit",side="Nod"} 414
}

// Fixed Assignment & Default Values
func Example_labelsFix() {
	// setup
	demo := metrics.NewRegister()
	measured := demo.Must2LabelRealSample("measured_celcius", "room", "source")
	setpoint := demo.Must1LabelReal("setpoint_celcius", "room")
	heating := demo.Must1LabelInteger("heating_watts", "room")
	heated := demo.Must1LabelCounterSample("radiator_joules_total", "room")
	cycles := demo.Must1LabelCounter("cycles_total", "room")

	// label composition
	roomNames := [...]string{"bedroom", "kitchen"}
	rooms := [len(roomNames)]struct {
		Measured *metrics.Sample
		Setpoint *metrics.Real
		Heating  *metrics.Integer
		Heated   *metrics.Sample
		Cycles   *metrics.Counter
	}{}
	for i, name := range roomNames {
		rooms[i].Measured = measured(name, "thermostat")
		rooms[i].Setpoint = setpoint(name)
		rooms[i].Heating = heating(name)
		rooms[i].Heated = heated(name)
		rooms[i].Cycles = cycles(name)
	}

	// measures
	rooms[0].Measured.Set(16.3, time.Date(2019, 2, 20, 17, 59, 46, 0, time.UTC))
	rooms[0].Setpoint.Set(19)
	rooms[0].Cycles.Add(1)
	rooms[0].Heated.Set(.27, time.Now())
	rooms[0].Heating.Set(1105)

	// print
	metrics.SkipTimestamp = true
	demo.WriteTo(os.Stdout)
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
	// setup
	demo := metrics.NewRegister()
	Duration := demo.Must1LabelHistogram("http_latency_seconds", "method", 0.001, 0.005, 0.01, 0.01)
	demo.MustHelp("http_latency_seconds", "Time from request initiation until response body retrieval.")

	// measures
	Duration("GET").Add(0.0768753)
	Duration("OPTIONS").Add(0.0001414)
	Duration("GET").Add(0.0022779)
	Duration("GET").Add(0.0018714)
	Duration("GET").Add(0.0023789)

	// print
	metrics.SkipTimestamp = true
	demo.WriteTo(os.Stdout)
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
	// setup
	demo := metrics.NewRegister()
	Duration := demo.Must2LabelHistogram("http_latency_seconds", "method", "status", 0.001, 0.005, 0.01, 0.01)
	demo.MustHelp("http_latency_seconds", "Time from request initiation until response body retrieval.")

	// measures
	Duration("GET", "2xx").Add(0.0768753)
	Duration("GET", "3xx").Add(0.0001414)
	Duration("GET", "2xx").Add(0.0022779)
	Duration("GET", "2xx").Add(0.0018714)
	Duration("GET", "2xx").Add(0.0023789)

	// print
	metrics.SkipTimestamp = true
	demo.WriteTo(os.Stdout)
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
