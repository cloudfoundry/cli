package v2actions

import (
	"fmt"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
)

// Route represents a CLI Route.
type Route ccv2.Route

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

	routes, warnings, err := actor.CloudControllerClient.GetSpaceRoutes(spaceGUID, nil)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	for _, route := range routes {
		apps, warnings, err := actor.CloudControllerClient.GetRouteApplications(route.GUID, nil)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return nil, allWarnings, err
		}

		if len(apps) == 0 {
			domain, warnings, err := actor.GetDomainByGUID(route.DomainFields.GUID)
			allWarnings = append(allWarnings, warnings...)
			if err != nil {
				return nil, allWarnings, err
			}

			orphanedRoutes = append(orphanedRoutes, Route{
				GUID:   route.GUID,
				Host:   route.Host,
				Domain: domain.Name,
				Path:   route.Path,
				Port:   route.Port,
			})
		}
	}

	if len(orphanedRoutes) == 0 {
		return nil, allWarnings, OrphanedRoutesNotFoundError{}
	}

	return orphanedRoutes, allWarnings, err
}

// DeleteRouteByGUID deletes the Route associated with the provided Route GUID.
func (actor Actor) DeleteRouteByGUID(routeGUID string) (Warnings, error) {
	warnings, err := actor.CloudControllerClient.DeleteRoute(routeGUID)
	return Warnings(warnings), err
}

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
