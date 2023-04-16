package media

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

const maxPayloadByte = 1048576

var invalidPayload = errors.New("invalid payload")

func getStreamList(repository StreamRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		streams := repository.AllStreams()
		if err := json.NewEncoder(w).Encode(streams); err != nil {
			http.Error(w, "reading resources", http.StatusInternalServerError)
		}
	}
}
func getStream(repository StreamRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := mux.Vars(r)["id"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if streamResource, ok := repository.StreamById(id); ok {
			if err := json.NewEncoder(w).Encode(streamResource); err != nil {
				http.Error(w, "stream invalid", http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}
}

func deleteStream(repository StreamRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := mux.Vars(r)["id"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if repository.DeleteStream(id) {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func createStream(repository StreamRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var stream StreamResource
		if err := getPayload(w, r, &stream); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		id := repository.AddStream(stream)
		w.Header().Set("Location", fmt.Sprintf("%s/%s", r.URL.String(), id))
		w.WriteHeader(http.StatusCreated)
	}
}

func updateStream(repository StreamRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var stream StreamResource
		if err := getPayload(w, r, &stream); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if ok := repository.StreamUpdate(stream); !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNoContent)

	}
}

func getPayload(w http.ResponseWriter, r *http.Request, stream *StreamResource) error {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		return invalidPayload
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxPayloadByte)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(&stream)
	if err != nil {
		return invalidPayload
	}

	return nil
}
