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

func storeConnInfo(conn *proxy.Conn, dsn string) error {
	info, err := parseDSN(dsn)
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

func parseDSN(dsn string) (*connInfo, error) {
	u, err := dburl.Parse(dsn)
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
