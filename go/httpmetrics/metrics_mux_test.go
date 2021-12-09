package httpmetrics

import (
	"net/http"
	"testing"

	"github.com/last9/last9-cdk/go/tests"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
	"gopkg.in/go-playground/assert.v1"
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
		mux.Handle("/api/", REDHandler(basicHandler()))

		srv := tests.MakeServer(bindMetrics(mux))
		defer srv.Close()

		ids, err := tests.SendTestRequests(srv.URL, 10)
		if err != nil {
			t.Fatal(err)
		}

		o, err := tests.GetMetrics(srv.URL)
		if err != nil {
			t.Fatal(err)
		}

		req := o["http_requests_duration_milliseconds"]
		var count uint64
		for _, m := range req.Metric {
			count += m.GetHistogram().GetSampleCount()
		}

		assert.Equal(t, len(ids), int(count))
		assert.Equal(t, req.GetType(), dto.MetricType_HISTOGRAM)
		assert.Equal(t, 6, assertLabels("/api", getDomain(srv), req))
	})

	t.Run("wrapped mux captures pattern", func(t *testing.T) {
		resetMetrics()
		mux := http.NewServeMux()
		mux.Handle("/api/", basicHandler())
		srv := tests.MakeServer(REDHandler(bindMetrics(mux)))
		defer srv.Close()

		if _, err := tests.SendTestRequests(srv.URL, 10); err != nil {
			t.Fatal(err)
		}

		o, err := tests.GetMetrics(srv.URL)
		if err != nil {
			t.Fatal(err)
		}

		req := o["http_requests_duration_milliseconds"]

		assert.Equal(t, req.GetType(), dto.MetricType_HISTOGRAM)
		assert.Equal(t, 6, assertLabels("/api", getDomain(srv), req))
		assert.Equal(t, 7, assertLabels("/api/", getDomain(srv), req))
		var count uint64
		for _, m := range req.Metric {
			count += m.GetHistogram().GetSampleCount()
		}
		assert.Equal(t, 10, int(count))
	})

	t.Run("wrapped mux with custom grouper", func(t *testing.T) {
		resetMetrics()
		mux := http.NewServeMux()
		mux.Handle("/api/", CustomREDHandler(
			func(r *http.Request, mux http.Handler) map[string]string {
				return map[string]string{
					labelPer: "my_custom_path_static",
				}
			},
			basicHandler(),
		))

		srv := tests.MakeServer(bindMetrics(mux))
		defer srv.Close()

		if _, err := tests.SendTestRequests(srv.URL, 10); err != nil {
			t.Fatal(err)
		}

		o, err := tests.GetMetrics(srv.URL)
		if err != nil {
			t.Fatal(err)
		}

		req := o["http_requests_duration_milliseconds"]
		assert.Equal(t, req.GetType(), dto.MetricType_HISTOGRAM)
		assert.Equal(t, 6, assertLabels("/api", getDomain(srv), req))
		assert.Equal(t, 7,
			assertLabels("my_custom_path_static", getDomain(srv), req))
		var count uint64
		for _, m := range req.Metric {
			count += m.GetHistogram().GetSampleCount()
		}
		assert.Equal(t, 10, int(count))
	})
}
