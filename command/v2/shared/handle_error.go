package shared

import (
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/command"
)

func HandleError(err error) error {
	switch e := err.(type) {
	case cloudcontroller.RequestError:
		return command.APIRequestError{Err: e.Err}
	case cloudcontroller.UnverifiedServerError:
		return command.InvalidSSLCertError{API: e.URL}
	case cloudcontroller.SSLValidationHostnameError:
		return command.SSLCertErrorError{Message: e.Message}
	case cloudcontroller.APINotFoundError:
		return command.APINotFoundError{URL: e.URL}

	case v2action.ApplicationNotFoundError:
		return command.ApplicationNotFoundError{Name: e.Name}
	case v2action.ServiceInstanceNotFoundError:
		return command.ServiceInstanceNotFoundError{Name: e.Name}
	}

	return err
}
