package sqlmetrics

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	proxy "github.com/shogo82148/go-sql-proxy"
)

// enabled is a local registry that keeps a track of what has been enabled
// and what hasn't been to avoid a re-initialization.
var enabled sync.Map

// driverName accepts the driver Identifier like :postgres, :mysql
// and returns the last9 version of it, Which can be referred to later.
// There is no magic in this function. The actual magic happens in RegisterDB
func driverName(d string) string {
	if strings.HasSuffix(d, "last9") {
		return d
	}

	return d + ":last9"
}

// isDriverEnabled looks for the driver string if it has been registered
// or enabled already. Remember everrytime you write
// import _ "github.com/lib/pq"
// it registers postgrs as a legit driver.
func isDriverEnabled(d string) bool {
	for _, driver := range sql.Drivers() {
		if driver == d {
			return true
		}
	}

	return false
}

// dbCtx is a struct that is passed around from Pre-* hook to Post-* hook
// The name is a misnomer but that's how the package owner had named it.
// The attributes are self-explanatory and carry their own comments.
type dbCtx struct {
	start time.Time // time at which the pre-hook was executed.
	query string    // the raw query string that was called.
	// labels map[string]string //labelSet for this context
}

func getQueryStatus(err error) queryStatus {
	if err != nil {
		return failure
	}

	return success
}

func RegisterDB(d string) (string, error) {
	return RegisterDBWithLabelMaker(d, defaultLabelMaker)
}

// TODO Implement an Unregister as well. Since the fn will be locked forever
// But since the method could be called from goroutines, it will need some
// sync mechanics. But until we get to that, we can just skip it until then.

func RegisterDBWithLabelMaker(d string, fn LabelMaker) (string, error) {
	// If this is an already registered driver, don't do anything.
	// SQL will take care of the registered driver for that database.
	if x, ok := enabled.Load(d); ok {
		return x.(string), nil
	}

	if !isDriverEnabled(d) {
		return "", errors.Errorf("%v has not been activated. Import it please", d)
	}

	// Just perform a blank open to extract the Driver() out of it.
	// Since the DSN is empty, this is a harmless operation and does not
	// leave behind an actual open connection.
	db, err := sql.Open(d, "")
	if err != nil {
		return "", errors.Wrapf(err, "init %v", d)
	}

	name := driverName(d)

	sql.Register(name, proxy.NewProxyContext(db.Driver(), &proxy.HooksContext{
		PreExec: func(
			c context.Context, stmt *proxy.Stmt, args []driver.NamedValue,
		) (interface{}, error) {
			return &dbCtx{start: time.Now(), query: stmt.QueryString}, nil
		},

		PostExec: func(
			c context.Context, ctx interface{}, stmt *proxy.Stmt,
			args []driver.NamedValue, result driver.Result, err error,
		) error {
			if ctx == nil {
				return nil
			}

			dc := ctx.(*dbCtx)
			if err := emitDuration(
				fn(dc.query), getQueryStatus(err), dc.start,
			); err != nil {
				log.Printf("%+v", err)
			}
			return nil
		},

		PreQuery: func(
			c context.Context, stmt *proxy.Stmt, args []driver.NamedValue,
		) (interface{}, error) {
			return &dbCtx{start: time.Now(), query: stmt.QueryString}, nil
		},

		PostQuery: func(
			c context.Context, ctx interface{}, stmt *proxy.Stmt,
			args []driver.NamedValue, rows driver.Rows, err error,
		) error {
			if ctx == nil {
				return nil
			}

			dc := ctx.(*dbCtx)
			if err := emitDuration(
				fn(dc.query), getQueryStatus(err), dc.start,
			); err != nil {
				log.Printf("%+v", err)
			}

			return nil
		},

		PreBegin: func(
			c context.Context, conn *proxy.Conn,
		) (interface{}, error) {
			return nil, nil
		},

		PostCommit: func(
			c context.Context, ctx interface{}, tx *proxy.Tx, err error,
		) error {
			return nil
		},

		PostRollback: func(
			c context.Context, ctx interface{}, tx *proxy.Tx, err error,
		) error {
			return nil
		},
	}))

	// mark this driver as enabled.
	enabled.Store(d, name)
	return name, nil
}
