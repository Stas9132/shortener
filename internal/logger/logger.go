package logger

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

var Log *logrus.Logger = logrus.New()

func Initialize(level string) error {
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	Log.SetLevel(lvl)
	return nil
}

func RequestLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		h.ServeHTTP(w, r)
		Log.WithFields(logrus.Fields{
			"uri":      r.URL.RequestURI(),
			"method":   r.Method,
			"duration": time.Since(t),
		}).Infoln("Request")
	})
}
