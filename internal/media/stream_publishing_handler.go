package media

import (
	"net/http"

	"github.com/shigde/sfu/internal/stream"
)

func publishLiveStream(streamService *stream.LiveStreamService, liveService *stream.LiveLobbyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		liveStream, _, err := getLiveStream(r, streamService)
		if err != nil {
			handleResourceError(w, err)
			return
		}

		user, err := getUserFromSession(w, r)
		if err != nil {
			return
		}

		if liveStream.Account.UUID != user.UUID {
			httpError(w, "forbidden", http.StatusForbidden, err)
			return
		}

		userId, _ := user.GetUuid()

		if err := liveService.StartLiveStream(r.Context(), liveStream, userId); err != nil {
			httpError(w, "error start live stream", http.StatusInternalServerError, err)
			return
		}
		w.WriteHeader(http.StatusCreated)
		return
	}
}

func stopLiveStream(streamService *stream.LiveStreamService, liveService *stream.LiveLobbyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//w.Header().Set("Content-Type", "application/json")
		//space, err := getSpace(r, manager)
		//if err != nil {
		//	handleResourceError(w, err)
		//	return
		//}
		//streams, err := space.LiveStreamRepo.All(r.Context())
		//if err != nil {
		//	httpError(w, "error reading stream list", http.StatusInternalServerError, err)
		//	return
		//}
		//
		//if err := json.NewEncoder(w).Encode(streams); err != nil {
		//	httpError(w, "error reading stream list", http.StatusInternalServerError, err)
		//}
	}
}

func getStatusOfLiveStream(streamService *stream.LiveStreamService, liveService *stream.LiveLobbyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//w.Header().Set("Content-Type", "application/json")
		//space, err := getSpace(r, manager)
		//if err != nil {
		//	handleResourceError(w, err)
		//	return
		//}
		//streams, err := space.LiveStreamRepo.All(r.Context())
		//if err != nil {
		//	httpError(w, "error reading stream list", http.StatusInternalServerError, err)
		//	return
		//}
		//
		//if err := json.NewEncoder(w).Encode(streams); err != nil {
		//	httpError(w, "error reading stream list", http.StatusInternalServerError, err)
		//}
	}
}
