package httpmetrics

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"

	"github.com/last9/last9-cdk/go/proc"
	"github.com/last9/last9-cdk/go/tests"
	dto "github.com/prometheus/client_model/go"
)

func resetMetrics() {
	tests.ResetMetrics(httpRequestsDuration)
}

func getDomain(s *httptest.Server) string {
	return strings.Split(s.URL, "//")[1]
}

// assertLabels returns a count of how many "expected" labels match
// for the provided endpoint pattern
// Example:
// given http requests that may or may not use a pattern identifier, the rest
// of the fields do not change in the metric.
// So a request to /api/1 will yield
// http_requests{program=,hostname=,status=,per=[either /api/1 or /api/:id}
// where program hostname and status won't change.
func assertLabels(per string, domain string, m *dto.MetricFamily) int {
	success := 0
	labels := m.GetMetric()[0].GetLabel()
	for _, l := range labels {
		val := *(l.Value)
		switch *(l.Name) {
		case proc.LabelTenant:
			if val == "" {
				success++
			}
		case proc.LabelCluster:
			if val == "" {
				success++
			}
		case proc.LabelProgram:
			if val == proc.GetProgamName() {
				success++
			}
		case proc.LabelHostname:
			if val == proc.GetHostname() {
				success++
			}
		case labelStatus:
			if val == strconv.Itoa(http.StatusOK) {
				success++
			}
		case labelPer:
			if val == per {
				success++
			}
		case labelDomain:
			if val == domain {
				success++
			}
		}
	}
	return success
}
