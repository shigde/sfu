package crypto

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rsa"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-fed/httpsig"
	"golang.org/x/exp/slog"
)

type keyGetter interface {
	GetPublicKey(ctx context.Context, actorIRI *url.URL) (string, error)
	GetPrivateKey(ctx context.Context, actorIRI *url.URL) (string, error)
	GetKeyPair(ctx context.Context, actorIRI *url.URL) (publicKey string, privateKey string, err error)
}

type Signer struct {
	keyStore keyGetter
}

func NewSigner(keyStore keyGetter) *Signer {
	return &Signer{keyStore}
}

// SignResponse will sign a response using the provided response body and public key.
func (s *Signer) SignResponse(w http.ResponseWriter, body []byte, publicKey PublicKey, privateKey *rsa.PrivateKey) error {
	return s.signResponse(privateKey, *publicKey.ID, body, w)
}

func (s *Signer) signResponse(privateKey crypto.PrivateKey, pubKeyID url.URL, body []byte, w http.ResponseWriter) error {
	prefs := []httpsig.Algorithm{httpsig.RSA_SHA256}
	digestAlgorithm := httpsig.DigestSha256

	// The "Date" and "Digest" headers must already be set on r, as well as r.URL.
	headersToSign := []string{}
	if body != nil {
		headersToSign = append(headersToSign, "digest")
	}

	signer, _, err := httpsig.NewSigner(prefs, digestAlgorithm, headersToSign, httpsig.Signature, 0)
	if err != nil {
		return err
	}

	// If r were a http.ResponseWriter, call SignResponse instead.
	return signer.SignResponse(privateKey, pubKeyID.String(), w, body)
}

// SignRequest will sign an ounbound request given the provided body.
func (s *Signer) SignRequest(req *http.Request, body []byte, actorIRI *url.URL) error {
	publicKeyString, privateKeyString, err := s.keyStore.GetKeyPair(context.Background(), actorIRI)

	if err != nil {
		return fmt.Errorf("getting crypto key pare from key store: %w", err)
	}

	publicKey := GetPublicKey(actorIRI, publicKeyString)
	privateKey := GetPrivateKey(privateKeyString)

	return s.signRequest(privateKey, publicKey.ID.String(), body, req)
}

func (s *Signer) signRequest(privateKey crypto.PrivateKey, pubKeyID string, body []byte, r *http.Request) error {
	prefs := []httpsig.Algorithm{httpsig.RSA_SHA256}
	digestAlgorithm := httpsig.DigestSha256

	date := time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
	r.Header["Date"] = []string{date}
	r.Header["Host"] = []string{r.URL.Host}
	r.Header["Accept"] = []string{`application/ld+json; profile="https://www.w3.org/ns/activitystreams"`}

	// The "Date" and "Digest" headers must already be set on r, as well as r.URL.
	headersToSign := []string{httpsig.RequestTarget, "host", "date"}
	if body != nil {
		headersToSign = append(headersToSign, "digest")
	}

	signer, _, err := httpsig.NewSigner(prefs, digestAlgorithm, headersToSign, httpsig.Signature, 0)
	if err != nil {
		return err
	}

	// If r were a http.ResponseWriter, call SignResponse instead.
	return signer.SignRequest(privateKey, pubKeyID, r, body)
}

// CreateSignedRequest will create a signed POST request of a payload to the provided destination.
func (s *Signer) CreateSignedRequest(payload []byte, url *url.URL, fromActorIRI *url.URL, release string) (*http.Request, error) {
	slog.Debug("Sending", "payload", string(payload), "toUrl", url)

	req, _ := http.NewRequest("POST", url.String(), bytes.NewBuffer(payload))

	ua := fmt.Sprintf("%s; https://stream.shig.de", release)
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Content-Type", "application/activity+json")

	if err := s.SignRequest(req, payload, fromActorIRI); err != nil {
		slog.Error("error signing request:", "err", err)
		return nil, err
	}

	return req, nil
}
