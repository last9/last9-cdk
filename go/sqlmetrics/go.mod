module github.com/last9-go-cdk/sqlmetrics

go 1.17

require (
	github.com/last9-cdk/proc v0.0.0-00010101000000-000000000000
	github.com/last9-cdk/tests v0.0.0-00010101000000-000000000000
	github.com/lib/pq v1.10.4
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/prometheus/client_model v0.2.0
	github.com/shogo82148/go-sql-proxy v0.6.1
	github.com/xo/dburl v0.9.0
	gotest.tools v2.2.0+incompatible
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/google/go-cmp v0.5.5 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/prometheus/common v0.32.1 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	golang.org/x/sys v0.0.0-20210603081109-ebe580a85c40 // indirect
	google.golang.org/protobuf v1.26.0-rc.1 // indirect
)

replace github.com/last9-cdk/proc => ../proc

replace github.com/last9-cdk/tests => ../tests
