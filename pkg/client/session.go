package client

import "net/http"

type Session struct {
	Bearer    string
	Cookie    *http.Cookie
	CsrfToken string
}

func (s *Session) SetBearer(bearer string) {
	s.Bearer = bearer
}
func (s *Session) SetCookie(cookie *http.Cookie) {
	s.Cookie = cookie
}

func (s *Session) SetCsrfToken(token string) {
	s.CsrfToken = token
}

func (s *Session) GetBearer() string {
	return s.Bearer
}

func (s *Session) GetCookie() *http.Cookie {
	return s.Cookie
}

func (s *Session) GetCsrfToken() string {
	return s.CsrfToken
}
