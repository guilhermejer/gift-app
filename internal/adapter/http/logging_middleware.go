package http

import (
	"bytes"
	"fmt"
	"io"
	nethttp "net/http"
	"time"

	"github.com/gift-app/api/internal/logger"
)

type responseRecorder struct {
	nethttp.ResponseWriter
	statusCode int
	bytes      int
	body       []byte
	truncated  bool
}

const maxLoggedPayloadBytes = 1 << 20

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
	remaining := maxLoggedPayloadBytes - len(rw.body)
	if remaining > 0 {
		toCopy := len(body)
		if toCopy > remaining {
			toCopy = remaining
			rw.truncated = true
		}
		rw.body = append(rw.body, body[:toCopy]...)
	} else if len(body) > 0 {
		rw.truncated = true
	}
	return n, err
}

func LoggingMiddleware(next nethttp.Handler) nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		start := time.Now()
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = fmt.Sprintf("req_%d", start.UnixNano())
		}

		requestPayload := ""
		if r.Body != nil {
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				requestPayload = fmt.Sprintf("<<failed to read request payload: %v>>", err)
			} else {
				requestPayload = string(bodyBytes)
				r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			}
		}

		recorder := newResponseRecorder(w)
		recorder.Header().Set("X-Request-ID", requestID)

		logger.LogRequest(requestID, r.Method, r.URL.Path, r.URL.RawQuery, r.RemoteAddr, r.UserAgent(), requestPayload)
		next.ServeHTTP(recorder, r)

		responsePayload := string(recorder.body)
		if recorder.truncated {
			responsePayload += "<<truncated>>"
		}

		logger.LogResponse(requestID, r.Method, r.URL.Path, recorder.statusCode, time.Since(start), recorder.bytes, responsePayload)
	})
}
