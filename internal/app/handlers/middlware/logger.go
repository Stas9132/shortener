package middlware

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"shortener/internal/logger"
	"time"
)

type responseData struct {
	status int
	size   int
}
type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

func (r loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

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
