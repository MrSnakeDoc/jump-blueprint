package mw

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/MrSnakeDoc/jump/internal/logger"
)

// statusWriter captures status code and bytes written.
type statusWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	// Ensure status is set if handler wrote body without calling WriteHeader.
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	return n, err
}

// logger returns a middleware that logs one line per HTTP request using the provided logger.
func Log(loggerClient logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := &statusWriter{ResponseWriter: w}

			next.ServeHTTP(ww, r)

			reqID := middleware.GetReqID(r.Context())
			loggerClient.Info("http_request",
				logger.String("method", r.Method),
				logger.String("path", r.URL.Path),
				logger.Int("status", ww.status),
				logger.Int("bytes", ww.bytes),
				logger.Duration("duration", time.Since(start)),
				logger.String("remote_ip", r.RemoteAddr),
				logger.String("user_agent", r.UserAgent()),
				logger.String("request_id", reqID),
			)
		})
	}
}
