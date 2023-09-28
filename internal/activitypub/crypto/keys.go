package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/url"

	"golang.org/x/exp/slog"
)

// GetPublicKey will return the public key for the provided actor.
func GetPublicKey(actorIRI *url.URL, key string) PublicKey {

	idURL, err := url.Parse(actorIRI.String() + "#main-key")
	if err != nil {
		slog.Error("unable to parse actor iri string", "idURL", idURL, "err", err)
	}

	return PublicKey{
		ID:           idURL,
		Owner:        actorIRI,
		PublicKeyPem: key,
	}
}

// GetPrivateKey will return the internal server private key.
func GetPrivateKey(key string) *rsa.PrivateKey {
	block, _ := pem.Decode([]byte(key))
	if block == nil {
		slog.Error("failed to parse PEM block containing the key")
		return nil
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		slog.Error("unable to parse private key", "err", err)
		return nil
	}

	return priv
}

// GenerateKeys will generate the private/public key pair needed for federation.
func GenerateKeys() ([]byte, []byte, error) {
	// generate key
	privatekey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		slog.Error("Cannot generate RSA key", "err", err)
		return nil, nil, err
	}
	publickey := &privatekey.PublicKey

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privatekey)
	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}
	privatePem := pem.EncodeToMemory(privateKeyBlock)

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publickey)
	if err != nil {
		slog.Error("error when dumping publickey:", "err", err)
		return nil, nil, err
	}
	publicKeyBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}
	publicPem := pem.EncodeToMemory(publicKeyBlock)

	return privatePem, publicPem, nil
}
