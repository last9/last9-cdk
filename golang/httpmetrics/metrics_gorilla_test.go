package httpmetrics

import (
	"net/http"
	"testing"

	"github.com/gorilla/mux"
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
		m.Handle("/api/{id}", Last9HttpHandler(gorillaHandler()))
		// bind metrics
		m.Handle("/metrics", promhttp.Handler())

		srv := makeServer(m)
		defer srv.Close()

		ids, err := sendTestRequests(srv.URL, 10)
		if err != nil {
			t.Fatal(err)
		}

		o, err := getMetrics(srv.URL)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(ids) > 0, true)
		assert.Equal(t, 1, len(o["http_requests_total"].GetMetric()))
		assert.Equal(t, 1, len(o["http_requests_duration"].GetMetric()))
		assert.Equal(t, 4, assertLabels("/api/{id}", o["http_requests_duration"]))
	})

	t.Run("wrapped gorilla mux captures path", func(t *testing.T) {
		resetMetrics()

		m := mux.NewRouter()
		m.Handle("/api/{id}", gorillaHandler())
		m.Handle("/metrics", promhttp.Handler())
		m.Use(Last9HttpHandler)
		srv := makeServer(Last9HttpHandler(m))
		defer srv.Close()

		ids, err := sendTestRequests(srv.URL, 2)
		if err != nil {
			t.Fatal(err)
		}

		o, err := getMetrics(srv.URL)
		if err != nil {
			t.Fatal(err)
		}

		// log.Println(o["http_requests_total"], o)
		assert.Equal(t, len(ids) > 0, true)
		assert.Equal(t, 1, len(o["http_requests_total"].GetMetric()))
		assert.Equal(t, 1, len(o["http_requests_duration"].GetMetric()))
		assert.Equal(t, 4, assertLabels("/api/{id}", o["http_requests_duration"]))
	})

	t.Run("gorilla mux middleware captures path", func(t *testing.T) {
		resetMetrics()

		m := mux.NewRouter()
		m.Handle("/api/{id}", gorillaHandler())
		m.Handle("/metrics", promhttp.Handler())
		m.Use(Last9HttpHandler)
		srv := makeServer(m)
		defer srv.Close()

		ids, err := sendTestRequests(srv.URL, 10)
		if err != nil {
			t.Fatal(err)
		}

		o, err := getMetrics(srv.URL)
		if err != nil {
			t.Fatal(err)
		}

		// log.Println(o["http_requests_total"], o)
		assert.Equal(t, len(ids) > 0, true)
		assert.Equal(t, 1, len(o["http_requests_total"].GetMetric()))
		assert.Equal(t, 1, len(o["http_requests_duration"].GetMetric()))
		assert.Equal(t, 4, assertLabels("/api/{id}", o["http_requests_duration"]))
	})
}
