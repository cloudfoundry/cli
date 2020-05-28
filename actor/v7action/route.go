package v7action

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/sorting"
)

type RouteSummary struct {
	resources.Route
	AppNames   []string
	DomainName string
	SpaceName  string
}

func (actor Actor) CreateRoute(spaceGUID, domainName, hostname, path string, port int) (resources.Route, Warnings, error) {
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
		Port:       port,
	})

	actorWarnings := Warnings(apiWarnings)
	allWarnings = append(allWarnings, actorWarnings...)

	if _, ok := err.(ccerror.RouteNotUniqueError); ok {
		return resources.Route{}, allWarnings, actionerror.RouteAlreadyExistsError{Err: err}
	}

	return route, allWarnings, err
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

func (actor Actor) parseRoutePath(routePath string) (string, string, string, string, string, Warnings, error) {
	var (
		warnings Warnings
		host     string
		path     string
		port     string
	)

	routeParts := strings.SplitN(routePath, "/", 2)
	domainAndPort := routeParts[0]
	if len(routeParts) > 1 {
		path = "/" + routeParts[1]
	}

	domainAndPortParts := strings.SplitN(domainAndPort, ":", 2)
	domainName := domainAndPortParts[0]
	if len(domainAndPortParts) > 1 {
		port = domainAndPortParts[1]
	}

	domain, allWarnings, err := actor.GetDomainByName(domainName)

	if err != nil {
		_, domainNotFound := err.(actionerror.DomainNotFoundError)
		domainParts := strings.SplitN(domainName, ".", 2)
		needToCheckForHost := domainNotFound && len(domainParts) > 1

		if !needToCheckForHost {
			return "", "", "", "", "", allWarnings, err
		}

		host = domainParts[0]
		domainName = domainParts[1]
		domain, warnings, err = actor.GetDomainByName(domainName)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return "", "", "", "", "", allWarnings, err
		}
	}

	return host, path, domainName, domain.GUID, port, allWarnings, err
}

func (actor Actor) GetRoute(routePath string, spaceGUID string) (resources.Route, Warnings, error) {
	filters := []ccv3.Query{
		{
			Key:    ccv3.SpaceGUIDFilter,
			Values: []string{spaceGUID},
		},
	}
	host, path, domainName, domainGUID, port, allWarnings, err := actor.parseRoutePath(routePath)
	if err != nil {
		return resources.Route{}, allWarnings, err
	}
	filters = append(filters, ccv3.Query{
		Key:    ccv3.DomainGUIDFilter,
		Values: []string{domainGUID},
	})
	filters = append(filters, ccv3.Query{Key: ccv3.HostsFilter,
		Values: []string{host},
	})
	filters = append(filters, ccv3.Query{Key: ccv3.PathsFilter,
		Values: []string{path},
	})
	filters = append(filters, ccv3.Query{Key: ccv3.PortsFilter,
		Values: []string{port},
	})
	routes, warnings, err := actor.CloudControllerClient.GetRoutes(filters...)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return resources.Route{}, allWarnings, err
	}
	if len(routes) == 0 {
		return resources.Route{}, allWarnings, actionerror.RouteNotFoundError{
			Host:       host,
			DomainName: domainName,
			Path:       path,
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
			DomainName: getDomainName(route.URL, route.Host, route.Path, route.Port),
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

func (actor Actor) GetRouteByAttributes(domain resources.Domain, hostname string, path string, port int) (resources.Route, Warnings, error) {
	var (
		routes     []resources.Route
		ccWarnings ccv3.Warnings
		err        error
	)

	if domain.HasProtocolType("tcp") && port == 0 {
		return resources.Route{}, Warnings(ccWarnings), actionerror.TCPLookupWithoutPort{}
	}

	if port != 0 {
		routes, ccWarnings, err = actor.CloudControllerClient.GetRoutes(
			ccv3.Query{Key: ccv3.DomainGUIDFilter, Values: []string{domain.GUID}},
			ccv3.Query{Key: ccv3.PortsFilter, Values: []string{fmt.Sprintf("%d", port)}},
		)
	} else {
		routes, ccWarnings, err = actor.CloudControllerClient.GetRoutes(
			ccv3.Query{Key: ccv3.DomainGUIDFilter, Values: []string{domain.GUID}},
			ccv3.Query{Key: ccv3.HostsFilter, Values: []string{hostname}},
			ccv3.Query{Key: ccv3.PathsFilter, Values: []string{path}},
		)
	}

	if err != nil {
		return resources.Route{}, Warnings(ccWarnings), err
	}

	if len(routes) < 1 {
		return resources.Route{}, Warnings(ccWarnings), actionerror.RouteNotFoundError{
			DomainName: domain.Name,
			DomainGUID: domain.GUID,
			Host:       hostname,
			Path:       path,
			Port:       port,
		}
	}

	return routes[0], Warnings(ccWarnings), nil
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

func getDomainName(fullURL, host, path string, port int) string {
	domainWithoutHost := strings.TrimPrefix(fullURL, host+".")
	domainWithoutPath := strings.TrimSuffix(domainWithoutHost, path)

	if port > 0 {
		portString := strconv.Itoa(port)
		domainWithoutPort := strings.TrimSuffix(domainWithoutPath, ":"+portString)
		return domainWithoutPort
	}

	return domainWithoutPath
}
