package authenticators

import (
	"regexp"

	"golang.org/x/crypto/ssh"
)

type CompositeAuthenticator struct {
	authenticators map[*regexp.Regexp]PasswordAuthenticator
}

func NewCompositeAuthenticator(passwordAuthenticators ...PasswordAuthenticator) *CompositeAuthenticator {
	authenticators := map[*regexp.Regexp]PasswordAuthenticator{}
	for _, a := range passwordAuthenticators {
		authenticators[a.UserRegexp()] = a
	}
	return &CompositeAuthenticator{authenticators: authenticators}
}

func (a *CompositeAuthenticator) Authenticate(metadata ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
	for userRegexp, authenticator := range a.authenticators {
		if userRegexp.MatchString(metadata.User()) {
			return authenticator.Authenticate(metadata, password)
		}
	}

	return nil, InvalidCredentialsErr
}
