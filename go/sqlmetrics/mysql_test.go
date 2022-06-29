package sqlmetrics

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/last9/last9-cdk/go/proc"
	"github.com/last9/last9-cdk/go/tests"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
	"gotest.tools/assert"
)

func getMySqlDSN() string {
	dsn := os.Getenv("LAST9_MYSQL_DSN")
	if dsn == "" {
		panic("LAST9_MYSQL_DSN is not set")
	}

	return dsn
}

func getMySqlDB(driverName string) (*sql.DB, error) {
	dsn := getMySqlDSN()
	// NOTE: This is the Second change that you make.
	// whatever the regsitered driver was, use <driver>:last9 to connect
	// instead of <driver>. And that's it.
	return sql.Open(driverName+":last9", dsn)
}

func TestMySql(t *testing.T) {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	srv := tests.MakeServer(mux)
	defer srv.Close()

	driverName := "mysql"
	t.Run("register cannot return error", func(t *testing.T) {
		// NOTE: This is the first change that you do.
		// Declare the driver that you would want to use.
		name, err := RegisterDriver(Options{Driver: driverName})
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "mysql:last9", name)
	})

	t.Run("connect to db", func(t *testing.T) {
		dsn := os.Getenv("LAST9_MYSQL_DSN")
		if dsn == "" {
			t.Fatal("LAST9_MYSQL_DSN is not set")
		}

		db, err := sql.Open("mysql", dsn)
		if err != nil {
			t.Fatal(err)
		}

		defer db.Close()

		rows, err := db.Query("SELECT 'Hello'")
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, err, nil)

		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, "Hello", name)
		}

		assert.Equal(t, nil, rows.Close())
	})

	t.Run("connect via proxy", func(t *testing.T) {
		db, err := getMySqlDB(driverName)

		if err != nil {
			t.Fatal(err)
		}

		defer db.Close()

		rows, err := db.Query("SELECT 'Hello'")
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, err, nil)

		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, "Hello", name)
		}

		assert.Equal(t, nil, rows.Close())
	})

	t.Run("create and execute", func(t *testing.T) {
		resetMetrics()

		ctx := context.Background()
		db, err := getMySqlDB(driverName)
		if err != nil {
			t.Fatal(err)
		}

		defer db.Close()

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}

		if _, err := tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS pets (name VARCHAR(20) UNIQUE, owner VARCHAR(20),
		species VARCHAR(20), sex CHAR(1), birth DATE, death DATE);
		`); err != nil {
			t.Fatal(err)
		}

		// Here, the query is executed on the transaction instance, and
		// not applied to the database yet
		if _, err = tx.ExecContext(ctx, `INSERT IGNORE INTO pets (name, species)
		VALUES ('Fido', 'dog'), ('Albert', 'cat');
		`); err != nil {
			// Incase we find any error in the query execution, rollback
			tx.Rollback()
			return
		}

		var catCount int
		// Run a query to get a count of all cats
		row := tx.QueryRow(`SELECT count(*) FROM pets WHERE species='cat';`)
		// Store the count in the `catCount` variable
		if err = row.Scan(&catCount); err != nil {
			tx.Rollback()
			t.Fatal(err)
			return
		}

		assert.Equal(t, 1, catCount)

		// Finally, if no errors are recieved from the queries,
		// commit the transaction this applies the above changes to our database
		if err = tx.Commit(); err != nil {
			tx.Rollback()
			t.Fatal(err)
		}

		o, err := tests.GetMetrics(srv.URL)
		if err != nil {
			t.Fatal(err)
		}

		req := o[expectedMetric]

		assert.Equal(t, req.GetType(), dto.MetricType_HISTOGRAM)
		for _, m := range req.GetMetric() {
			h := m.GetHistogram()
			// log.Println("Labels : ", m.GetLabel())

			assert.Equal(t, 1, int(h.GetSampleCount()))
			assert.Equal(t, true, labelSetContains(
				m.GetLabel(), map[string]string{
					"cluster":  "",
					"tenant":   "",
					"dbname":   "last9",
					"hostname": proc.GetHostname(),
					"program":  proc.GetProgamName(),
					"status":   "success",
					"table":    "",
				}),
			)
		}
	})

	t.Run("conn metrics", func(t *testing.T) {
		dsn := getMySqlDSN()
		labels, err := makeStatsLabels(driverName, dsn)
		if err != nil {
			t.Fatal(err)
		}

		db, err := getMySqlDB(driverName)
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()

		emitStats(db.Stats(), labels)

		o, err := tests.GetMetrics(srv.URL)
		if err != nil {
			t.Fatal(err)
		}

		idle, ok := o["last9_sql_connections_idle"]
		assert.Equal(t, true, ok)
		assert.Equal(t, idle.GetType(), dto.MetricType_GAUGE)

		used, ok := o["last9_sql_connections_in_use"]
		assert.Equal(t, true, ok)
		assert.Equal(t, used.GetType(), dto.MetricType_GAUGE)

		max, ok := o["last9_sql_connections_max_open"]
		assert.Equal(t, true, ok)
		assert.Equal(t, max.GetType(), dto.MetricType_GAUGE)

		wait_duration, ok := o["last9_sql_connections_wait_duration_total"]
		assert.Equal(t, true, ok)
		assert.Equal(t, wait_duration.GetType(), dto.MetricType_COUNTER)

		wait_total, ok := o["last9_sql_connections_wait_total"]
		assert.Equal(t, true, ok)
		assert.Equal(t, wait_total.GetType(), dto.MetricType_COUNTER)
	})

}
