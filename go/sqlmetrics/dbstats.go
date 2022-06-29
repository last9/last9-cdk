package sqlmetrics

import (
	"database/sql"
	"time"

	"github.com/last9/last9-cdk/go/proc"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	dbLabels = []string{
		proc.LabelHostname, "dbname", "dbhost",
		proc.LabelProgram, proc.LabelTenant, proc.LabelCluster,
	}

	maxOpenConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: prometheus.BuildFQName(
				proc.Namespace,
				subsystem,
				"connections_max_open",
			),
			Help: "Maximum number of open connections to the database.",
		},
		dbLabels,
	)

	connectionsInUse = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: prometheus.BuildFQName(
				proc.Namespace,
				subsystem,
				"connections_in_use",
			),
			Help: "The number of connections currently in use.",
		},
		dbLabels,
	)

	connectionsIdle = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: prometheus.BuildFQName(
				proc.Namespace,
				subsystem,
				"connections_idle",
			),
			Help: "The number of idle connections.",
		},
		dbLabels,
	)

	// WaitCount is a counter of total no. of connections waited for.
	waitCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: prometheus.BuildFQName(
				proc.Namespace,
				subsystem,
				"connections_wait_total",
			),
			Help: "The total number of connections waited for",
		},
		dbLabels,
	)

	waitDuration = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: prometheus.BuildFQName(
				proc.Namespace,
				subsystem,
				"connections_wait_duration_total",
			),
			Help: "The total time blocked waiting for a new connection.",
		},
		dbLabels,
	)
)

func init() {
	prometheus.MustRegister(maxOpenConnections)
	prometheus.MustRegister(connectionsInUse)
	prometheus.MustRegister(connectionsIdle)
	prometheus.MustRegister(waitCount)
	prometheus.MustRegister(waitDuration)
}

// emitStats accepts a DBStats struct and sets or increments all the
// appropriate metrics.
func emitStats(s sql.DBStats, labels LabelSet) {
	ls := labels.ToMap()
	maxOpenConnections.With(ls).Set(float64(s.MaxOpenConnections))
	connectionsInUse.With(ls).Set(float64(s.InUse))
	connectionsIdle.With(ls).Set(float64(s.Idle))
	waitCount.With(ls).Add(float64(s.WaitCount))
	waitDuration.With(ls).Add(float64(s.WaitDuration))
}

// make Labels to be used for DB Stats
func makeStatsLabels(driver, dsn string) (LabelSet, error) {
	info, err := parseDSN(driver, dsn)
	if err != nil {
		return nil, errors.Wrap(err, "parse dsn emit db")
	}

	labels := LabelSet{}
	for _, k := range dbLabels {
		labels[k] = ""
	}

	labels[proc.LabelProgram] = proc.GetProgamName()
	labels[proc.LabelHostname] = proc.GetHostname()

	return labels.Merge(info.LabelSet()), nil
}

// EmitDBStats accepts a Database connection and starts a per-minute tick
// to emit the gauges and counters corresponding the connections spawned
// by this binary. It's fairly light weight with minimal allocation so
// performance should not really be a concern here.
func EmitDBStats(db *sql.DB, driver, dsn string) error {
	l, err := makeStatsLabels(driver, dsn)
	if err != nil {
		return errors.Wrap(err, "stats labels")
	}

	ticker := time.NewTicker(60 * time.Second)
	for range ticker.C {
		emitStats(db.Stats(), l)
	}

	return nil
}
