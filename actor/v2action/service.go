package v2action

import "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"

type Service ccv2.Service

func (actor Actor) GetService(serviceGUID string) (Service, Warnings, error) {
	service, warnings, err := actor.CloudControllerClient.GetService(serviceGUID)
	return Service(service), Warnings(warnings), err
}
