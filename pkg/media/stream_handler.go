package media

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/shigde/sfu/pkg/engine"
)

func getStreamList(repository *engine.RtpStreamRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		streams := repository.All()
		if err := json.NewEncoder(w).Encode(streams); err != nil {
			http.Error(w, "reading resources", http.StatusInternalServerError)
		}
	}
}
func getStream(repository *engine.RtpStreamRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := mux.Vars(r)["id"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if streamResource, ok := repository.FindById(id); ok {
			if err := json.NewEncoder(w).Encode(streamResource); err != nil {
				http.Error(w, "stream invalid", http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}
}

func deleteStream(repository *engine.RtpStreamRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := mux.Vars(r)["id"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if repository.Delete(id) {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func createStream(repository *engine.RtpStreamRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var stream engine.RtpStream
		if err := getStreamResourcePayload(w, r, &stream); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		id := repository.Add(&stream)
		w.Header().Set("Location", fmt.Sprintf("%s/%s", r.URL.String(), id))
		w.WriteHeader(http.StatusCreated)
	}
}

func updateStream(repository *engine.RtpStreamRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var stream engine.RtpStream
		if err := getStreamResourcePayload(w, r, &stream); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if ok := repository.Update(&stream); !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNoContent)

	}
}

func getStreamResourcePayload(w http.ResponseWriter, r *http.Request, stream *engine.RtpStream) error {
	dec, err := getPayload(w, r)
	if err != nil {
		return err
	}

	if err := dec.Decode(&stream); err != nil {
		return invalidPayload
	}

	return nil
}
