package httpmetrics

import (
	"net/http"
	"testing"

	"github.com/last9/last9-cdk/go/proc"
	"github.com/last9/last9-cdk/go/tests"
	"github.com/last9/pat"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/go-playground/assert.v1"
)

func patHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(r.URL.Query().Get(":id")))
		}
}

func patHandlerNoWriteHeader() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(r.URL.Query().Get(":id")))
		}
}

func TestPatMux(t *testing.T) {
	t.Run("wrapped pat handler captures path", func(t *testing.T) {
		resetMetrics()

		m := pat.New()
		m.Get("/api/:id", REDHandler(patHandler()))
		m.Get("/metrics", promhttp.Handler())
		srv := tests.MakeServer(m)
		defer srv.Close()

		ids, err := tests.SendTestRequests(srv.URL, 2)
		if err != nil {
			t.Fatal(err)
		}

		o, err := tests.GetMetrics(srv.URL)
		if err != nil {
			t.Fatal(err)
		}

		req := o["http_requests_duration_milliseconds"]
		assert.Equal(t, len(ids) > 0, true)
		assert.Equal(t, 1, len(req.GetMetric()))
		assert.Equal(t, 7, assertLabels("/api/:id", getDomain(srv), req))
	})

	t.Run("wrapped pat mux captures path", func(t *testing.T) {
		resetMetrics()

		m := pat.New()
		m.Get("/api/:id", patHandler())
		m.Get("/metrics", promhttp.Handler())
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

		req := o["http_requests_duration_milliseconds"]
		assert.Equal(t, len(ids) > 0, true)
		assert.Equal(t, 1, len(req.GetMetric()))
		assert.Equal(t, 7, assertLabels("/api/:id", getDomain(srv), req))
	})

	t.Run("pat mux uses middleware", func(t *testing.T) {
		resetMetrics()

		m := pat.New()
		m.Get("/api/:id", patHandler())
		m.Get("/metrics", promhttp.Handler())
		m.Use(REDHandler)
		srv := tests.MakeServer(m)
		defer srv.Close()

		ids, err := tests.SendTestRequests(srv.URL, 2)
		if err != nil {
			t.Fatal(err)
		}

		o, err := tests.GetMetrics(srv.URL)
		if err != nil {
			t.Fatal(err)
		}

		req := o["http_requests_duration_milliseconds"]
		assert.Equal(t, len(ids) > 0, true)
		assert.Equal(t, 1, len(req.GetMetric()))
		assert.Equal(t, 7, assertLabels("/api/:id", getDomain(srv), req))
	})

	t.Run("pat mux uses reudundant middlewares", func(t *testing.T) {
		resetMetrics()

		m := pat.New()
		m.Get("/api/:id", REDHandler(patHandler()))
		m.Get("/metrics", promhttp.Handler())
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

		req := o["http_requests_duration_milliseconds"]
		assert.Equal(t, len(ids) > 0, true)
		assert.Equal(t, 1, len(req.GetMetric()))
		assert.Equal(t, 7, assertLabels("/api/:id", getDomain(srv), req))
	})

	t.Run("pat custom label middlewares", func(t *testing.T) {
		resetMetrics()

		m := pat.New()
		m.Get("/api/:id", patHandler())
		m.Get("/metrics", promhttp.Handler())
		m.Use(REDHandlerWithLabelMaker(
			func(r *http.Request, mux http.Handler) map[string]string {
				return map[string]string{
					labelPer:         "/path",
					proc.LabelTenant: r.URL.Query().Get(":id"),
				}
			},
		))

		srv := tests.MakeServer(m)
		defer srv.Close()

		if _, err := tests.SendTestRequests(srv.URL, 5); err != nil {
			t.Fatal(err)
		}

		o, err := tests.GetMetrics(srv.URL)
		if err != nil {
			t.Fatal(err)
		}

		req := o["http_requests_duration_milliseconds"]
		for _, m := range req.GetMetric() {
			for _, l := range m.GetLabel() {
				if l.GetName() == "tenant" && l.GetValue() == "" {
					t.Fatal("Expected non empty tenants")
				}
			}
		}
	})

	t.Run("status should != 0 even if WriteHeader is not called explicitly", func(t *testing.T){
		resetMetrics()

		m := pat.New()
		m.Get("/api/:id", REDHandler(patHandlerNoWriteHeader()))
		m.Get("/metrics", promhttp.Handler())
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

		req := o["http_requests_duration_milliseconds"]
		assert.Equal(t, len(ids) > 0, true)
		assert.Equal(t, 1, len(req.GetMetric()))
		assert.Equal(t, 7, assertLabels("/api/:id", getDomain(srv), req))
	})
}
