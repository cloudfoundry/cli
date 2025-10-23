package shared

import (
	"code.cloudfoundry.org/cfnetworking-cli-api/cfnetworking/cfnetv1"
	"code.cloudfoundry.org/cfnetworking-cli-api/cfnetworking/wrapper"
	oldUAA "code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/v8/api/uaa"
	"code.cloudfoundry.org/cli/v8/command"
	"code.cloudfoundry.org/cli/v8/command/translatableerror"
)

// uaaClientAdapter adapts the v8 UAA client to the interface expected by cfnetworking-cli-api
type uaaClientAdapter struct {
	client *uaa.Client
}

func (a *uaaClientAdapter) RefreshAccessToken(refreshToken string) (oldUAA.RefreshedTokens, error) {
	tokens, err := a.client.RefreshAccessToken(refreshToken)
	if err != nil {
		return oldUAA.RefreshedTokens{}, err
	}
	return oldUAA.RefreshedTokens{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		Type:         tokens.Type,
	}, nil
}

// NewNetworkingClient creates a new cfnetworking client.
func NewNetworkingClient(apiURL string, config command.Config, uaaClient *uaa.Client, ui command.UI) (*cfnetv1.Client, error) {
	if apiURL == "" {
		return nil, translatableerror.CFNetworkingEndpointNotFoundError{}
	}

	wrappers := []cfnetv1.ConnectionWrapper{}

	verbose, location := config.Verbose()
	if verbose {
		wrappers = append(wrappers, wrapper.NewRequestLogger(ui.RequestLoggerTerminalDisplay()))
	}
	if location != nil {
		wrappers = append(wrappers, wrapper.NewRequestLogger(ui.RequestLoggerFileWriter(location)))
	}

	authWrapper := wrapper.NewUAAAuthentication(&uaaClientAdapter{client: uaaClient}, config)
	wrappers = append(wrappers, authWrapper)

	wrappers = append(wrappers, wrapper.NewRetryRequest(config.RequestRetryCount()))

	return cfnetv1.NewClient(cfnetv1.Config{
		AppName:           config.BinaryName(),
		AppVersion:        config.BinaryVersion(),
		DialTimeout:       config.DialTimeout(),
		SkipSSLValidation: config.SkipSSLValidation(),
		URL:               apiURL,
		Wrappers:          wrappers,
	}), nil
}
