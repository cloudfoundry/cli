package keys

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"code.cloudfoundry.org/diego-ssh/helpers"
	"golang.org/x/crypto/ssh"
)

//go:generate counterfeiter -o fake_keys/fake_key_pair.go . KeyPair
type KeyPair interface {
	PrivateKey() ssh.Signer
	PEMEncodedPrivateKey() string

	PublicKey() ssh.PublicKey
	Fingerprint() string
	AuthorizedKey() string
}

//go:generate counterfeiter -o fake_keys/fake_ssh_key_factory.go . SSHKeyFactory
type SSHKeyFactory interface {
	NewKeyPair(bits int) (KeyPair, error)
}

var RSAKeyPairFactory SSHKeyFactory = &keyPairFactory{}

type keyPairFactory struct{}

func (r *keyPairFactory) NewKeyPair(bits int) (KeyPair, error) {
	return newRSA(bits)
}

type rsaKeyPair struct {
	encodedPrivateKey string
	privateKey        ssh.Signer
}

func newRSA(bits int) (KeyPair, error) {
	key, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, err
	}

	err = key.Validate()
	if err != nil {
		return nil, err
	}

	encodedPrivateKey := pem.EncodeToMemory(&pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   x509.MarshalPKCS1PrivateKey(key),
	})

	privateKey, err := ssh.ParsePrivateKey(encodedPrivateKey)
	if err != nil {
		return nil, err
	}

	return &rsaKeyPair{
		encodedPrivateKey: string(encodedPrivateKey),
		privateKey:        privateKey,
	}, nil
}

func (k *rsaKeyPair) PrivateKey() ssh.Signer {
	return k.privateKey
}

func (k *rsaKeyPair) PEMEncodedPrivateKey() string {
	return k.encodedPrivateKey
}

func (k *rsaKeyPair) PublicKey() ssh.PublicKey {
	return k.privateKey.PublicKey()
}

func (k *rsaKeyPair) Fingerprint() string {
	return helpers.MD5Fingerprint(k.PublicKey())
}

func (k *rsaKeyPair) AuthorizedKey() string {
	return string(ssh.MarshalAuthorizedKey(k.PublicKey()))
}
