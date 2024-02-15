package middleware

import (
	"github.com/Stas9132/shortener/config"
	"net"
	"net/http"
)

// TrustedSubnet middleware
func TrustedSubnet(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, ipnet, err := net.ParseCIDR(config.C.TrustedSubnet)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		if !ipnet.Contains(net.ParseIP(r.RemoteAddr)) {
			http.Error(w, "You're out of trusted subnet", http.StatusForbidden)
			return
		}
		h.ServeHTTP(w, r)
	})
}
