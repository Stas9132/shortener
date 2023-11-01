package middlware

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"net/http"
	"shortener/internal/logger"
	"time"
)

const key = "secret_key"

type issuer struct {
}

func GetIssuer(ctx context.Context) string {
	s, ok := ctx.Value(issuer{}).(string)
	if !ok {
		logger.Warn("No issuer")
	}
	return s
}

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
		iss := uuid.NewString()
		if err == nil {
			token, err2 := jwt.ParseWithClaims(c.Value, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.New("unexpected signing method")
				}
				return []byte(key), nil
			})
			if err2 == nil {
				if claims, ok := token.Claims.(*jwt.MapClaims); ok && token.Valid {
					iss, _ = (*claims)["iss"].(string)
				}
			}
			err = err2
		}
		if err != nil {
			j, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"iss": iss,
				"exp": time.Now().Add(72 * time.Hour).Unix(),
			}).SignedString([]byte(key))
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
		}, r.WithContext(context.WithValue(r.Context(), issuer{}, iss)))
	})
}
