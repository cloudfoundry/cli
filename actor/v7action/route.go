package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
)

type RouteDestination struct {
	GUID string
	App  RouteDestinationApp
}
type RouteDestinationApp ccv3.RouteDestinationApp

type Route struct {
	GUID       string
	SpaceGUID  string
	DomainGUID string
	Host       string
	Path       string
	DomainName string
	SpaceName  string
}

func (actor Actor) CreateRoute(orgName, spaceName, domainName, hostname, path string) (Route, Warnings, error) {
	allWarnings := Warnings{}
	domain, warnings, err := actor.GetDomainByName(domainName)
	allWarnings = append(allWarnings, warnings...)

	if err != nil {
		return Route{}, allWarnings, err
	}

	org, warnings, err := actor.GetOrganizationByName(orgName)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return Route{}, allWarnings, err
	}

	space, warnings, err := actor.GetSpaceByNameAndOrganization(spaceName, org.GUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return Route{}, allWarnings, err
	}

	if path != "" && string(path[0]) != "/" {
		path = "/" + path
	}
	route, apiWarnings, err := actor.CloudControllerClient.CreateRoute(ccv3.Route{
		SpaceGUID:  space.GUID,
		DomainGUID: domain.GUID,
		Host:       hostname,
		Path:       path,
	})

	actorWarnings := Warnings(apiWarnings)
	allWarnings = append(allWarnings, actorWarnings...)

	if _, ok := err.(ccerror.RouteNotUniqueError); ok {
		return Route{}, allWarnings, actionerror.RouteAlreadyExistsError{Err: err}
	}

	return Route{
		GUID:       route.GUID,
		Host:       route.Host,
		Path:       route.Path,
		SpaceGUID:  route.SpaceGUID,
		DomainGUID: route.DomainGUID,
		SpaceName:  spaceName,
		DomainName: domainName,
	}, allWarnings, err
}

func (actor Actor) GetRouteDestinations(routeGUID string) ([]RouteDestination, Warnings, error) {
	destinations, warnings, err := actor.CloudControllerClient.GetRouteDestinations(routeGUID)

	actorDestinations := []RouteDestination{}
	for _, dst := range destinations {
		actorDestinations = append(actorDestinations, RouteDestination{
			GUID: dst.GUID,
			App:  RouteDestinationApp(dst.App),
		})
	}

	return actorDestinations, Warnings(warnings), err
}

