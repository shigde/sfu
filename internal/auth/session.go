package auth

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
)

var (
	ErrNoUserSession           = errors.New("no user in session")
	ErrNotAuthenticatedSession = errors.New("not authenticated session")
	ErrNoCsrfTokenInSession    = errors.New("no csrf token in session")
)
var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))

func StartSession(w http.ResponseWriter, r *http.Request) error {
	user, ok := PrincipalFromContext(r.Context())
	if !ok {
		return fmt.Errorf("no loged in user")
	}

	userUuid, err := user.GetUuid()
	if err != nil {
		return fmt.Errorf("read user uuid: %w", err)
	}

	session, err := store.Get(r, "session.id")
	if err != nil {
		return fmt.Errorf("start session by reding session: %w", err)
	}
	session.ID = uuid.New().String()
	session.Values["authenticated"] = true
	session.Values["userUuid"] = userUuid.String()
	if err = session.Save(r, w); err != nil {
		return fmt.Errorf("start session by saving session id: %w", err)
	}
	return nil
}

func DeleteSession(w http.ResponseWriter, r *http.Request) error {
	session, err := getSession(r)
	if err != nil {
		return fmt.Errorf("reading session: %w", err)
	}
	session.Options.MaxAge = -1

	if err = session.Save(r, w); err != nil {
		return fmt.Errorf("deleting session: %w", err)
	}

	return nil
}

func GetPrincipalFromSession(r *http.Request) (*Principal, error) {
	session, err := getSession(r)
	if err != nil {
		return nil, fmt.Errorf("reading session: %w", err)
	}

	userId, ok := session.Values["userUuid"].(string)
	if !ok || len(userId) == 0 {
		return nil, ErrNoUserSession
	}

	return &Principal{
		UUID: userId,
	}, nil
}

func getSession(r *http.Request) (*sessions.Session, error) {
	session, err := store.Get(r, "session.id")
	if err != nil {
		return nil, fmt.Errorf("start session by reding session: %w", err)
	}

	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		return nil, ErrNotAuthenticatedSession
	}
	return session, nil
}
