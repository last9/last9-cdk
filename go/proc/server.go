package proc

import (
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ServeMetrics exposes whatever prometheus metrics are, on specified Port
func ServeMetrics(port int) {
	log.Println("Serving metrics on", port)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
