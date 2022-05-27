package httpmetrics

import (
	"net/http"
	"path"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/mux"
	"github.com/last9/pat"
)

// LabelMaker is your factory of labelValues if:
// - Your mux does not provide a way to emit path patterns and emitting
// individual metric path will basically just explode the cardinality of
// the metric storage system
// - You want to construct metrics based on custom part of the request.
// You can emit a map of key values, as long as it is a part of the default
// labelSet.
// If you have a multi-tenant or a multi-cluster deployment, you can
// very well emit those too.
type LabelMaker func(r *http.Request, mux http.Handler) map[string]string

func figureOutLabelMaker(r *http.Request, m http.Handler) map[string]string {
	var perPath string

	switch t := m.(type) {
	case *http.ServeMux:
		_, p := t.Handler(r)
		perPath = p
		break
	case *mux.Router: // gorilla mux uses this
		if cr := mux.CurrentRoute(r); cr != nil {
			if p, err := cr.GetPathTemplate(); err == nil {
				perPath = p
				break
			}
		}
	case *chi.Mux:
		perPath = chi.RouteContext(r.Context()).RoutePattern()
	default:
		// pat
		if rk := r.Context().Value(pat.RouteKey); rk != nil {
			perPath = rk.(string)
			break
		} else if cr := mux.CurrentRoute(r); cr != nil {
			if p, err := cr.GetPathTemplate(); err == nil {
				perPath = p
				break
			}
		}
		// go-chi
		if chiCtx, ok := r.Context().Value(chi.RouteCtxKey).(*chi.Context); ok && chiCtx != nil {
			perPath = chiCtx.RoutePattern()
			break
		}
	}

	if len(perPath) == 0 {
		perPath = path.Clean(r.URL.Path)
	}

	return map[string]string{labelPer: perPath}
}
