package sqlmetrics

import (
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	proxy "github.com/shogo82148/go-sql-proxy"
	"github.com/xo/dburl"
)

type connInfo struct {
	dsn       string
	dbName    string
	dbHost    string
	createdAt int
}

func (c *connInfo) LabelSet() LabelSet {
	if c == nil {
		return LabelSet{}
	}

	return LabelSet{
		"dbname": c.dbName,
		"dbhost": c.dbHost,
	}
}

var (
	connMap sync.Map
)

func storeConnInfo(conn *proxy.Conn, driverName, dsn string) error {
	info, err := parseDSN(driverName, dsn)
	if err != nil {
		return errors.Wrap(err, "parse dsn")
	}

	connMap.Store(conn, info)
	return nil
}

func loadConnInfo(conn *proxy.Conn) *connInfo {
	info, ok := connMap.Load(conn)
	if !ok {
		return nil
	}

	return info.(*connInfo)
}

// Note: the DB Url library expects a scheme in the DSN for
// all drivers, while the standard library `database/sql` expects some
// DSNs to be passed without a scheme (e.g. mysql) and some with
// a scheme (e.g. postgres).
// This function hence expects you to pass a driver name along with the DSN for
// successfully parsing the DSN.
func parseDSN(driver, dsn string) (*connInfo, error) {

	var SchemefullDSN string

	switch strings.Split(driver, ":")[0] {
	case "postgres":
		{
			SchemefullDSN = dsn
		}
	case "mysql":
		{
			//dsn = scheme + username + ":" + password + "@" + host + ":" + port + "/" + dbname
			SchemefullDSN = strings.Split(driver, ":")[0] + "://" + strings.Split(dsn, "@")[0] + "@" + strings.Split(strings.Split(dsn, "(")[1], ")")[0] + strings.Split(strings.Split(dsn, "(")[1], ")")[1]
		}
	}

	u, err := dburl.Parse(SchemefullDSN)
	if err != nil {
		return nil, errors.Wrap(err, "invalid dsn")
	}

	dbName := u.URL.Path
	if strings.HasPrefix(dbName, "/") {
		dbName = dbName[1:]
	} else if dbName == "" {
		dbName = u.URL.Opaque
	}

	dbHost := u.URL.Host
	if dbHost == "" {
		dbHost = "localhost"
	}

	return &connInfo{
		dsn:       dsn,
		dbName:    dbName,
		dbHost:    dbHost,
		createdAt: int(time.Now().Unix()),
	}, nil
}
