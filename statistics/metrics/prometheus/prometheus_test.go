package prometheus

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"time"
)

var (
	cpuTemp = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cpu_temperature_celsius",
		Help: "Current temperature of the CPU.",
	})
	hdFailures = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hd_errors_total",
		Help: "Number of hard-disk errors.",
	})
)

func init() {
	prometheus.MustRegister(cpuTemp)
	prometheus.MustRegister(hdFailures)
}

func main() {
	go func() {
		for {
			cpuTemp.Set(65.3)
			hdFailures.Inc()

			time.Sleep(1 * time.Second)
		}
	}()
	//hdFailures.Add(1)

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8080", nil)
}
