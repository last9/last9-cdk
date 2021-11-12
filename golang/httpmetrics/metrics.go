package httpmetrics

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/last9/pat"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests per path",
		},
		[]string{"per", "hostname", "domain", "method", "program", "status"},
	)

	httpRequestsDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_requests_duration",
			Help: "HTTP requests duration per path",
		},
		[]string{"per", "hostname", "domain", "method", "program", "status"},
	)
)

func init() {
	// Metrics have to be registered to be exposed:
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestsDuration)
}

type responseWriter struct {
	w    http.ResponseWriter
	resp []byte
	code int
}

func (rw *responseWriter) Header() http.Header {
	return rw.w.Header()
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.code = statusCode
	rw.w.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	// rw.w.WriteHeader(rw.code)
	if rw.code >= http.StatusInternalServerError {
		rw.resp = data
	}

	return rw.w.Write(data)
}

func (rw *responseWriter) CloseNotify() <-chan bool {
	return rw.w.(http.CloseNotifier).CloseNotify()
}

type Grouper func(r *http.Request, mux http.Handler) string

type last9Ctx string

var enableMiddleware = last9Ctx("enabled")

func middlewarePreEnabled(r *http.Request) bool {
	rv := r.Context().Value(middlewarePreEnabled)
	if rv != nil && rv.(string) == "true" {
		return true
	}

	return false
}

func Last9HttpPatternHandler(g Grouper, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &responseWriter{w: w}

		if middlewarePreEnabled(r) {
			next.ServeHTTP(rw, r)
			return
		}

		start := time.Now()
		labels := prometheus.Labels{
			"program":  getProgamName(),
			"hostname": getHostname(),
			"domain":   r.URL.Host,
			"method":   r.Method,
		}

		ctx := context.WithValue(r.Context(), enableMiddleware, "true")
		r = r.WithContext(ctx)

		defer func() {
			// Status code and path can only be known AFTER the mux was invoked.
			labels["per"] = g(r, next)
			labels["status"] = strconv.Itoa(rw.code)

			httpRequestsTotal.With(labels).Inc()
			httpRequestsDuration.With(labels).Observe(
				float64(time.Since(start).Milliseconds()),
			)
		}()

		//add the header
		//call the wrapped handler
		next.ServeHTTP(rw, r)

	})
}

func Last9HttpHandler(next http.Handler) http.Handler {
	switch t := next.(type) {
	case *mux.Router:
		t.Use(Last9HttpHandler)
		return t
	case *pat.PatternServeMux:
		t.Use(Last9HttpHandler)
		return t
	}
	return Last9HttpPatternHandler(figureOutGrouper, next)
}

func figureOutGrouper(r *http.Request, m http.Handler) string {
	switch t := m.(type) {
	case *http.ServeMux:
		_, p := t.Handler(r)
		return p
	case *mux.Router: // gorilla mux uses this
		if cr := mux.CurrentRoute(r); cr != nil {
			if p, err := cr.GetPathTemplate(); err == nil {
				return p
			}
		}
	default:
		if rk := r.Context().Value(pat.RouteKey); rk != nil {
			return rk.(string)
		} else if cr := mux.CurrentRoute(r); cr != nil {
			if p, err := cr.GetPathTemplate(); err == nil {
				return p
			}
		}
	}

	return path.Clean(r.URL.Path)
}

// ServeMetrics exposes whatever prometheus metrics are, on specified Port
func ServeMetrics(port int) {
	log.Println("Serving metrics on", port)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
