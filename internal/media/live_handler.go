package media

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	http2 "github.com/shigde/sfu/internal/http"
	"github.com/shigde/sfu/internal/stream"
)

func publishLiveStream(streamService *stream.LiveStreamService, liveService *stream.LiveLobbyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		liveStream, userId, err := readingRequestData(w, r, streamService)
		if err != nil {
			return
		}

		streamInfo, err := getStreamLiveInfoPayload(w, r)
		if err != nil {
			httpError(w, "invalid payload", http.StatusBadRequest, err)
			return
		}

		if err := liveService.StartLiveStream(r.Context(), liveStream, streamInfo, userId); err != nil {
			httpError(w, "error start live stream", http.StatusInternalServerError, err)
			return
		}
		w.WriteHeader(http.StatusCreated)
		return
	}
}

func stopLiveStream(streamService *stream.LiveStreamService, liveService *stream.LiveLobbyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		liveStream, userId, err := readingRequestData(w, r, streamService)
		if err != nil {
			return
		}

		if err := liveService.StopLiveStream(r.Context(), liveStream, userId); err != nil {
			httpError(w, "error can not stop live stream", http.StatusBadRequest, err)
			return
		}
		w.WriteHeader(http.StatusCreated)
		return
	}
}

func getStatusOfLiveStream(streamService *stream.LiveStreamService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		liveStream, _, err := getLiveStream(r, streamService)
		if err != nil {
			handleResourceError(w, err)
			return
		}

		if _, err = getUserFromSession(w, r); err != nil {
			httpError(w, "forbidden", http.StatusForbidden, err)
			return
		}

		if err := json.NewEncoder(w).Encode(liveStream.Lobby); err != nil {
			httpError(w, "stream invalid", http.StatusInternalServerError, err)
		}

		w.WriteHeader(http.StatusOK)
	}
}

func getStreamLiveInfoPayload(w http.ResponseWriter, r *http.Request) (*stream.LiveStreamInfo, error) {
	dec, err := http2.GetJsonPayload(w, r)
	if err != nil {
		return nil, err
	}
	info := &stream.LiveStreamInfo{}
	if err := dec.Decode(info); err != nil {
		return nil, http2.InvalidPayload
	}
	return info, nil
}

func readingRequestData(w http.ResponseWriter, r *http.Request, streamService *stream.LiveStreamService) (*stream.LiveStream, uuid.UUID, error) {
	liveStream, _, err := getLiveStream(r, streamService)
	if err != nil {
		handleResourceError(w, err)
		return nil, uuid.Nil, err
	}

	user, err := getUserFromSession(w, r)
	if err != nil {
		httpError(w, "forbidden", http.StatusForbidden, err)
		return nil, uuid.Nil, err
	}

	if liveStream.Account.UUID != user.UUID {
		httpError(w, "forbidden", http.StatusForbidden, err)
		return nil, uuid.Nil, err
	}

	userId, err := user.GetUuid()
	if err != nil {
		httpError(w, "internal error", http.StatusInternalServerError, err)
		return nil, uuid.Nil, err
	}
	return liveStream, userId, nil
}
