package v7action

import "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"

type Info ccv3.Info

func (actor Actor) GetRootResponse() (Info, Warnings, error) {
	info, warnings, err := actor.CloudControllerClient.GetInfo()
	if err != nil {
		return Info{}, Warnings(warnings), err
	}
	return Info(info), Warnings(warnings), nil
}
