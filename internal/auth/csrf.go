package auth

import (
	"math/rand"
	"net/http"
	"time"
)

const csrfTokenHEADER = "X-Csrf-Token"
const expiration = 1 * time.Hour

var (
	letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	csrf    = newManager()
)

func CsrfMiddleware(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqToken := r.Header.Get(csrfTokenHEADER)
		if len(reqToken) < 1 {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		user, err := GetPrincipalFromSession(r)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		if currentToken := csrf.getToken(user.UUID); currentToken != reqToken {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		SetNewCsrfToken(w, user.UUID)
		f(w, r)
	}
}

func SetNewCsrfToken(w http.ResponseWriter, user string) {
	token := newCsrfToken()
	csrf.setToken(user, token, expiration)
	w.Header().Set(csrfTokenHEADER, token)
}

func newCsrfToken() string {
	n := 34
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}
