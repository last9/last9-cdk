package httpmetrics

import (
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/last9/last9-cdk/go/tests"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/go-playground/assert.v1"
)

func gochiHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(chi.URLParam(r, "id")))
	})
}

func TestGoChiMux(t *testing.T) {
	t.Run("wrapped go-chi handler captures path", func(t *testing.T) {
		resetMetrics()
		m := chi.NewRouter()
		m.Handle("/api/{id}", REDHandler(gochiHandler()))
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
		// fmt.Println(rms.GetMetric())
		assert.Equal(t, 1, len(rms.GetMetric()))
		assert.Equal(t, 7,
			assertLabels("/api/{id}", getDomain(srv), rms))
	})

	t.Run("wrapped go-chi mux captures path", func(t *testing.T) {
		resetMetrics()

		m := chi.NewRouter()
		m.Use(REDHandlerWithLabelMaker(figureOutLabelMaker))
		m.Handle("/api/{id}", gochiHandler())
		m.Handle("/metrics", promhttp.Handler())

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
		assert.Equal(t, 7,
			assertLabels("/api/{id}", getDomain(srv), rms))
	})

	t.Run("go-chi mux middleware captures path", func(t *testing.T) {
		resetMetrics()

		m := chi.NewRouter()
		m.Use(REDHandler)
		m.Handle("/api/{id}", gochiHandler())
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
		assert.Equal(t, 7,
			assertLabels("/api/{id}", getDomain(srv), rms))
	})
}
