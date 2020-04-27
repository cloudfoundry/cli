package v7action

import (
	"sort"
	"strings"

	"code.cloudfoundry.org/cli/resources"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/util/sorting"
)

type RouteSummary struct {
	resources.Route
	AppNames   []string
	DomainName string
	SpaceName  string
}

func (actor Actor) CreateRoute(spaceGUID, domainName, hostname, path string) (resources.Route, Warnings, error) {
	allWarnings := Warnings{}
	domain, warnings, err := actor.GetDomainByName(domainName)
	allWarnings = append(allWarnings, warnings...)

	if err != nil {
		return resources.Route{}, allWarnings, err
	}

	route, apiWarnings, err := actor.CloudControllerClient.CreateRoute(resources.Route{
		SpaceGUID:  spaceGUID,
		DomainGUID: domain.GUID,
		Host:       hostname,
		Path:       path,
	})

	actorWarnings := Warnings(apiWarnings)
	allWarnings = append(allWarnings, actorWarnings...)

	if _, ok := err.(ccerror.RouteNotUniqueError); ok {
		return resources.Route{}, allWarnings, actionerror.RouteAlreadyExistsError{Err: err}
	}

	return resources.Route{
		GUID:       route.GUID,
		Host:       route.Host,
		Path:       route.Path,
		SpaceGUID:  route.SpaceGUID,
		DomainGUID: route.DomainGUID,
	}, allWarnings, err
}

func (actor Actor) GetRouteDestinations(routeGUID string) ([]resources.RouteDestination, Warnings, error) {
	destinations, warnings, err := actor.CloudControllerClient.GetRouteDestinations(routeGUID)

	var actorDestinations []resources.RouteDestination
	for _, dst := range destinations {
		actorDestinations = append(actorDestinations, resources.RouteDestination{
			GUID: dst.GUID,
			App:  resources.RouteDestinationApp(dst.App),
		})
	}

	return actorDestinations, Warnings(warnings), err
}

func (actor Actor) GetRouteDestinationByAppGUID(routeGUID string, appGUID string) (resources.RouteDestination, Warnings, error) {
	allDestinations, warnings, err := actor.GetRouteDestinations(routeGUID)
	if err != nil {
		return resources.RouteDestination{}, warnings, err
	}

	for _, destination := range allDestinations {
		if destination.App.GUID == appGUID && destination.App.Process.Type == constant.ProcessTypeWeb {
			return destination, warnings, nil
		}
	}

	return resources.RouteDestination{}, warnings, actionerror.RouteDestinationNotFoundError{
		AppGUID:     appGUID,
		ProcessType: constant.ProcessTypeWeb,
		RouteGUID:   routeGUID,
	}
}

func (actor Actor) GetRoutesBySpace(spaceGUID string, labelSelector string) ([]resources.Route, Warnings, error) {
	allWarnings := Warnings{}
	queries := []ccv3.Query{
		ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{spaceGUID}},
	}
	if len(labelSelector) > 0 {
		queries = append(queries, ccv3.Query{Key: ccv3.LabelSelectorFilter, Values: []string{labelSelector}})
	}

	routes, warnings, err := actor.CloudControllerClient.GetRoutes(queries...)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	return routes, allWarnings, nil
}

func (actor Actor) parseRoutePath(routePath string) (string, string, string, string, Warnings, error) {
	var warnings Warnings
	var hostPart = ""
	var pathPart = ""
	routeParts := strings.SplitN(routePath, "/", 2)
	domainName := routeParts[0]
	if len(routeParts) > 1 {
		pathPart = "/" + routeParts[1]
	}
	domainParts := strings.SplitN(domainName, ".", 2)
	domainHasMultipleParts := len(domainParts) > 1

	domain, allWarnings, err := actor.GetDomainByName(domainName)

	_, domainNotFound := err.(actionerror.DomainNotFoundError)

	needToCheckForHost := domainNotFound && domainHasMultipleParts
	if err != nil && !needToCheckForHost {
		return "", "", "", "", allWarnings, err
	}

	if needToCheckForHost {
		domainName = domainParts[1]
		hostPart = domainParts[0]
		domain, warnings, err = actor.GetDomainByName(domainName)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return "", "", "", "", allWarnings, err
		}
	}

	return hostPart, pathPart, domainName, domain.GUID, allWarnings, nil
}

func (actor Actor) GetRoute(routePath string, spaceGUID string) (resources.Route, Warnings, error) {
	filters := []ccv3.Query{
		ccv3.Query{
			Key:    ccv3.SpaceGUIDFilter,
			Values: []string{spaceGUID},
		},
	}
	hostPart, pathPart, domainName, domainGUID, allWarnings, err := actor.parseRoutePath(routePath)
	if err != nil {
		return resources.Route{}, allWarnings, err
	}
	filters = append(filters, ccv3.Query{
		Key:    ccv3.DomainGUIDFilter,
		Values: []string{domainGUID},
	})
	filters = append(filters, ccv3.Query{Key: ccv3.HostsFilter,
		Values: []string{hostPart},
	})
	filters = append(filters, ccv3.Query{Key: ccv3.PathsFilter,
		Values: []string{pathPart},
	})
	routes, warnings, err := actor.CloudControllerClient.GetRoutes(filters...)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return resources.Route{}, allWarnings, err
	}
	if len(routes) == 0 {
		return resources.Route{}, allWarnings, actionerror.RouteNotFoundError{
			Host:       hostPart,
			DomainName: domainName,
			Path:       pathPart,
		}
	}

	return routes[0], allWarnings, nil
}

