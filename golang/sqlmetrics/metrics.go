package sqlmetrics

import (
	"time"

	"github.com/last9-cdk/proc"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	defaultLabels = []string{
		"per", proc.LabelHostname, "table", "dbname", "dbhost", "status",
		proc.LabelProgram, proc.LabelTenant, proc.LabelCluster,
	}

	sqlQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "sql_query_duration_milliseconds",
			Help:    "SQL duration per query",
			Buckets: proc.LatencyBins,
		},
		defaultLabels,
	)
)

func init() {
	// Metrics have to be registered to be exposed:
	prometheus.MustRegister(sqlQueryDuration)
}

// queryStatus is an enumeration.
type queryStatus int

const (
	success queryStatus = iota
	failure
)

// String method on queryStatus comes handy when printing or creating labels.
func (q queryStatus) String() string {
	return [...]string{"success", "failure"}[q]
}

func emitDuration(
	ls map[string]string, status queryStatus, start time.Time,
) error {
	labels := map[string]string{}
	for _, k := range defaultLabels {
		labels[k] = ""
	}

	labels[proc.LabelProgram] = proc.GetProgamName()
	labels[proc.LabelHostname] = proc.GetHostname()
	labels["status"] = status.String()

	for k, v := range ls {
		for _, l := range defaultLabels {
			if k == l && v != "" {
				labels[k] = v
			}
		}
	}

	sqlQueryDuration.With(labels).Observe(
		float64(time.Since(start).Milliseconds()),
	)

	return nil
}
