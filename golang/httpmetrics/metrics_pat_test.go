package httpmetrics

import (
	"log"
	"net/http"
	"testing"

	"github.com/last9/pat"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/go-playground/assert.v1"
)

func patHandler() http.HandlerFunc {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)

			log.Println("-----1 pat", r.URL.Query())
			log.Println("-----1 pat", r.Context().Value(pat.RouteKey))
			w.Write([]byte(r.URL.Query().Get(":id")))
		},
	)
}

func TestPatMux(t *testing.T) {
	t.Run("wrapped pat handler captures path", func(t *testing.T) {
		resetMetrics()

		m := pat.New()
		m.Get("/api/:id", Last9HttpHandler(patHandler()))
		m.Get("/metrics", promhttp.Handler())
		srv := makeServer(m)
		defer srv.Close()

		ids, err := sendTestRequests(srv.URL, 2)
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
		assert.Equal(t, 4, assertLabels("/api/:id", o["http_requests_duration"]))
	})

	t.Run("wrapped pat mux captures path", func(t *testing.T) {
		resetMetrics()

		m := pat.New()
		m.Get("/api/:id", patHandler())
		m.Get("/metrics", promhttp.Handler())
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
		assert.Equal(t, 4, assertLabels("/api/:id", o["http_requests_duration"]))
	})
}
