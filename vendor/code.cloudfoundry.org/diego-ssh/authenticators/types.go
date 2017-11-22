package authenticators

import (
	"regexp"

	"code.cloudfoundry.org/lager"

	"golang.org/x/crypto/ssh"
)

//go:generate counterfeiter -o fake_authenticators/fake_public_key_authenticator.go . PublicKeyAuthenticator
type PublicKeyAuthenticator interface {
	Authenticate(metadata ssh.ConnMetadata, publicKey ssh.PublicKey) (*ssh.Permissions, error)
	PublicKey() ssh.PublicKey
}

//go:generate counterfeiter -o fake_authenticators/fake_password_authenticator.go . PasswordAuthenticator
type PasswordAuthenticator interface {
	UserRegexp() *regexp.Regexp
	Authenticate(metadata ssh.ConnMetadata, password []byte) (*ssh.Permissions, error)
}

//go:generate counterfeiter -o fake_authenticators/fake_permissions_builder.go . PermissionsBuilder
type PermissionsBuilder interface {
	Build(logger lager.Logger, processGuid string, index int, metadata ssh.ConnMetadata) (*ssh.Permissions, error)
}
