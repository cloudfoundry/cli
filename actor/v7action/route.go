package v7action

import (
	"sort"
	"strconv"
	"strings"

	"code.cloudfoundry.org/cli/v9/actor/actionerror"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/v9/resources"
	"code.cloudfoundry.org/cli/v9/util/batcher"
	"code.cloudfoundry.org/cli/v9/util/extract"
	"code.cloudfoundry.org/cli/v9/util/lookuptable"
	"code.cloudfoundry.org/cli/v9/util/railway"
	"code.cloudfoundry.org/cli/v9/util/sorting"
)

type RouteSummary struct {
	resources.Route
	AppNames            []string
	AppProtocols        []string
	DomainName          string
	SpaceName           string
	ServiceInstanceName string
}

func (actor Actor) CreateRoute(spaceGUID, domainName, hostname, path string, port int, options map[string]*string) (resources.Route, Warnings, error) {
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
		Options:    options,
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

func (actor Actor) GetRouteDestinationByAppGUID(route resources.Route, appGUID string) (resources.RouteDestination, error) {
	for _, destination := range route.Destinations {
		if destination.App.GUID == appGUID && destination.App.Process.Type == constant.ProcessTypeWeb {
			return destination, nil
		}
	}

	return resources.RouteDestination{}, actionerror.RouteDestinationNotFoundError{
		AppGUID:     appGUID,
		ProcessType: constant.ProcessTypeWeb,
		RouteGUID:   route.GUID,
	}
}

func (actor Actor) GetRoutesBySpace(spaceGUID string, labelSelector string) ([]resources.Route, Warnings, error) {
	allWarnings := Warnings{}
	queries := []ccv3.Query{
		{Key: ccv3.SpaceGUIDFilter, Values: []string{spaceGUID}},
		{Key: ccv3.PerPage, Values: []string{ccv3.MaxPerPage}},
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

func (actor Actor) parseRoutePath(routePath string) (string, string, string, resources.Domain, Warnings, error) {
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
			return "", "", "", resources.Domain{}, allWarnings, err
		}

		host = domainParts[0]
		domainName = domainParts[1]
		domain, warnings, err = actor.GetDomainByName(domainName)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return "", "", "", resources.Domain{}, allWarnings, err
		}
	}

	return host, path, port, domain, allWarnings, err
}

func (actor Actor) GetRoute(routePath string, spaceGUID string) (resources.Route, Warnings, error) {
	filters := []ccv3.Query{
		{Key: ccv3.SpaceGUIDFilter, Values: []string{spaceGUID}},
		{Key: ccv3.PerPage, Values: []string{"1"}},
		{Key: ccv3.Page, Values: []string{"1"}},
	}

	host, path, port, domain, allWarnings, err := actor.parseRoutePath(routePath)
	if err != nil {
		return resources.Route{}, allWarnings, err
	}

	filters = append(filters, ccv3.Query{Key: ccv3.DomainGUIDFilter,
		Values: []string{domain.GUID},
	})
	filters = append(filters, ccv3.Query{Key: ccv3.HostsFilter,
		Values: []string{host},
	})
	filters = append(filters, ccv3.Query{Key: ccv3.PathsFilter,
		Values: []string{path},
	})

	if domain.IsTCP() {
		filters = append(filters, ccv3.Query{Key: ccv3.PortsFilter,
			Values: []string{port},
		})
	}

	routes, warnings, err := actor.CloudControllerClient.GetRoutes(filters...)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return resources.Route{}, allWarnings, err
	}
	if len(routes) == 0 {
		return resources.Route{}, allWarnings, actionerror.RouteNotFoundError{
			Host:       host,
			DomainName: domain.Name,
			Path:       path,
		}
	}

	return routes[0], allWarnings, nil
}

