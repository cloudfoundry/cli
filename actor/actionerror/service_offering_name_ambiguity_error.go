package actionerror

import "code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccerror"

type ServiceOfferingNameAmbiguityError struct {
	ccerror.ServiceOfferingNameAmbiguityError
}

func (e ServiceOfferingNameAmbiguityError) Error() string {
	return e.ServiceOfferingNameAmbiguityError.Error() + "\nSpecify a broker by using the '-b' flag."
}
