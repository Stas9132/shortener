package middleware

import (
	"github.com/Stas9132/shortener/internal/logger"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type responseData struct {
	status int
	size   int
}
type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

// Write - overridden method
func (r loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

// WriteHeader - overridden method
func (r loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// RequestLogger - middleware
func RequestLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   &responseData{},
		}
		h.ServeHTTP(lw, r)
		logger.WithFields(logrus.Fields{
			"uri":      r.URL.RequestURI(),
			"method":   r.Method,
			"duration": time.Since(t),
		}).Infoln("Request")
		logger.WithFields(logrus.Fields{
			"status": lw.responseData.status,
			"size":   lw.responseData.size,
		}).Infoln("Response")
	})
}
