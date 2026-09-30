package middlewares

import (
	"fmt"
	"net/http"
)

func HSTS(maxAge int, includeSubdomains bool, preload bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if maxAge < 0 {
				maxAge = 0
			}

			value := fmt.Sprintf("max-age=%v", maxAge)

			if includeSubdomains {
				value += "; includeSubDomains"
			}

			if preload && maxAge >= 31536000 {
				value += "; preload"
			}

			w.Header().Set("Strict-Transport-Security", value)

			next.ServeHTTP(w, r)
		})
	}
}