func (actor Actor) GetRoutesByOrg(orgGUID string, labelSelector string) ([]resources.Route, Warnings, error) {
	allWarnings := Warnings{}
	queries := []ccv3.Query{
		{Key: ccv3.OrganizationGUIDFilter, Values: []string{orgGUID}},
		{Key: ccv3.PerPage, Values: []string{ccv3.MaxPerPage}},
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

func (actor Actor) GetApplicationMapForRoute(route resources.Route) (map[string]resources.Application, Warnings, error) {
	var v7Warning Warnings
	apps, v7Warning, err := actor.GetApplicationsByGUIDs(extract.UniqueList("Destinations.App.GUID", route))

	appMap := make(map[string]resources.Application)
	for _, a := range apps {
		appMap[a.GUID] = a
	}
	return appMap, v7Warning, err
}

func (actor Actor) GetRouteSummaries(routes []resources.Route) ([]RouteSummary, Warnings, error) {
	var (
		spaces           []resources.Space
		apps             []resources.Application
		routeBindings    []resources.RouteBinding
		serviceInstances []resources.ServiceInstance
	)

	warnings, err := railway.Sequentially(
		func() (ccv3.Warnings, error) {
			return batcher.RequestByGUID(
				extract.UniqueList("SpaceGUID", routes),
				func(guids []string) (ccv3.Warnings, error) {
					batch, _, warnings, err := actor.CloudControllerClient.GetSpaces(ccv3.Query{
						Key:    ccv3.GUIDFilter,
						Values: guids,
					})
					spaces = append(spaces, batch...)
					return warnings, err
				},
			)
		},
		func() (warnings ccv3.Warnings, err error) {
			var v7Warning Warnings
			apps, v7Warning, err = actor.GetApplicationsByGUIDs(extract.UniqueList("Destinations.App.GUID", routes))
			return ccv3.Warnings(v7Warning), err
		},
		func() (warnings ccv3.Warnings, err error) {
			return batcher.RequestByGUID(
				extract.UniqueList("GUID", routes),
				func(guids []string) (ccv3.Warnings, error) {
					batch, included, warnings, err := actor.CloudControllerClient.GetRouteBindings(
						ccv3.Query{Key: ccv3.Include, Values: []string{"service_instance"}},
						ccv3.Query{Key: ccv3.RouteGUIDFilter, Values: guids},
					)
					routeBindings = append(routeBindings, batch...)
					serviceInstances = append(serviceInstances, included.ServiceInstances...)
					return warnings, err
				},
			)
		},
	)

	if err != nil {
		return nil, Warnings(warnings), err
	}

	spaceNamesByGUID := lookuptable.NameFromGUID(spaces)
	appNamesByGUID := lookuptable.NameFromGUID(apps)
	serviceInstanceNameByGUID := lookuptable.NameFromGUID(serviceInstances)

	serviceInstanceNameByRouteGUID := make(map[string]string)
	for _, routeBinding := range routeBindings {
		serviceInstanceNameByRouteGUID[routeBinding.RouteGUID] = serviceInstanceNameByGUID[routeBinding.ServiceInstanceGUID]
	}

	var routeSummaries []RouteSummary
	for _, route := range routes {
		var appNames []string

		protocolSet := map[string]bool{}
		for _, destination := range route.Destinations {
			appNames = append(appNames, appNamesByGUID[destination.App.GUID])
			protocolSet[destination.Protocol] = true
		}

		var appProtocols []string
		if len(protocolSet) > 0 {
			appProtocols = make([]string, 0, len(protocolSet))
			for key := range protocolSet {
				appProtocols = append(appProtocols, key)
			}
			sort.Strings(appProtocols)
		}

		routeSummaries = append(routeSummaries, RouteSummary{
			Route:               route,
			AppNames:            appNames,
			AppProtocols:        appProtocols,
			SpaceName:           spaceNamesByGUID[route.SpaceGUID],
			DomainName:          getDomainName(route.URL, route.Host, route.Path, route.Port),
			ServiceInstanceName: serviceInstanceNameByRouteGUID[route.GUID],
		})
	}

	sort.Slice(routeSummaries, func(i, j int) bool {
		return sorting.LessIgnoreCase(routeSummaries[i].SpaceName, routeSummaries[j].SpaceName)
	})

	return routeSummaries, Warnings(warnings), nil
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

func (actor Actor) DeleteRoute(domainName, hostname, path string, port int) (Warnings, error) {
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

	route, actorWarnings, err := actor.GetRouteByAttributes(domain, hostname, path, port)

	allWarnings = append(allWarnings, actorWarnings...)

	if err != nil {
		return allWarnings, err
	}

	jobURL, apiWarnings, err := actor.CloudControllerClient.DeleteRoute(route.GUID)
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

	queries := []ccv3.Query{
		{Key: ccv3.DomainGUIDFilter, Values: []string{domain.GUID}},
		{Key: ccv3.HostsFilter, Values: []string{hostname}},
		{Key: ccv3.PathsFilter, Values: []string{path}},
		{Key: ccv3.PerPage, Values: []string{"1"}},
		{Key: ccv3.Page, Values: []string{"1"}},
	}

	if domain.IsTCP() {
		queries = append(queries, ccv3.Query{Key: ccv3.PortsFilter, Values: []string{strconv.Itoa(port)}})
	}

	routes, ccWarnings, err = actor.CloudControllerClient.GetRoutes(queries...)
	if err != nil {
		return resources.Route{}, Warnings(ccWarnings), err
	}

	if len(routes) < 1 {
		return resources.Route{}, Warnings(ccWarnings), actionerror.RouteNotFoundError{
			DomainName: domain.Name,
			Host:       hostname,
			Path:       path,
			Port:       port,
		}
	}

	return routes[0], Warnings(ccWarnings), nil
}

func (actor Actor) MapRoute(routeGUID string, appGUID string, destinationProtocol string) (Warnings, error) {
	warnings, err := actor.CloudControllerClient.MapRoute(routeGUID, appGUID, destinationProtocol)
	return Warnings(warnings), err
}

func (actor Actor) UpdateRoute(routeGUID string, options map[string]*string) (resources.Route, Warnings, error) {
	route, warnings, err := actor.CloudControllerClient.UpdateRoute(routeGUID, options)
	return route, Warnings(warnings), err
}

func (actor Actor) UpdateDestination(routeGUID string, destinationGUID string, protocol string) (Warnings, error) {
	warnings, err := actor.CloudControllerClient.UpdateDestination(routeGUID, destinationGUID, protocol)
	return Warnings(warnings), err
}
func (actor Actor) UnmapRoute(routeGUID string, destinationGUID string) (Warnings, error) {
	warnings, err := actor.CloudControllerClient.UnmapRoute(routeGUID, destinationGUID)
	return Warnings(warnings), err
}
func (actor Actor) ShareRoute(routeGUID string, spaceGUID string) (Warnings, error) {
	warnings, err := actor.CloudControllerClient.ShareRoute(routeGUID, spaceGUID)
	return Warnings(warnings), err
}

func (actor Actor) UnshareRoute(routeGUID string, spaceGUID string) (Warnings, error) {
	warnings, err := actor.CloudControllerClient.UnshareRoute(routeGUID, spaceGUID)
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

func (actor Actor) MoveRoute(routeGUID string, spaceGUID string) (Warnings, error) {
	warnings, err := actor.CloudControllerClient.MoveRoute(routeGUID, spaceGUID)
	return Warnings(warnings), err
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
