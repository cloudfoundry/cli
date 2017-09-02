package v2action

import (
	"fmt"
	"path"
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/types"
	log "github.com/sirupsen/logrus"
)

// OrphanedRoutesNotFoundError is an error wrapper that represents the case
// when no orphaned routes are found.
type OrphanedRoutesNotFoundError struct{}

// Error method to display the error message.
func (OrphanedRoutesNotFoundError) Error() string {
	return "No orphaned routes were found."
}

// RouteInDifferentSpaceError is returned when the route exists in a different
// space than the one requesting it.
type RouteInDifferentSpaceError struct {
	Route string
}

func (RouteInDifferentSpaceError) Error() string {
	return "route registered to another space"
}

// RouteNotFoundError is returned when a route cannot be found
type RouteNotFoundError struct {
	Host       string
	DomainGUID string
}

func (e RouteNotFoundError) Error() string {
	return fmt.Sprintf("Route with host %s and domain guid %s not found", e.Host, e.DomainGUID)
}

// RouteAlreadyExistsError is returned when a route already exists
type RouteAlreadyExistsError struct {
	Route Route
}

func (e RouteAlreadyExistsError) Error() string {
	return fmt.Sprintf("Route %s already exists", e.Route)
}

type Routes []Route

func (rs Routes) Summary() string {
	formattedRoutes := []string{}
	for _, route := range rs {
		formattedRoutes = append(formattedRoutes, route.String())
	}
	return strings.Join(formattedRoutes, ", ")
}

// Route represents a CLI Route.
type Route struct {
	Domain    Domain
	GUID      string
	Host      string
	Path      string
	Port      types.NullInt
	SpaceGUID string
}

// String formats the route in a human readable format.
func (r Route) String() string {
	routeString := r.Domain.Name

	if r.Port.IsSet {
		routeString = fmt.Sprintf("%s:%d", routeString, r.Port.Value)
	}

	if r.Host != "" {
		routeString = fmt.Sprintf("%s.%s", r.Host, routeString)
	}

	if r.Path != "" {
		routeString = path.Join(routeString, r.Path)
	}

	return routeString
}

func (actor Actor) BindRouteToApplication(routeGUID string, appGUID string) (Warnings, error) {
	_, warnings, err := actor.CloudControllerClient.BindRouteToApplication(routeGUID, appGUID)
	if _, ok := err.(ccerror.InvalidRelationError); ok {
		return Warnings(warnings), RouteInDifferentSpaceError{}
	}
	return Warnings(warnings), err
}

func (actor Actor) CreateRoute(route Route, generatePort bool) (Route, Warnings, error) {
	if route.Path != "" && !strings.HasPrefix(route.Path, "/") {
		route.Path = fmt.Sprintf("/%s", route.Path)
	}
	returnedRoute, warnings, err := actor.CloudControllerClient.CreateRoute(ActorToCCRoute(route), generatePort)
	return CCToActorRoute(returnedRoute, route.Domain), Warnings(warnings), err
}

