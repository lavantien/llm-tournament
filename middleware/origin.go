package middleware

import (
	"log"
	"net/http"
	"net/url"
	"strings"
)

// CheckOrigin rejects state-changing requests whose Origin header names a
// different host. Browsers send Origin on every cross-site form post, so a
// webpage in another tab cannot drive this local server's mutating routes,
// while curl and same-origin traffic pass through untouched.
func CheckOrigin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isSafeMethod(r.Method) {
			if origin := r.Header.Get("Origin"); origin != "" && !sameHost(origin, r.Host) {
				log.Printf("Rejected cross-origin %s %s from %s", r.Method, r.URL.Path, origin)
				http.Error(w, "Cross-origin request rejected", http.StatusForbidden)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func isSafeMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
}

func sameHost(originRaw, host string) bool {
	u, err := url.Parse(originRaw)
	if err != nil || u.Host == "" {
		return false
	}
	return strings.EqualFold(u.Host, host)
}
