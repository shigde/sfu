package request

import (
	"crypto/rsa"
	"encoding/json"
	"net/http"

	"github.com/shigde/sfu/internal/activitypub/crypto"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"golang.org/x/exp/slog"
)

type SignedResponse struct {
	signer *crypto.Signer
}

func NewSignedResponse(signer *crypto.Signer) *SignedResponse {
	return &SignedResponse{
		signer: signer,
	}
}

func (r *SignedResponse) WriteResponse(payload []byte, w http.ResponseWriter, publicKey crypto.PublicKey, privateKey *rsa.PrivateKey) error {
	w.Header().Set("Content-Type", "application/activity+json")

	if err := r.signer.SignResponse(w, payload, publicKey, privateKey); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("unable to sign response", "err", err)
		return err
	}

	if _, err := w.Write(payload); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	return nil
}

func (r *SignedResponse) WriteStreamResponse(item vocab.Type, w http.ResponseWriter, publicKey crypto.PublicKey, privateKey *rsa.PrivateKey) error {
	var jsonmap map[string]interface{}
	jsonmap, _ = streams.Serialize(item)
	b, err := json.Marshal(jsonmap)
	if err != nil {
		return err
	}

	return r.WriteResponse(b, w, publicKey, privateKey)
}
