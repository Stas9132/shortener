package middlware

import (
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"shortener/internal/logger"
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
			j, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{}).SigningString()
			if err != nil {
				logger.WithField("error", err).Errorln("error while create jwt token")
			}
			c = &http.Cookie{
				Name:  "auth",
				Value: j,
			}
		}
		h.ServeHTTP(authWriter{
			c:              c,
			ResponseWriter: w,
		}, r)
	})
}
