package v2action

import (
	"fmt"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
)

// Route represents a CLI Route.
type Route struct {
	Domain    Domain
	GUID      string
	Host      string
	Path      string
	Port      int
	SpaceGUID string
}

// String formats the route in a human readable format.
func (r Route) String() string {
	routeString := r.Domain.Name

	if r.Port != 0 {
		routeString = fmt.Sprintf("%s:%d", routeString, r.Port)
		return routeString
	}

	if r.Host != "" {
		routeString = fmt.Sprintf("%s.%s", r.Host, routeString)
	}

	if r.Path != "" {
		routeString = fmt.Sprintf("%s%s", routeString, r.Path)
	}

	return routeString
}

// OrphanedRoutesNotFoundError is an error wrapper that represents the case
// when no orphaned routes are found.
type OrphanedRoutesNotFoundError struct{}

// Error method to display the error message.
func (e OrphanedRoutesNotFoundError) Error() string {
	return fmt.Sprintf("No orphaned routes were found.")
}

// RouteNotFoundError is returned when a route cannot be found
type RouteNotFoundError struct {
	Host       string
	DomainGUID string
}

func (e RouteNotFoundError) Error() string {
	return fmt.Sprintf("Route with host %s and domain guid %s not found", e.Host, e.DomainGUID)
}

// RouteInDifferentSpaceError is returned when the route exists in a different
// space than the one requesting it.
type RouteInDifferentSpaceError struct {
	Route string
}

func (e RouteInDifferentSpaceError) Error() string {
	return fmt.Sprintf("route registered to another space")
}

func (actor Actor) BindRouteToApplication(routeGUID string, appGUID string) (Warnings, error) {
	_, warnings, err := actor.CloudControllerClient.BindRouteToApplication(routeGUID, appGUID)
	if _, ok := err.(ccerror.InvalidRelationError); ok {
		return Warnings(warnings), RouteInDifferentSpaceError{}
	}
	return Warnings(warnings), err
}

func (actor Actor) CreateRoute(route Route, generatePort bool) (Route, Warnings, error) {
	returnedRoute, warnings, err := actor.CloudControllerClient.CreateRoute(actorToCCRoute(route), generatePort)
	return ccToActorRoute(returnedRoute, route.Domain), Warnings(warnings), err
}

// GetOrphanedRoutesBySpace returns a list of orphaned routes associated with
// the provided Space GUID.
func (actor Actor) GetOrphanedRoutesBySpace(spaceGUID string) ([]Route, Warnings, error) {
	var (
		orphanedRoutes []Route
		allWarnings    Warnings
	)

	routes, warnings, err := actor.GetSpaceRoutes(spaceGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	for _, route := range routes {
		apps, warnings, err := actor.GetRouteApplications(route.GUID, nil)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return nil, allWarnings, err
		}

		if len(apps) == 0 {
			orphanedRoutes = append(orphanedRoutes, route)
		}
	}

	if len(orphanedRoutes) == 0 {
		return nil, allWarnings, OrphanedRoutesNotFoundError{}
	}

	return orphanedRoutes, allWarnings, nil
}

// GetApplicationRoutes returns a list of routes associated with the provided
// Application GUID.
func (actor Actor) GetApplicationRoutes(applicationGUID string) ([]Route, Warnings, error) {
	var allWarnings Warnings
	ccv2Routes, warnings, err := actor.CloudControllerClient.GetApplicationRoutes(applicationGUID, nil)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	routes, domainWarnings, err := actor.applyDomain(ccv2Routes)

	return routes, append(allWarnings, domainWarnings...), err
}

// GetSpaceRoutes returns a list of routes associated with the provided Space
// GUID.
func (actor Actor) GetSpaceRoutes(spaceGUID string) ([]Route, Warnings, error) {
	var allWarnings Warnings
	ccv2Routes, warnings, err := actor.CloudControllerClient.GetSpaceRoutes(spaceGUID, nil)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	routes, domainWarnings, err := actor.applyDomain(ccv2Routes)

	return routes, append(allWarnings, domainWarnings...), err
}

// DeleteRoute deletes the Route associated with the provided Route GUID.
func (actor Actor) DeleteRoute(routeGUID string) (Warnings, error) {
	warnings, err := actor.CloudControllerClient.DeleteRoute(routeGUID)
	return Warnings(warnings), err
}

// GetRouteByHostAndDomain returns the HTTP route with the matching host and
// the associate domain GUID.
func (actor Actor) GetRouteByHostAndDomain(host string, domainGUID string) (Route, Warnings, error) {
	ccv2Routes, warnings, err := actor.CloudControllerClient.GetRoutes([]ccv2.Query{
		{Filter: ccv2.HostFilter, Operator: ccv2.EqualOperator, Value: host},
		{Filter: ccv2.DomainGUIDFilter, Operator: ccv2.EqualOperator, Value: domainGUID},
	})
	if err != nil {
		return Route{}, Warnings(warnings), err
	}

	if len(ccv2Routes) == 0 {
		return Route{}, Warnings(warnings), RouteNotFoundError{Host: host, DomainGUID: domainGUID}
	}

	routes, domainWarnings, err := actor.applyDomain(ccv2Routes)
	if err != nil {
		return Route{}, append(Warnings(warnings), domainWarnings...), err
	}

	return routes[0], append(Warnings(warnings), domainWarnings...), err
}

func (actor Actor) CheckRoute(route Route) (bool, Warnings, error) {
	exists, warnings, err := actor.CloudControllerClient.CheckRoute(actorToCCRoute(route))
	return exists, Warnings(warnings), err
}

func (actor Actor) applyDomain(ccv2Routes []ccv2.Route) ([]Route, Warnings, error) {
	var routes []Route
	var allWarnings Warnings

	for _, ccv2Route := range ccv2Routes {
		domain, warnings, err := actor.GetDomain(ccv2Route.DomainGUID)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return nil, allWarnings, err
		}
		routes = append(routes, ccToActorRoute(ccv2Route, domain))
	}

	return routes, allWarnings, nil
}

func actorToCCRoute(route Route) ccv2.Route {
	return ccv2.Route{
		DomainGUID: route.Domain.GUID,
		Host:       route.Host,
		Path:       route.Path,
		Port:       route.Port,
		SpaceGUID:  route.SpaceGUID,
	}
}

func ccToActorRoute(ccv2Route ccv2.Route, domain Domain) Route {
	return Route{
		Domain:    domain,
		GUID:      ccv2Route.GUID,
		Host:      ccv2Route.Host,
		Path:      ccv2Route.Path,
		Port:      ccv2Route.Port,
		SpaceGUID: ccv2Route.SpaceGUID,
	}
}
