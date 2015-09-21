package credential

import (
	"errors"

	"github.com/cloudfoundry/cli/plugin"
)

type CredentialFactory interface {
	Get() (Credential, error)
}

type credFactory struct {
	cli plugin.CliConnection
}

func NewCredentialFactory(cli plugin.CliConnection) CredentialFactory {
	return &credFactory{cli: cli}
}

type Credential struct {
	Token string
}

func (credFactory *credFactory) Get() (Credential, error) {
	var cred Credential

	output, err := credFactory.cli.CliCommandWithoutTerminalOutput("oauth-token")
	if err != nil {
		return cred, errors.New("Failed to acquire oauth token")
	}

	cred.Token = output[len(output)-1]

	return cred, nil
}
