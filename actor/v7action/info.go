package v7action

import "code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3"

type Root ccv3.Root
type Info ccv3.Info

func (actor Actor) GetRootResponse() (Root, Warnings, error) {
	root, warnings, err := actor.CloudControllerClient.GetRoot()
	if err != nil {
		return Root{}, Warnings(warnings), err
	}
	return Root(root), Warnings(warnings), nil
}

func (actor Actor) GetInfoResponse() (Info, Warnings, error) {
	info, warnings, err := actor.CloudControllerClient.GetInfo()
	if err != nil {
		return Info{}, Warnings(warnings), err
	}
	return Info(info), Warnings(warnings), nil
}
