package actionerror

import "code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccerror"

func EnrichAPIErrors(e error) error {
	switch err := e.(type) {
	case ccerror.ServiceOfferingNameAmbiguityError:
		return ServiceOfferingNameAmbiguityError{err}
	default:
		return e
	}
}
