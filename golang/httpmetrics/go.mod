module github.com/last9/last9-cdk/golang/httpmetrics

go 1.17

require (
	github.com/gorilla/mux v1.8.0
	github.com/last9-cdk/proc v0.0.0-00010101000000-000000000000
	github.com/last9-cdk/tests v0.0.0-00010101000000-000000000000
	github.com/last9/pat v0.0.0-20211111093525-daacb495b5a9
	github.com/prometheus/client_golang v1.11.0
	github.com/prometheus/client_model v0.2.0
	gopkg.in/go-playground/assert.v1 v1.2.1
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/common v0.32.1 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	golang.org/x/sys v0.0.0-20210603081109-ebe580a85c40 // indirect
	google.golang.org/protobuf v1.26.0-rc.1 // indirect
)

replace github.com/last9-cdk/proc => ../proc

replace github.com/last9-cdk/tests => ../tests
