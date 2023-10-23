package handler

import (
	"encoding/json"
	"errors"
	"html"
	"net/http"
	"net/url"

	"github.com/shigde/sfu/internal/activitypub/instance"
	"github.com/shigde/sfu/internal/activitypub/models"
	"github.com/shigde/sfu/internal/activitypub/outbox"
	"github.com/shigde/sfu/internal/activitypub/services"
	"golang.org/x/exp/slog"
)

var (
	invalidContentType = errors.New("invalid content type")
	invalidPayload     = errors.New("invalid payload")
)

func GetRegisterHandler(
	config *instance.FederationConfig,
	actorService *services.ActorService,
	followRep *models.FollowRepository,
	sender *outbox.Sender,
) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if !config.Enable {
			http.Error(w, errNoFederationSupport.Error(), http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		accountPayload, err := getJsonAccountPayload(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if accountPayload.RegisterToken != config.RegisterToken {
			http.Error(w, "", http.StatusForbidden)
		}

		instanceActor, err := actorService.GetLocalInstanceActor(r.Context())
		if err != nil {
			slog.Error("getting local instance actor", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		remoteInstance, err := actorService.CreateActorFromRemoteAccount(r.Context(), accountPayload.AccountIri.String(), instanceActor)
		if err != nil {
			slog.Error("getting remote account as actor", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		follow := models.NewFollow(instanceActor, remoteInstance, config)
		follow, err = followRep.Add(r.Context(), follow)
		if err != nil {
			slog.Error("saving fallow", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := sender.SendFollowRequest(follow); err != nil {
			slog.Error("sending fallow request", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		return
	}
}

func getAccountIriFromGet(r *http.Request) (*url.URL, error) {
	iriString := r.URL.Query().Get("accountIri")
	if len(iriString) == 0 {
		return nil, errAccountIriNotFound
	}
	iri := html.UnescapeString(iriString)
	iriUrl, err := url.ParseRequestURI(iri)
	if err != nil {
		return nil, errAccountIriInvalid
	}

	return iriUrl, nil
}

func getJsonAccountPayload(w http.ResponseWriter, r *http.Request) (*registerPayload, error) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		return nil, invalidContentType
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var registerPayload registerPayload
	if err := dec.Decode(&registerPayload); err != nil {
		return nil, invalidPayload
	}

	iri := html.UnescapeString(registerPayload.AccountUrl)
	iriUrl, err := url.ParseRequestURI(iri)
	if err != nil {
		return nil, errAccountIriInvalid
	}
	registerPayload.AccountIri = iriUrl

	return &registerPayload, nil
}

type registerPayload struct {
	AccountUrl    string   `json:"accountUrl"`
	RegisterToken string   `json:"registerToken"`
	AccountIri    *url.URL `json:"-"`
}
