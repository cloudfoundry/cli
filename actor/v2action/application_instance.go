package v2action

import (
	"fmt"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
)

type ApplicationInstance ccv2.ApplicationInstanceStatus

// ApplicationInstancesNotFoundError is returned when a requested application is not
// found.
type ApplicationInstancesNotFoundError struct {
	ApplicationGUID string
}

func (e ApplicationInstancesNotFoundError) Error() string {
	return fmt.Sprintf("Application instances '%s' not found.", e.ApplicationGUID)
}

func (instance ApplicationInstance) StartTime() time.Time {
	return time.Now().Add(-1 * time.Duration(instance.Uptime) * time.Second)
}

func (actor Actor) GetApplicationInstancesByApplication(guid string) ([]ApplicationInstance, Warnings, error) {
	ccAppInstances, warnings, err := actor.CloudControllerClient.GetApplicationInstanceStatusesByApplication(guid)

	appInstances := []ApplicationInstance{}

	for _, appInstance := range ccAppInstances {
		appInstances = append(appInstances, ApplicationInstance(appInstance))
	}
	if _, ok := err.(cloudcontroller.ResourceNotFoundError); ok {
		return nil, Warnings(warnings), ApplicationInstancesNotFoundError{ApplicationGUID: guid}
	}

	return appInstances, Warnings(warnings), err
}
