package middleware

import (
	"context"
	"errors"
	"fmt"
	"github.com/Stas9132/shortener/internal/logger"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const key = "secret_key"

// Issuer struct
type Issuer struct {
	ID    string
	State string
}

// GetIssuer from context
func GetIssuer(ctx context.Context) *Issuer {
	s, ok := ctx.Value(Issuer{}).(*Issuer)
	if !ok {
		logger.Warn("No issuer")
	}
	if s == nil {
		return &Issuer{}
	}
	return s
}

type authWriter struct {
	c *http.Cookie
	http.ResponseWriter
}

// WriteHeader overridden metod
func (w authWriter) WriteHeader(statusCode int) {
	http.SetCookie(w, w.c)
	w.Header().Set("Authorization", w.c.Value)
	w.ResponseWriter.WriteHeader(statusCode)
}

// Authorization middleware
func Authorization(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("auth")
		iss := Issuer{
			ID:    uuid.NewString(),
			State: "NEW",
		}
		if err == nil {
			token, err2 := jwt.ParseWithClaims(c.Value, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.New("unexpected signing method")
				}
				return []byte(key), nil
			})
			if err2 == nil {
				if claims, ok := token.Claims.(*jwt.MapClaims); ok && token.Valid {
					id, _ := (*claims)["iss"].(string)
					iss = Issuer{
						ID:    id,
						State: "ESTABLISHED",
					}
				}
			}
			err = err2
		}
		if err != nil {
			logger.WithField("error", err).Info("Token error")
			j, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"iss": iss.ID,
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
		fmt.Println("aaaaaaaaaaaaaaaaaaaaaaaa", iss)
		h.ServeHTTP(authWriter{
			c:              c,
			ResponseWriter: w,
		}, r.WithContext(context.WithValue(r.Context(), Issuer{}, &iss)))
	})
}
