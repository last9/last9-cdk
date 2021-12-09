package httpmetrics

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/last9/last9-cdk/go/proc"
	"github.com/last9/pat"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	labelDomain = "domain"
	labelMethod = "method"
	labelStatus = "status"
	labelPer    = "per"
)

var (
	// defaultLabels that are provided to the Request metric.
	defaultLabels = []string{
		labelPer, proc.LabelHostname, labelDomain, labelMethod, proc.LabelProgram,
		labelStatus, proc.LabelTenant, proc.LabelCluster,
	}

	// the ONLY metric that we emit is httpRequestsDuration
	// which can provide for all three values:
	// - Rate (every histogram has a _sum and _count!!)
	// - Errors (by observing the status)
	// - Duration (It's a histogram!!)
	httpRequestsDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_requests_duration_milliseconds",
			Help:    "HTTP requests duration per path",
			Buckets: proc.LatencyBins,
		},
		defaultLabels,
	)

	enableMiddleware = last9Ctx("enabled")
)

func init() {
	// Metrics have to be registered to be exposed:
	prometheus.MustRegister(httpRequestsDuration)
}

// ResponseWriter is a status hijacker since http.ResponseWriter is an
// interface and unlike http.Request it cannot expose the value of the status
// once previously set during the lifetime of a handler.
// We rely on the status code to be emitted as one of the labels.
type ResponseWriter struct {
	w    http.ResponseWriter
	resp []byte
	code int
}

func (rw *ResponseWriter) Header() http.Header {
	return rw.w.Header()
}

// Code returns the statusCode on the way out. Do note that if this code is
// 0 that means that the Write was not called yet. It will be a non-zero
// status only once the Write has been called.
func (rw *ResponseWriter) Code() int {
	return rw.code
}

func (rw *ResponseWriter) WriteHeader(statusCode int) {
	rw.code = statusCode
	rw.w.WriteHeader(statusCode)
}

func (rw *ResponseWriter) Write(data []byte) (int, error) {
	// rw.w.WriteHeader(rw.code)
	if rw.code >= http.StatusInternalServerError {
		rw.resp = data
	} else if rw.code == 0 {
		rw.code = http.StatusOK
	}

	return rw.w.Write(data)
}

func (rw *ResponseWriter) CloseNotify() <-chan bool {
	return rw.w.(http.CloseNotifier).CloseNotify()
}

type last9Ctx string

// middlewarePreEnabled looks for context key to rule out if the middleware
// was pre-applied.
//
// When could this happen?
// Imagine a scenario where a handler was wrapped as a middleware
// as 		m.Get("/api/:id", REDHandler(patHandler()))
// and subsequently, the whole mux was ALSO wrapped
// as 		m.Use(REDHandler)
// Only of the two middlewares is worth executing.
func middlewarePreEnabled(r *http.Request) bool {
	rv := r.Context().Value(middlewarePreEnabled)
	if rv != nil && rv.(string) == "true" {
		return true
	}

	return false
}

// CustomREDHandler is a 3rd way to wrap a handler with a custom labelMaker
// This may become the most favorible middleware for developers who want to
// wrap each handler of theirs, instead of wrapping the entire mux.
//
// In scenarios where the mux does not support Middlewares out-of-the-box
// like pat, does not have a .Use method this function becomes the go-to.
//
// How to use?
// mux.Handle("/api/", CustomREDHandler(labelMaker, basicHandler()))
func CustomREDHandler(g LabelMaker, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &ResponseWriter{w: w}

		// If the middleware was already executed, skip this.
		// read the function definition for scenarios where this is applicable.
		if middlewarePreEnabled(r) {
			next.ServeHTTP(rw, r)
			return
		}

		start := time.Now()
		labels := prometheus.Labels{
			proc.LabelProgram:  proc.GetProgamName(),
			proc.LabelHostname: proc.GetHostname(),
			proc.LabelTenant:   "", // default tenant is empty
			proc.LabelCluster:  "", // default cluster is empty
			labelDomain:        r.Host,
			labelMethod:        r.Method,
		}

		ctx := context.WithValue(r.Context(), enableMiddleware, "true")
		r = r.WithContext(ctx)

		defer func() {
			// Status code and path can only be known AFTER the mux was invoked.
			// Some of the middleware alter the request BUT they create a new
			// request with context so the original request is untampered.
			// Hence delay this as late as possible to get the freshest/latest
			// value of the parameters.
			for k, v := range g(r, next) {
				// run through the defaultLabels and attempt to set, ONLY if
				// its an expected labelKey. An attempt to set something else
				// results in prometheus client library panicking, and that would
				// mean NO metric.
				for _, l := range defaultLabels {
					if k == l {
						labels[k] = v
						break
					}
				}
			}

			// Status code can only be known AFTER the mux was invoked.
			labels[labelStatus] = strconv.Itoa(rw.code)

			httpRequestsDuration.With(labels).Observe(
				float64(time.Since(start).Milliseconds()),
			)
		}()

		//call the wrapped handler
		next.ServeHTTP(rw, r)

	})
}

// REDHandlerWithLabelMaker is the 2nd choice of wrapping the entire Mux
// with a middleware. Passing the middleware to a mux is a fairly common
// technique with the likes of gorilla etc.
// How to Use?
// m.Use(REDHandlerWithLabelMaker(labelMaker))
func REDHandlerWithLabelMaker(g LabelMaker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		switch t := next.(type) {
		case *mux.Router:
			t.Use(REDHandlerWithLabelMaker(g))
			return t
		case *pat.PatternServeMux:
			t.Use(REDHandlerWithLabelMaker(g))
			return t
		}
		return CustomREDHandler(g, next)
	}
}

// REDHandler is a REDHandlerWithLabelMaker where default labelMaker is used.
// If you have custom metric emission where you need to extract unique parts
// of the request path, body etc. use REDHandlerWithLabelMaker instead
var REDHandler = REDHandlerWithLabelMaker(figureOutLabelMaker)

// ServeMetrics exposes whatever prometheus metrics are, on specified Port
func ServeMetrics(port int) {
	proc.ServeMetrics(port)
}
