package tests

import (
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/common/expfmt"

	dto "github.com/prometheus/client_model/go"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	rand.Seed(time.Now().UnixNano())
}

func getRandomId() string {
	min := 10
	max := 30
	return strconv.Itoa(rand.Intn(max-min+1) + min)
}

// ResetMetrics resets all counters to 0 values, before each test.
func ResetMetrics(metrics ...interface{ Reset() }) {
	for _, m := range metrics {
		m.Reset()
	}
}

// SendTestRequests sends a handful of sample requests to the server address
// and returns the random IDs that were used in those calls, in sequence.
func SendTestRequests(addr string, num int) ([]string, error) {
	var out []string

	for x := 0; x < num; x++ {
		r := getRandomId()
		out = append(out, r)
		res, err := http.Get(addr + "/api/" + r)
		if err != nil {
			return nil, errors.Wrap(err, "get sample")
		}

		defer res.Body.Close()
		if _, err := ioutil.ReadAll(res.Body); err != nil {
			return nil, errors.Wrap(err, "get sample")
		}
	}

	return out, nil
}

// GetMetrics returns a dump of the current Prometheus metrics.
func GetMetrics(addr string) (map[string]*dto.MetricFamily, error) {
	res, err := http.Get(addr + "/metrics")
	if err != nil {
		return nil, errors.Wrap(err, "metrics get")
	}

	defer res.Body.Close()

	out, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "body get")
	}

	var parser expfmt.TextParser
	m, err := parser.TextToMetricFamilies(strings.NewReader(string(out)))
	if err != nil {
		return nil, errors.Wrap(err, "parse metrics")
	}

	return m, nil

}

// makeServer is just syntatcit sugar around http.Handler that does nothing
// other than wrap it under a httptest.Handler
func MakeServer(mux http.Handler) *httptest.Server {
	return httptest.NewServer(mux)
}