func (actor Actor) GetRoutesByOrg(orgGUID string, labelSelector string) ([]resources.Route, Warnings, error) {
	allWarnings := Warnings{}
	queries := []ccv3.Query{
		ccv3.Query{Key: ccv3.OrganizationGUIDFilter, Values: []string{orgGUID}},
	}
	if len(labelSelector) > 0 {
		queries = append(queries, ccv3.Query{Key: ccv3.LabelSelectorFilter, Values: []string{labelSelector}})
	}

	routes, warnings, err := actor.CloudControllerClient.GetRoutes(queries...)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	return routes, allWarnings, nil
}

func (actor Actor) GetRouteSummaries(routes []resources.Route) ([]RouteSummary, Warnings, error) {
	var allWarnings Warnings
	var routeSummaries []RouteSummary

	destinationAppGUIDsByRouteGUID := make(map[string][]string)
	destinationAppGUIDs := make(map[string]bool)
	var uniqueAppGUIDs []string

	spaceGUIDs := make(map[string]struct{})
	var uniqueSpaceGUIDs []string

	for _, route := range routes {
		if _, seen := spaceGUIDs[route.SpaceGUID]; !seen {
			spaceGUIDs[route.SpaceGUID] = struct{}{}
			uniqueSpaceGUIDs = append(uniqueSpaceGUIDs, route.SpaceGUID)
		}

		for _, destination := range route.Destinations {
			appGUID := destination.App.GUID

			if _, ok := destinationAppGUIDs[appGUID]; !ok {
				destinationAppGUIDs[appGUID] = true
				uniqueAppGUIDs = append(uniqueAppGUIDs, appGUID)
			}

			destinationAppGUIDsByRouteGUID[route.GUID] = append(destinationAppGUIDsByRouteGUID[route.GUID], appGUID)
		}
	}

	spaces, _, ccv3Warnings, err := actor.CloudControllerClient.GetSpaces(ccv3.Query{
		Key:    ccv3.GUIDFilter,
		Values: uniqueSpaceGUIDs,
	})
	allWarnings = append(allWarnings, ccv3Warnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	spaceNamesByGUID := make(map[string]string)
	for _, space := range spaces {
		spaceNamesByGUID[space.GUID] = space.Name
	}

	apps, warnings, err := actor.GetApplicationsByGUIDs(uniqueAppGUIDs)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	appNamesByGUID := make(map[string]string)
	for _, app := range apps {
		appNamesByGUID[app.GUID] = app.Name
	}

	for _, route := range routes {
		var appNames []string

		appGUIDs := destinationAppGUIDsByRouteGUID[route.GUID]
		for _, appGUID := range appGUIDs {
			appNames = append(appNames, appNamesByGUID[appGUID])
		}

		routeSummaries = append(routeSummaries, RouteSummary{
			Route:      route,
			AppNames:   appNames,
			SpaceName:  spaceNamesByGUID[route.SpaceGUID],
			DomainName: getDomainName(route.URL, route.Host, route.Path),
		})
	}

	sort.Slice(routeSummaries, func(i, j int) bool {
		return sorting.LessIgnoreCase(routeSummaries[i].SpaceName, routeSummaries[j].SpaceName)
	})

	return routeSummaries, allWarnings, nil
}

func (actor Actor) DeleteOrphanedRoutes(spaceGUID string) (Warnings, error) {
	var allWarnings Warnings

	jobURL, warnings, err := actor.CloudControllerClient.DeleteOrphanedRoutes(spaceGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	warnings, err = actor.CloudControllerClient.PollJob(jobURL)
	allWarnings = append(allWarnings, warnings...)

	return allWarnings, err
}

func (actor Actor) DeleteRoute(domainName, hostname, path string) (Warnings, error) {
	allWarnings := Warnings{}
	domain, warnings, err := actor.GetDomainByName(domainName)
	allWarnings = append(allWarnings, warnings...)

	if _, ok := err.(actionerror.DomainNotFoundError); ok {
		allWarnings = append(allWarnings, err.Error())
		return allWarnings, nil
	}

	if err != nil {
		return allWarnings, err
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

func (actor Actor) GetRouteByAttributes(domainName string, domainGUID string, hostname string, path string) (resources.Route, Warnings, error) {
	routes, ccWarnings, err := actor.CloudControllerClient.GetRoutes(
		ccv3.Query{Key: ccv3.DomainGUIDFilter, Values: []string{domainGUID}},
		ccv3.Query{Key: ccv3.HostsFilter, Values: []string{hostname}},
		ccv3.Query{Key: ccv3.PathsFilter, Values: []string{path}},
	)

	if err != nil {
		return resources.Route{}, Warnings(ccWarnings), err
	}

	if len(routes) < 1 {
		return resources.Route{}, Warnings(ccWarnings), actionerror.RouteNotFoundError{
			DomainName: domainName,
			DomainGUID: domainGUID,
			Host:       hostname,
			Path:       path,
		}
	}

	return resources.Route{
		GUID:       routes[0].GUID,
		Host:       routes[0].Host,
		Path:       routes[0].Path,
		SpaceGUID:  routes[0].SpaceGUID,
		DomainGUID: routes[0].DomainGUID,
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

func (actor Actor) GetApplicationRoutes(appGUID string) ([]resources.Route, Warnings, error) {
	allWarnings := Warnings{}

	routes, warnings, err := actor.CloudControllerClient.GetApplicationRoutes(appGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	if len(routes) == 0 {
		return nil, allWarnings, err
	}

	return routes, allWarnings, nil
}

func getDomainName(fullURL, host, path string) string {
	domainWithoutHost := strings.TrimPrefix(fullURL, host+".")
	return strings.TrimSuffix(domainWithoutHost, path)
}
