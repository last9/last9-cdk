package httpmetrics

import (
	"net/http"
	"testing"

	"gopkg.in/go-playground/assert.v1"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
)

func basicHandler() http.HandlerFunc {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Ok"))
		},
	)
}

// bindMetrics attaches a metric handler to the mux that is already serving
// the rqeuests
func bindMetrics(mux *http.ServeMux) http.Handler {
	mux.Handle("/metrics", promhttp.Handler())
	return mux
}

func TestMux(t *testing.T) {
	t.Run("wrapped http handler won't capture path", func(t *testing.T) {
		resetMetrics()

		mux := http.NewServeMux()
		mux.Handle("/api/", Last9HttpHandler(basicHandler()))

		srv := makeServer(bindMetrics(mux))
		defer srv.Close()

		ids, err := sendTestRequests(srv.URL, 10)
		if err != nil {
			t.Fatal(err)
		}

		o, err := getMetrics(srv.URL)
		if err != nil {
			t.Fatal(err)
		}

		// Count the no. of requests made to the first ID.
		idCount := map[string]int{}
		for x := 0; x < len(ids); x++ {
			if _, ok := idCount[ids[0]]; !ok {
				idCount[ids[x]] = 0
			}
			idCount[ids[x]]++
		}

		assert.Equal(t, len(idCount), len(o["http_requests_total"].GetMetric()))
		assert.Equal(t, o["http_requests_duration"].GetType(), dto.MetricType_HISTOGRAM)
		assert.Equal(t, 4, assertLabels("/api", getDomain(srv), o["http_requests_duration"]))
	})

	t.Run("wrapped mux captures pattern", func(t *testing.T) {
		resetMetrics()
		mux := http.NewServeMux()
		mux.Handle("/api/", basicHandler())
		srv := makeServer(Last9HttpHandler(bindMetrics(mux)))
		defer srv.Close()

		if _, err := sendTestRequests(srv.URL, 10); err != nil {
			t.Fatal(err)
		}

		o, err := getMetrics(srv.URL)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, o["http_requests_duration"].GetType(), dto.MetricType_HISTOGRAM)
		assert.Equal(t, 4, assertLabels("/api", getDomain(srv), o["http_requests_duration"]))
		assert.Equal(t, 5, assertLabels("/api/", getDomain(srv), o["http_requests_duration"]))
		assert.Equal(t, *(o["http_requests_total"].GetMetric()[0].Counter.Value), 10.0)
	})

	t.Run("wrapped mux with custom grouper", func(t *testing.T) {
		resetMetrics()
		mux := http.NewServeMux()
		mux.Handle("/api/", Last9HttpPatternHandler(
			func(r *http.Request, mux http.Handler) string {
				return "my_custom_path_static"
			},
			basicHandler(),
		))

		srv := makeServer(bindMetrics(mux))
		defer srv.Close()

		if _, err := sendTestRequests(srv.URL, 10); err != nil {
			t.Fatal(err)
		}

		o, err := getMetrics(srv.URL)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, o["http_requests_duration"].GetType(), dto.MetricType_HISTOGRAM)
		assert.Equal(t, 4, assertLabels("/api", getDomain(srv), o["http_requests_duration"]))
		assert.Equal(t, 5, assertLabels("my_custom_path_static", getDomain(srv), o["http_requests_duration"]))
		assert.Equal(t, *(o["http_requests_total"].GetMetric()[0].Counter.Value), 10.0)
	})
}
