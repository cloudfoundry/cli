package v2action

import (
	"fmt"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
)

// Route represents a CLI Route.
type Route struct {
	GUID   string
	Host   string
	Domain string
	Path   string
	Port   int
}

// String formats the route in a human readable format.
func (r Route) String() string {
	routeString := r.Domain

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

// GetOrphanedRoutesBySpace returns a list of orphaned routes associated with
// the provided Space GUID.
func (actor Actor) GetOrphanedRoutesBySpace(spaceGUID string) ([]Route, Warnings, error) {
	var (
		orphanedRoutes []Route
		allWarnings    Warnings
	)

	routes, warnings, err := actor.GetSpaceRoutes(spaceGUID, nil)
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

// GetApplicationRoutes returns a list of routes associated with the provided Application GUID
func (actor Actor) GetApplicationRoutes(applicationGUID string, query []ccv2.Query) ([]Route, Warnings, error) {
	var allWarnings Warnings
	ccv2Routes, warnings, err := actor.CloudControllerClient.GetApplicationRoutes(applicationGUID, query)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	routes, domainWarnings, err := actor.applyDomain(ccv2Routes)

	return routes, append(allWarnings, domainWarnings...), err
}

// GetSpaceRoutes returns a list of routes associated with the provided Space GUID
func (actor Actor) GetSpaceRoutes(spaceGUID string, query []ccv2.Query) ([]Route, Warnings, error) {
	var allWarnings Warnings
	ccv2Routes, warnings, err := actor.CloudControllerClient.GetSpaceRoutes(spaceGUID, query)
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

func (actor Actor) applyDomain(ccv2Routes []ccv2.Route) ([]Route, Warnings, error) {
	var routes []Route
	var allWarnings Warnings

	for _, ccv2Route := range ccv2Routes {
		domain, warnings, err := actor.GetDomain(ccv2Route.DomainGUID)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return nil, allWarnings, err
		}
		routes = append(routes, Route{
			GUID:   ccv2Route.GUID,
			Host:   ccv2Route.Host,
			Domain: domain.Name,
			Path:   ccv2Route.Path,
			Port:   ccv2Route.Port,
		})
	}

	return routes, allWarnings, nil
}
