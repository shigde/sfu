package request

import (
	"net/http"

	"github.com/gorilla/csrf"
)

const (
	csrfTokenHEADER = "X-Csrf-Token"
	csrfTokenCookie = "csrf"
)

var csrfMiddleware = csrf.Protect([]byte("32-byte-long-auth-key"),
	csrf.RequestHeader(csrfTokenHEADER),
	csrf.CookieName(csrfTokenCookie),
)

func Csrf(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		csrfMiddleware(f).ServeHTTP(w, r)
	}
}
