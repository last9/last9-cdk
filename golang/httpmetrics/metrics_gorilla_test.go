package httpmetrics

import (
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/last9-cdk/tests"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/go-playground/assert.v1"
)

func gorillaHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(vars["id"]))
	})
}

func TestGorillaMux(t *testing.T) {
	t.Run("wrapped gorilla handler captures path", func(t *testing.T) {
		resetMetrics()

		m := mux.NewRouter()
		m.Handle("/api/{id}", REDHandler(gorillaHandler()))
		// bind metrics
		m.Handle("/metrics", promhttp.Handler())

		srv := tests.MakeServer(m)
		defer srv.Close()

		ids, err := tests.SendTestRequests(srv.URL, 10)
		if err != nil {
			t.Fatal(err)
		}

		o, err := tests.GetMetrics(srv.URL)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(ids) > 0, true)
		rms := o["http_requests_duration_milliseconds"]
		assert.Equal(t, 1, len(rms.GetMetric()))
		assert.Equal(t, 7, assertLabels("/api/{id}", getDomain(srv), rms))
	})

	t.Run("wrapped gorilla mux captures path", func(t *testing.T) {
		resetMetrics()

		m := mux.NewRouter()
		m.Handle("/api/{id}", gorillaHandler())
		m.Handle("/metrics", promhttp.Handler())
		m.Use(REDHandler)
		srv := tests.MakeServer(REDHandler(m))
		defer srv.Close()

		ids, err := tests.SendTestRequests(srv.URL, 2)
		if err != nil {
			t.Fatal(err)
		}

		o, err := tests.GetMetrics(srv.URL)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(ids) > 0, true)
		rms := o["http_requests_duration_milliseconds"]
		assert.Equal(t, 1, len(rms.GetMetric()))
		assert.Equal(t, 7, assertLabels("/api/{id}", getDomain(srv), rms))
	})

	t.Run("gorilla mux middleware captures path", func(t *testing.T) {
		resetMetrics()

		m := mux.NewRouter()
		m.Handle("/api/{id}", gorillaHandler())
		m.Handle("/metrics", promhttp.Handler())
		m.Use(REDHandler)
		srv := tests.MakeServer(m)
		defer srv.Close()

		ids, err := tests.SendTestRequests(srv.URL, 10)
		if err != nil {
			t.Fatal(err)
		}

		o, err := tests.GetMetrics(srv.URL)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(ids) > 0, true)
		rms := o["http_requests_duration_milliseconds"]
		assert.Equal(t, 1, len(rms.GetMetric()))
		assert.Equal(t, 7, assertLabels("/api/{id}", getDomain(srv), rms))
	})
}
