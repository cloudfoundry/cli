package common

import (
	"code.cloudfoundry.org/cli/actors/v3actions"
	"code.cloudfoundry.org/cli/api/cloudcontroller"
)

func HandleError(err error) error {
	switch e := err.(type) {
	case cloudcontroller.RequestError:
		return APIRequestError{Err: e.Err}
	case cloudcontroller.UnverifiedServerError:
		return InvalidSSLCertError{API: e.URL}

	case v3actions.ApplicationNotFoundError:
		return ApplicationNotFoundError{Name: e.Name}
	case v3actions.RunTaskError:
		return RunTaskError{Message: e.Error()}
	case v3actions.TaskWorkersUnavailableError:
		return RunTaskError{Message: "Task workers are unavailable."}
	}

	return err
}
