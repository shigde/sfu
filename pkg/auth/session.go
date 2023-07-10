package auth

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))

func StartSession(w http.ResponseWriter, r *http.Request) error {
	user, ok := PrincipalFromContext(r.Context())
	if !ok {
		return fmt.Errorf("no loged in user")
	}

	userUuid, err := user.GetUUid()
	if err != nil {
		return fmt.Errorf("read user uuid: %w", err)
	}

	session, err := store.Get(r, "session.id")
	if err != nil {
		return fmt.Errorf("start session by reding session: %w", err)
	}

	session.Values["authenticated"] = true
	session.Values["userUuid"] = userUuid
	if err = session.Save(r, w); err != nil {
		return fmt.Errorf("start session by saving session id: %w", err)
	}
	return nil
}

func HasActiveSession(w http.ResponseWriter, r *http.Request) error {
	session, err := store.Get(r, "session.id")
	if err != nil {
		return fmt.Errorf("start session by reding session: %w", err)
	}
	session.Values["authenticated"] = true
	if err = session.Save(r, w); err != nil {
		return fmt.Errorf("start session by saving session id: %w", err)
	}
	return nil
}
