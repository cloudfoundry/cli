package director

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
	"golang.org/x/crypto/ssh"
)

type SSHOpts struct {
	Username  string
	PublicKey string
}

func NewSSHOpts(uuidGen boshuuid.Generator) (SSHOpts, string, error) {
	privKey, pubKey, err := makeSSHKeyPair()
	if err != nil {
		return SSHOpts{}, "", bosherr.WrapErrorf(err, "Generating SSH key pair")
	}

	nameSuffix, err := uuidGen.Generate()
	if err != nil {
		return SSHOpts{}, "", bosherr.WrapErrorf(err, "Generating unique SSH user suffix")
	}

	// username must be <= 20 for Windows
	nameSuffix = strings.Replace(nameSuffix, "-", "", -1)[0:15]

	sshOpts := SSHOpts{
		Username:  "bosh_" + nameSuffix,
		PublicKey: string(pubKey),
	}

	return sshOpts, string(privKey), nil
}

func makeSSHKeyPair() ([]byte, []byte, error) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	privKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privKey),
	}

	privKeyBuf := bytes.NewBufferString("")

	err = pem.Encode(privKeyBuf, privKeyPEM)
	if err != nil {
		return nil, nil, err
	}

	pub, err := ssh.NewPublicKey(&privKey.PublicKey)
	if err != nil {
		return nil, nil, err
	}

	return privKeyBuf.Bytes(), ssh.MarshalAuthorizedKey(pub), nil
}
