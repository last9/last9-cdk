## Sqlmetrics Unit tests
```bash
export LAST9_PQSQL_DSN="postgres://testuser:testpassword@localhost:5432/last9"
export LAST9_MYSQL_DSN="mysql:password@tcp(localhost:8306)/last9"
go test
```