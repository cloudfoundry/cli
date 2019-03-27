package v7action

import "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"

func (actor Actor) CreateSharedDomain(domainName string, internal bool) (Warnings, error) {
	_, warnings, err := actor.CloudControllerClient.CreateDomain(ccv3.Domain{
		Name:     domainName,
		Internal: internal,
	})
	return Warnings(warnings), err
}
