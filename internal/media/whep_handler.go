package media

import (
	"crypto/md5"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/shigde/sfu/internal/auth"
	"github.com/shigde/sfu/internal/stream"
)

func whep(spaceManager spaceGetCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/sdp")
		if err := auth.StartSession(w, r); err != nil {
			httpError(w, "error", http.StatusInternalServerError, err)
		}

		liveStream, space, err := getLiveStream(r, spaceManager)
		if err != nil {
			handleResourceError(w, err)
			return
		}

		offer, err := getSdpPayload(w, r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		user, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		userId, err := user.GetUuid()
		if err != nil {
			httpError(w, "error user", http.StatusBadRequest, err)
			return
		}

		answer, err := space.ListenLobby(r.Context(), offer, liveStream, userId)
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
