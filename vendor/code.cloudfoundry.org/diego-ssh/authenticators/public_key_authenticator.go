package authenticators

import (
	"bytes"
	"errors"

	"golang.org/x/crypto/ssh"
)

type publicKeyAuthenticator struct {
	publicKey          ssh.PublicKey
	marshaledPublicKey []byte
}

func NewPublicKeyAuthenticator(publicKey ssh.PublicKey) PublicKeyAuthenticator {
	return &publicKeyAuthenticator{
		publicKey:          publicKey,
		marshaledPublicKey: publicKey.Marshal(),
	}
}

func (a *publicKeyAuthenticator) PublicKey() ssh.PublicKey {
	return a.publicKey
}

func (a *publicKeyAuthenticator) Authenticate(conn ssh.ConnMetadata, publicKey ssh.PublicKey) (*ssh.Permissions, error) {
	if bytes.Equal(publicKey.Marshal(), a.marshaledPublicKey) {
		return &ssh.Permissions{}, nil
	}

	return nil, errors.New("authentication failed")
}
