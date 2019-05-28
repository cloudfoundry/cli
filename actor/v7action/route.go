package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

type Route struct {
	GUID       string
	SpaceGUID  string
	DomainGUID string
	Host       string
	Path       string
	DomainName string
	SpaceName  string
}

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

func (actor Actor) GetRoutesBySpace(spaceGUID string) ([]Route, Warnings, error) {
	allWarnings := Warnings{}

	routes, warnings, err := actor.CloudControllerClient.GetRoutes(ccv3.Query{
		Key:    ccv3.SpaceGUIDFilter,
		Values: []string{spaceGUID},
	})
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	spaces, warnings, err := actor.CloudControllerClient.GetSpaces(ccv3.Query{
		Key:    ccv3.GUIDFilter,
		Values: []string{spaceGUID},
	})
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	domainGUIDs := []string{}
	for _, route := range routes {
		domainGUIDs = append(domainGUIDs, route.DomainGUID)
	}

	domains, warnings, err := actor.CloudControllerClient.GetDomains(ccv3.Query{
		Key:    ccv3.GUIDFilter,
		Values: domainGUIDs,
	})
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	spacesByGUID := map[string]ccv3.Space{}
	for _, space := range spaces {
		spacesByGUID[space.GUID] = space
	}

	domainsByGUID := map[string]ccv3.Domain{}
	for _, domain := range domains {
		domainsByGUID[domain.GUID] = domain
	}

	actorRoutes := []Route{}
	for _, route := range routes {
		actorRoutes = append(actorRoutes, Route{
			GUID:       route.GUID,
			Host:       route.Host,
			Path:       route.Path,
			SpaceGUID:  route.SpaceGUID,
			DomainGUID: route.DomainGUID,
			SpaceName:  spacesByGUID[route.SpaceGUID].Name,
			DomainName: domainsByGUID[route.DomainGUID].Name,
		})
	}

	return actorRoutes, allWarnings, nil
}
