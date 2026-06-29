package http

import (
	"fmt"
	nethttp "net/http"
	"time"

	"github.com/gift-app/api/internal/logger"
)

type responseRecorder struct {
	nethttp.ResponseWriter
	statusCode int
	bytes      int
}

func newResponseRecorder(w nethttp.ResponseWriter) *responseRecorder {
	return &responseRecorder{ResponseWriter: w, statusCode: nethttp.StatusOK}
}

func (rw *responseRecorder) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseRecorder) Write(body []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(body)
	rw.bytes += n
	return n, err
}

func LoggingMiddleware(next nethttp.Handler) nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		start := time.Now()
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = fmt.Sprintf("req_%d", start.UnixNano())
		}

		recorder := newResponseRecorder(w)
		recorder.Header().Set("X-Request-ID", requestID)

		logger.LogRequest(requestID, r.Method, r.URL.Path, r.URL.RawQuery, r.RemoteAddr, r.UserAgent())
		next.ServeHTTP(recorder, r)
		logger.LogResponse(requestID, r.Method, r.URL.Path, recorder.statusCode, time.Since(start), recorder.bytes)
	})
}
