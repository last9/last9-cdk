package sqlmetrics

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/blastrain/vitess-sqlparser/sqlparser"
	pg_query "github.com/pganalyze/pg_query_go"
	"github.com/tidwall/gjson"
)

type LabelSet map[string]string

// Merge accepts a set of LabelSets and merge them into this LabelSet.
func (l LabelSet) Merge(ls ...LabelSet) LabelSet {
	for _, m := range ls {
		for k, v := range m {
			if v != "" {
				l[k] = v
			}
		}
	}

	return l
}

// ToMap converts a LabelSet to map[string]string
func (l LabelSet) ToMap() map[string]string {
	return map[string]string(l)
}

// A type that user can extend to Parse a query and extract less verbose
// or more relevant labels out of it.
type LabelMaker func(string, string) LabelSet

const idealLabelLen = 20

// The default labelSet to be exported is just a query, that too trimmed down
// to 140 charachters only. Queries can be large and can really bring down
// the metric server to its knees if left untapped. If this behaviour is not
// desired, a user can anwyay implement their own QToLabelSet and emit the
// raw query as-it-is.
func defaultLabelMaker(driver, q string) LabelSet {
	statementName := getStatementName(driver, q)

	if len(q) > idealLabelLen {
		q = q[:idealLabelLen] + "..."
	}

	return LabelSet{"per": q, "statement": statementName}
}

// queryStatus is an enumeration.
type queryStatus int

const (
	success queryStatus = iota
	failure
)

// String method on queryStatus comes handy when printing or
// creating labels.
func (q queryStatus) String() string {
	return [...]string{"success", "failure"}[q]
}

// Accepts a driver name and query and returns the statement name of the first level query.
func getStatementName(driver, q string) string {

	switch strings.Split(driver, ":")[0] {
	case "postgres":
		{
			// Parse query and get query AST
			tree, err := pg_query.ParseToJSON(q)
			if err != nil {
				log.Println("Error parsing query:", q, "Error occured :", err)
				return ""
			}

			// Get the statement name of the first level query
			statement := gjson.Get(tree[1:len(tree)-1], "..0.RawStmt.stmt.@keys.0")

			return statement.String()
		}
	case "mysql":
		{
			// Parse query and get query AST
			stmt, err := sqlparser.Parse(q)
			if err != nil {
				log.Println("Error parsing query:", q, "Error :", err)
				return ""
			}

			// convert AST to json string
			stmtJson, err := json.Marshal(stmt)
			if err != nil {
				log.Println("Error in conversion to json :", q, "Error :", err)
				return ""
			}

			// Parse the json and get the statement name of the first level query
			if gjson.Get(string(stmtJson), "SelectExprs").String() != "" {
				return "select"
			}
			return gjson.Get(string(stmtJson), "Action").String()
		}
	}

	return ""
}
