package common

import (
	"code.cloudfoundry.org/cli/actors/v2actions"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
)

func HandleError(err error) error {
	switch e := err.(type) {
	case ccv2.RequestError:
		return APIRequestError{Err: e.Err}
	case ccv2.UnverifiedServerError:
		return InvalidSSLCertError{API: e.URL}

	case v2actions.ApplicationNotFoundError:
		return ApplicationNotFoundError{Name: e.Name}
	case v2actions.ServiceInstanceNotFoundError:
		return ServiceInstanceNotFoundError{Name: e.Name}
	}

	return err
}
