package shared

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	ccWrapper "code.cloudfoundry.org/cli/api/cloudcontroller/wrapper"
	"code.cloudfoundry.org/cli/api/uaa"
	uaaWrapper "code.cloudfoundry.org/cli/api/uaa/wrapper"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

// NewClients creates a new V3 Cloud Controller client and UAA client using the
// passed in config.
func NewClients(config command.Config, ui command.UI, targetCF bool) (*ccv3.Client, *uaa.Client, error) {
	ccWrappers := []ccv3.ConnectionWrapper{}

	verbose, location := config.Verbose()
	if verbose {
		ccWrappers = append(ccWrappers, ccWrapper.NewRequestLogger(ui.RequestLoggerTerminalDisplay()))
	}
	if location != nil {
		ccWrappers = append(ccWrappers, ccWrapper.NewRequestLogger(ui.RequestLoggerFileWriter(location)))
	}

	authWrapper := ccWrapper.NewUAAAuthentication(nil, config)

	ccWrappers = append(ccWrappers, authWrapper)
	ccWrappers = append(ccWrappers, ccWrapper.NewRetryRequest(2))

	ccClient := ccv3.NewClient(ccv3.Config{
		AppName:    config.BinaryName(),
		AppVersion: config.BinaryVersion(),
		Wrappers:   ccWrappers,
	})

	if !targetCF {
		return ccClient, nil, nil
	}

	if config.Target() == "" {
		return nil, nil, translatableerror.NoAPISetError{
			BinaryName: config.BinaryName(),
		}
	}

	_, err := ccClient.TargetCF(ccv3.TargetSettings{
		URL:               config.Target(),
		SkipSSLValidation: config.SkipSSLValidation(),
		DialTimeout:       config.DialTimeout(),
	})
	if err != nil {
		if v3Err, ok := err.(ccerror.V3UnexpectedResponseError); ok && v3Err.ResponseCode == http.StatusNotFound {
			return nil, nil, V3APIDoesNotExistError{Message: err.Error()}
		}

		return nil, nil, HandleError(err)
	}

	uaaClient := uaa.NewClient(uaa.Config{
		AppName:           config.BinaryName(),
		AppVersion:        config.BinaryVersion(),
		ClientID:          config.UAAOAuthClient(),
		ClientSecret:      config.UAAOAuthClientSecret(),
		DialTimeout:       config.DialTimeout(),
		SkipSSLValidation: config.SkipSSLValidation(),
		URL:               ccClient.UAA(),
	})

	if verbose {
		uaaClient.WrapConnection(uaaWrapper.NewRequestLogger(ui.RequestLoggerTerminalDisplay()))
	}
	if location != nil {
		uaaClient.WrapConnection(uaaWrapper.NewRequestLogger(ui.RequestLoggerFileWriter(location)))
	}

	uaaClient.WrapConnection(uaaWrapper.NewUAAAuthentication(uaaClient, config))
	uaaClient.WrapConnection(uaaWrapper.NewRetryRequest(2))

	authWrapper.SetClient(uaaClient)

	return ccClient, uaaClient, nil
}
