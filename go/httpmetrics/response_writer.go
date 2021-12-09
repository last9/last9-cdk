package httpmetrics

import (
	"net/http"
	"sync"
)

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

var rwPool = sync.Pool{
	New: func() interface{} {
		return new(ResponseWriter)
	},
}

// NewResponseWriter returns a new ResponseWriter from memory pool.
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	o := rwPool.Get().(*ResponseWriter)

	// reset the fields to their default values.
	o.w = w
	return o
}

// FinishResponseWriter puts back the object to the pool
func FinishResponseWriter(rw *ResponseWriter) {
	if rw == nil {
		return
	}

	rw.w = nil
	rw.code = 0
	rw.resp = rw.resp[:0]
	rwPool.Put(rw)
}