func (actor Actor) CreateRouteWithExistenceCheck(orgGUID string, spaceName string, route Route, generatePort bool) (Route, Warnings, error) {
	space, warnings, spaceErr := actor.GetSpaceByOrganizationAndName(orgGUID, spaceName)
	if spaceErr != nil {
		return Route{}, Warnings(warnings), spaceErr
	}
	route.SpaceGUID = space.GUID

	if route.Domain.GUID == "" {
		domains, orgDomainWarnings, getDomainErr := actor.GetDomainsByNameAndOrganization([]string{route.Domain.Name}, orgGUID)
		warnings = append(warnings, orgDomainWarnings...)
		if getDomainErr != nil {
			return Route{}, warnings, getDomainErr
		} else if len(domains) == 0 {
			return Route{}, warnings, DomainNotFoundError{Name: route.Domain.Name}
		}
		route.Domain.GUID = domains[0].GUID
	}

	foundRoute, spaceRouteWarnings, findErr := actor.FindRouteBoundToSpaceWithSettings(route)
	warnings = append(warnings, spaceRouteWarnings...)
	routeAlreadyExists := true
	if _, ok := findErr.(RouteNotFoundError); ok {
		routeAlreadyExists = false
	} else if findErr != nil {
		return Route{}, Warnings(warnings), findErr
	}

	if routeAlreadyExists {
		return Route{}, Warnings(warnings), RouteAlreadyExistsError{Route: foundRoute}
	}

	createdRoute, createRouteWarnings, createErr := actor.CreateRoute(route, generatePort)
	warnings = append(warnings, createRouteWarnings...)
	if createErr != nil {
		return Route{}, Warnings(warnings), createErr
	}

	return createdRoute, Warnings(warnings), nil
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
		apps, warnings, err := actor.GetRouteApplications(route.GUID)
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
func (actor Actor) GetApplicationRoutes(applicationGUID string) (Routes, Warnings, error) {
	var allWarnings Warnings
	ccv2Routes, warnings, err := actor.CloudControllerClient.GetApplicationRoutes(applicationGUID)
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
	ccv2Routes, warnings, err := actor.CloudControllerClient.GetSpaceRoutes(spaceGUID)
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

func (actor Actor) CheckRoute(route Route) (bool, Warnings, error) {
	exists, warnings, err := actor.CloudControllerClient.CheckRoute(ActorToCCRoute(route))
	return exists, Warnings(warnings), err
}

// FindRouteBoundToSpaceWithSettings finds the route with the given host,
// domain and space.  If it is unable to find the route, it will check if it
// exists anywhere in the system. When the route exists in another space,
// RouteInDifferentSpaceError is returned.
func (actor Actor) FindRouteBoundToSpaceWithSettings(route Route) (Route, Warnings, error) {
	// TODO: Use a more generic search mechanism to support path, port, and no host
	existingRoute, warnings, err := actor.GetRouteByHostAndDomain(route.Host, route.Domain.GUID)
	if routeNotFoundErr, ok := err.(RouteNotFoundError); ok {
		// This check only works for API versions 2.55 or higher. It will return
		// false for anything below that.
		log.Infoln("checking route existence for: %s", route)
		exists, checkRouteWarnings, chkErr := actor.CheckRoute(route)
		if chkErr != nil {
			log.Errorln("check route:", err)
			return Route{}, append(Warnings(warnings), checkRouteWarnings...), chkErr
		}

		// This will happen if the route exists in a space to which the user does
		// not have access.
		if exists {
			log.Errorf("unable to find route %s in current space", route.String())
			return Route{}, append(Warnings(warnings), checkRouteWarnings...), RouteInDifferentSpaceError{Route: route.String()}
		}

		log.Warnf("negative existence check for route %s - returning partial route", route.String())
		log.Debugf("partialRoute: %#v", route)
		return Route{}, append(Warnings(warnings), checkRouteWarnings...), routeNotFoundErr
	} else if err != nil {
		log.Errorln("finding route:", err)
		return Route{}, Warnings(warnings), err
	}

	if existingRoute.SpaceGUID != route.SpaceGUID {
		log.WithFields(log.Fields{
			"targeted_space_guid": route.SpaceGUID,
			"existing_space_guid": existingRoute.SpaceGUID,
		}).Errorf("route exists in different space the user has access to")
		return Route{}, Warnings(warnings), RouteInDifferentSpaceError{Route: route.String()}
	}

	log.Debugf("found route: %#v", existingRoute)
	return existingRoute, Warnings(warnings), err
}

// GetRouteByHostAndDomain returns the HTTP route with the matching host and
// the associate domain GUID.
func (actor Actor) GetRouteByHostAndDomain(host string, domainGUID string) (Route, Warnings, error) {
	ccv2Routes, warnings, err := actor.CloudControllerClient.GetRoutes(
		ccv2.Query{
			Filter:   ccv2.HostFilter,
			Operator: ccv2.EqualOperator,
			Values:   []string{host},
		},
		ccv2.Query{
			Filter:   ccv2.DomainGUIDFilter,
			Operator: ccv2.EqualOperator,
			Values:   []string{domainGUID},
		},
	)
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

func ActorToCCRoute(route Route) ccv2.Route {
	return ccv2.Route{
		DomainGUID: route.Domain.GUID,
		GUID:       route.GUID,
		Host:       route.Host,
		Path:       route.Path,
		Port:       route.Port,
		SpaceGUID:  route.SpaceGUID,
	}
}

func CCToActorRoute(ccv2Route ccv2.Route, domain Domain) Route {
	return Route{
		Domain:    domain,
		GUID:      ccv2Route.GUID,
		Host:      ccv2Route.Host,
		Path:      ccv2Route.Path,
		Port:      ccv2Route.Port,
		SpaceGUID: ccv2Route.SpaceGUID,
	}
}

func (actor Actor) applyDomain(ccv2Routes []ccv2.Route) (Routes, Warnings, error) {
	var routes Routes
	var allWarnings Warnings

	for _, ccv2Route := range ccv2Routes {
		domain, warnings, err := actor.GetDomain(ccv2Route.DomainGUID)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return nil, allWarnings, err
		}
		routes = append(routes, CCToActorRoute(ccv2Route, domain))
	}

	return routes, allWarnings, nil
}
