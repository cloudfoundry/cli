package common

import (
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/api/cloudcontroller"
)

func HandleError(err error) error {
	switch e := err.(type) {
	case cloudcontroller.RequestError:
		return APIRequestError{Err: e.Err}
	case cloudcontroller.UnverifiedServerError:
		return InvalidSSLCertError{API: e.URL}
	case cloudcontroller.APINotFoundError:
		return APINotFoundError{URL: e.URL}

	case v2action.ApplicationNotFoundError:
		return ApplicationNotFoundError{Name: e.Name}
	case v2action.ServiceInstanceNotFoundError:
		return ServiceInstanceNotFoundError{Name: e.Name}
	}

	return err
}
