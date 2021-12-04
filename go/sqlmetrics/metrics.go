package sqlmetrics

import (
	"time"

	"github.com/last9/last9-cdk/go/proc"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	subsystem     = "sql"
	defaultLabels = []string{
		"per", proc.LabelHostname, "table", "dbname", "dbhost", "status",
		proc.LabelProgram, proc.LabelTenant, proc.LabelCluster,
	}

	sqlQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: prometheus.BuildFQName(
				proc.Namespace,
				subsystem,
				"query_duration_milliseconds",
			),
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

func emitDuration(
	ls LabelSet, status queryStatus, start time.Time,
) error {
	labels := LabelSet{}
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

	sqlQueryDuration.With(labels.ToMap()).Observe(
		float64(time.Since(start).Milliseconds()),
	)

	return nil
}
