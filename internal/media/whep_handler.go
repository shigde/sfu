package media

import (
	"crypto/md5"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/auth"
	"github.com/shigde/sfu/internal/stream"
)

func whepOffer(spaceManager spaceGetCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/sdp")
		user, err := auth.GetPrincipalFromSession(r)
		if err != nil {
			switch {
			case errors.Is(err, auth.ErrNotAuthenticatedSession):
				httpError(w, "no session", http.StatusForbidden, err)
			case errors.Is(err, auth.ErrNoUserSession):
				httpError(w, "no user session", http.StatusForbidden, err)
			default:
				httpError(w, "internal error", http.StatusInternalServerError, err)
			}
			return
		}

		userId, err := user.GetUuid()
		if err != nil {
			httpError(w, "error user", http.StatusBadRequest, err)
			return
		}

		liveStream, space, err := getLiveStream(r, spaceManager)
		if err != nil {
			handleResourceError(w, err)
			return
		}

		answer, err := space.StartListenLobby(r.Context(), liveStream, userId)
		if err != nil && errors.Is(err, stream.ErrLobbyNotActive) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if err != nil {
			httpError(w, "error build whep", http.StatusInternalServerError, err)
			return
		}

		response := []byte(answer.SDP)
		hash := md5.Sum(response)

		w.WriteHeader(http.StatusCreated)
		contentLen, err := w.Write(response)
		if err != nil {
			httpError(w, "error build response", http.StatusInternalServerError, err)
			return
		}

		w.Header().Set("etag", fmt.Sprintf("%x", hash))
		w.Header().Set("Content-Length", strconv.Itoa(contentLen))
	}
}

func whepAnswer(spaceManager spaceGetCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/sdp")
		user, err := auth.GetPrincipalFromSession(r)
		if err != nil {
			switch {
			case errors.Is(err, auth.ErrNotAuthenticatedSession):
				httpError(w, "no session", http.StatusForbidden, err)
			case errors.Is(err, auth.ErrNoUserSession):
				httpError(w, "no user session", http.StatusForbidden, err)
			default:
				httpError(w, "internal error", http.StatusInternalServerError, err)
			}
			return
		}
		userId, err := user.GetUuid()
		if err != nil {
			httpError(w, "error user", http.StatusBadRequest, err)
			return
		}

		liveStream, space, err := getLiveStream(r, spaceManager)
		if err != nil {
			handleResourceError(w, err)
			return
		}

		answer, err := getSdpPayload(w, r, webrtc.SDPTypeAnswer)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		_, err = space.ListenLobby(r.Context(), answer, liveStream, userId)
		if err != nil && errors.Is(err, stream.ErrLobbyNotActive) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if err != nil {
			httpError(w, "error build whep", http.StatusInternalServerError, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		if err != nil {
			httpError(w, "error build response", http.StatusInternalServerError, err)
			return
		}

		w.Header().Set("Content-Length", "0")
	}
}
