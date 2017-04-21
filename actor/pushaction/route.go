package pushaction

import (
	"code.cloudfoundry.org/cli/actor/v2action"
	log "github.com/Sirupsen/logrus"
)

// FindOrReturnPartialRoute finds the route with the given host and domain. If
// it is unable to find the route, it will return back the partial route. When
// the route exists in another space, RouteInDifferentSpaceError is returned.
func (actor Actor) FindOrReturnPartialRoute(route v2action.Route) (v2action.Route, Warnings, error) {
	// This check only works for API versions 2.55 or higher. It will return
	// false for anything below that.
	log.Infoln("checking route existance for:", route.String())
	exists, warnings, err := actor.V2Actor.CheckRoute(route)
	if err != nil {
		log.Errorln("check route:", err)
		return v2action.Route{}, Warnings(warnings), err
	}

	if exists {
		log.Debug("route exists")

		// TODO: Use a more generic search mechanism to support path, port, and no host
		existingRoute, routeWarnings, err := actor.V2Actor.GetRouteByHostAndDomain(route.Host, route.Domain.GUID)
		if _, ok := err.(v2action.RouteNotFoundError); ok {
			log.Errorf("unable to find route %s in current space", route.String())
			return v2action.Route{}, append(Warnings(warnings), routeWarnings...), v2action.RouteInDifferentSpaceError{Route: route.String()}
		} else if err != nil {
			log.Errorln("finding route:", err)
			return v2action.Route{}, append(Warnings(warnings), routeWarnings...), err
		}

		if existingRoute.SpaceGUID != route.SpaceGUID {
			log.WithFields(log.Fields{
				"targeted_space_guid": route.SpaceGUID,
				"existing_space_guid": existingRoute.SpaceGUID,
			}).Errorf("route exists in different space the user has access to")
			return v2action.Route{}, append(Warnings(warnings), routeWarnings...), v2action.RouteInDifferentSpaceError{Route: route.String()}
		}

		log.Debugf("found route: %#v", existingRoute)
		return existingRoute, append(Warnings(warnings), routeWarnings...), err
	}

	log.Warnf("negitive existence check for route %s - returning partial route", route.String())
	log.Debugf("partialRoute: %#v", route)
	return route, Warnings(warnings), nil
}

// GetRouteWithDefaultDomain returns a route with the host and the default org
// domain. This may be a partial route (ie no GUID) if the route does not
// exist.
func (actor Actor) GetRouteWithDefaultDomain(host string, orgGUID string, spaceGUID string) (v2action.Route, Warnings, error) {
	defaultDomain, warnings, err := actor.DefaultDomain(orgGUID)
	if err != nil {
		log.Errorln("could not find default domains:", err.Error())
		return v2action.Route{}, Warnings(warnings), err
	}

	route, routeWarnings, err := actor.FindOrReturnPartialRoute(v2action.Route{
		Domain:    defaultDomain,
		Host:      host,
		SpaceGUID: spaceGUID,
	})
	return route, append(Warnings(warnings), routeWarnings...), err
}

func (actor Actor) routeInList(route v2action.Route, routes []v2action.Route) bool {
	for _, r := range routes {
		if r.GUID == route.GUID {
			return true
		}
	}

	return false
}