func (actor Actor) GetRouteDestinationByAppGUID(routeGUID string, appGUID string) (RouteDestination, Warnings, error) {
	allDestinations, warnings, err := actor.GetRouteDestinations(routeGUID)
	if err != nil {
		return RouteDestination{}, warnings, err
	}

	for _, destination := range allDestinations {
		if destination.App.GUID == appGUID && destination.App.Process.Type == constant.ProcessTypeWeb {
			return destination, warnings, nil
		}
	}

	return RouteDestination{}, warnings, actionerror.RouteDestinationNotFoundError{
		AppGUID:     appGUID,
		ProcessType: constant.ProcessTypeWeb,
		RouteGUID:   routeGUID,
	}
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

	domainGUIDsSet := map[string]struct{}{}
	domainGUIDs := []string{}
	for _, route := range routes {
		if _, ok := domainGUIDsSet[route.DomainGUID]; ok {
			continue
		}
		domainGUIDsSet[route.DomainGUID] = struct{}{}
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

func (actor Actor) GetRoutesByOrg(orgGUID string) ([]Route, Warnings, error) {
	allWarnings := Warnings{}

	routes, warnings, err := actor.CloudControllerClient.GetRoutes(ccv3.Query{
		Key:    ccv3.OrganizationGUIDFilter,
		Values: []string{orgGUID},
	})
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	spaceGUIDsSet := map[string]struct{}{}
	domainGUIDsSet := map[string]struct{}{}
	spacesQuery := ccv3.Query{Key: ccv3.GUIDFilter, Values: []string{}}
	domainsQuery := ccv3.Query{Key: ccv3.GUIDFilter, Values: []string{}}

	for _, route := range routes {
		if _, ok := spaceGUIDsSet[route.SpaceGUID]; !ok {
			spacesQuery.Values = append(spacesQuery.Values, route.SpaceGUID)
			spaceGUIDsSet[route.SpaceGUID] = struct{}{}
		}

		if _, ok := domainGUIDsSet[route.DomainGUID]; !ok {
			domainsQuery.Values = append(domainsQuery.Values, route.DomainGUID)
			domainGUIDsSet[route.DomainGUID] = struct{}{}
		}
	}

	spaces, warnings, err := actor.CloudControllerClient.GetSpaces(spacesQuery)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	domains, warnings, err := actor.CloudControllerClient.GetDomains(domainsQuery)
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

func (actor Actor) DeleteRoute(domainName, hostname, path string) (Warnings, error) {
	allWarnings := Warnings{}
	domain, warnings, err := actor.GetDomainByName(domainName)
	allWarnings = append(allWarnings, warnings...)

	if err != nil {
		return allWarnings, err
	}

	if path != "" && string(path[0]) != "/" {
		path = "/" + path
	}

	queryArray := []ccv3.Query{
		{Key: ccv3.DomainGUIDFilter, Values: []string{domain.GUID}},
		{Key: ccv3.HostsFilter, Values: []string{hostname}},
		{Key: ccv3.PathsFilter, Values: []string{path}},
	}

	routes, apiWarnings, err := actor.CloudControllerClient.GetRoutes(queryArray...)

	actorWarnings := Warnings(apiWarnings)
	allWarnings = append(allWarnings, actorWarnings...)

	if err != nil {
		return allWarnings, err
	}

	if len(routes) == 0 {
		return allWarnings, actionerror.RouteNotFoundError{
			DomainName: domainName,
			Host:       hostname,
			Path:       path,
		}
	}

	jobURL, apiWarnings, err := actor.CloudControllerClient.DeleteRoute(routes[0].GUID)
	actorWarnings = Warnings(apiWarnings)
	allWarnings = append(allWarnings, actorWarnings...)

	if err != nil {
		return allWarnings, err
	}

	pollJobWarnings, err := actor.CloudControllerClient.PollJob(jobURL)
	allWarnings = append(allWarnings, Warnings(pollJobWarnings)...)

	return allWarnings, err
}

func (actor Actor) GetRouteByAttributes(domainName string, domainGUID string, hostname string, path string) (Route, Warnings, error) {
	if path != "" && string(path[0]) != "/" {
		path = "/" + path
	}

	ccRoutes, ccWarnings, err := actor.CloudControllerClient.GetRoutes(
		ccv3.Query{Key: ccv3.DomainGUIDFilter, Values: []string{domainGUID}},
		ccv3.Query{Key: ccv3.HostsFilter, Values: []string{hostname}},
		ccv3.Query{Key: ccv3.PathsFilter, Values: []string{path}},
	)

	if err != nil {
		return Route{}, Warnings(ccWarnings), err
	}

	if len(ccRoutes) < 1 {
		return Route{}, Warnings(ccWarnings), actionerror.RouteNotFoundError{
			DomainName: domainName,
			DomainGUID: domainGUID,
			Host:       hostname,
			Path:       path,
		}
	}

	return Route{
		GUID:       ccRoutes[0].GUID,
		Host:       ccRoutes[0].Host,
		Path:       ccRoutes[0].Path,
		SpaceGUID:  ccRoutes[0].SpaceGUID,
		DomainGUID: ccRoutes[0].DomainGUID,
	}, Warnings(ccWarnings), nil
}

func (actor Actor) MapRoute(routeGUID string, appGUID string) (Warnings, error) {
	warnings, err := actor.CloudControllerClient.MapRoute(routeGUID, appGUID)
	return Warnings(warnings), err
}

func (actor Actor) UnmapRoute(routeGUID string, destinationGUID string) (Warnings, error) {
	warnings, err := actor.CloudControllerClient.UnmapRoute(routeGUID, destinationGUID)
	return Warnings(warnings), err
}
