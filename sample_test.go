package metrics_test

import (
	"net/http"
	"os"
	"time"

	"github.com/pascaldekloe/metrics"
	"github.com/pascaldekloe/metrics/gostat"
)

var (
	Started = time.Now()
	Uptime  = metrics.MustCounterSample("uptime_seconds", "Duration since start.")

	LogSize = metrics.MustRealSample("log_bytes", "Size reported by the filesystem.")
	LogIdle = metrics.MustRealSample("log_idle_seconds", "Duration since last change.")
)

func ExampleSample_lazy() {
	// mount exposition point
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		// update standard samples
		gostat.Capture()

		// update custom samples
		now := time.Now()
		Uptime.Set(float64(now.Sub(Started))/float64(time.Second), now)
		log, err := os.Stat("./mission.log")
		if err == nil {
			now = time.Now()
			LogSize.Set(float64(log.Size()), now)
			LogIdle.Set(float64(now.Sub(log.ModTime()))/float64(time.Second), now)
		}

		// serve serialized
		metrics.ServeHTTP(w, r)
	})
	// Output:
}
