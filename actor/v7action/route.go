package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

type Route ccv3.Route

func (actor Actor) CreateRoute(orgName, spaceName, domainName, hostname, path string) (Warnings, error) {
	allWarnings := Warnings{}
	domain, warnings, err := actor.GetDomainByName(domainName)
	allWarnings = append(allWarnings, warnings...)

	if err != nil {
		return allWarnings, err
	}

	org, warnings, err := actor.GetOrganizationByName(orgName)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	space, warnings, err := actor.GetSpaceByNameAndOrganization(spaceName, org.GUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	if path != "" && string(path[0]) != "/" {
		path = "/" + path
	}
	_, apiWarnings, err := actor.CloudControllerClient.CreateRoute(ccv3.Route{
		SpaceGUID:  space.GUID,
		DomainGUID: domain.GUID,
		Host:       hostname,
		Path:       path,
	})

	actorWarnings := Warnings(apiWarnings)
	allWarnings = append(allWarnings, actorWarnings...)

	if _, ok := err.(ccerror.RouteNotUniqueError); ok {
		return allWarnings, actionerror.RouteAlreadyExistsError{Err: err}
	}

	return allWarnings, err
}
