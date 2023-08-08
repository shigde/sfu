package auth

import (
	"math/rand"
	"net/http"
	"time"
)

const (
	requestTokenHEADER = "X-Req-Token"
	expiration         = 1 * time.Hour
)

var (
	letters      = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	tokenManager = newManager()
)

func TokenMiddleware(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqToken := r.Header.Get(requestTokenHEADER)
		if len(reqToken) < 1 {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		user, err := GetPrincipalFromSession(r)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		if currentToken := tokenManager.getToken(user.UUID); currentToken != reqToken {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		SetNewRequestToken(w, user.UUID)
		f(w, r)
	}
}

func SetNewRequestToken(w http.ResponseWriter, user string) {
	token := newRequestToken()
	tokenManager.setToken(user, token, expiration)
	w.Header().Set(requestTokenHEADER, token)
}

func newRequestToken() string {
	n := 34
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}
