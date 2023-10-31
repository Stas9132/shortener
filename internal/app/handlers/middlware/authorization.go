package middlware

import (
	"github.com/google/uuid"
	"net/http"
)

type authWriter struct {
	c *http.Cookie
	http.ResponseWriter
}

func (w authWriter) WriteHeader(statusCode int) {
	http.SetCookie(w, w.c)
	w.Header().Set("Authorization", w.c.Value)
	w.ResponseWriter.WriteHeader(statusCode)
}

func Authorization(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("auth")
		if err != nil {
			c = &http.Cookie{
				Name:  "auth",
				Value: uuid.NewString(),
			}
		}
		h.ServeHTTP(authWriter{
			c:              c,
			ResponseWriter: w,
		}, r)
	})
}
