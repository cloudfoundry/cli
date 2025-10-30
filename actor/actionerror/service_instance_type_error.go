package actionerror

import (
	"fmt"

	"code.cloudfoundry.org/cli/v8/resources"
)

type ServiceInstanceTypeError struct {
	Name         string
	RequiredType resources.ServiceInstanceType
}

func (e ServiceInstanceTypeError) Error() string {
	return fmt.Sprintf("The service instance '%s' is not %s", e.Name, e.RequiredType)
}
