package pushaction

import (
	"code.cloudfoundry.org/cli/actor/v2action"
	log "github.com/sirupsen/logrus"
)

func (actor Actor) CreateRoutes(config ApplicationConfig) (ApplicationConfig, bool, Warnings, error) {
	log.Info("creating routes")

	var routes []v2action.Route
	var createdRoutes bool
	var allWarnings Warnings

	for _, route := range config.DesiredRoutes {
		if route.GUID == "" {
			log.Debugf("creating route: %#v", route)

			createdRoute, warnings, err := actor.V2Actor.CreateRoute(route, false)
			allWarnings = append(allWarnings, warnings...)
			if err != nil {
				log.Errorln("creating route:", err)
				return ApplicationConfig{}, true, allWarnings, err
			}
			routes = append(routes, createdRoute)

			createdRoutes = true
		} else {
			log.WithField("route", route).Debug("already exists, skipping")
			routes = append(routes, route)
		}
	}
	config.DesiredRoutes = routes

	return config, createdRoutes, allWarnings, nil
}

func (actor Actor) BindRoutes(config ApplicationConfig) (ApplicationConfig, bool, Warnings, error) {
	log.Info("binding routes")

	var boundRoutes bool
	var allWarnings Warnings

	for _, route := range config.DesiredRoutes {
		if !actor.routeInListByGUID(route, config.CurrentRoutes) {
			log.Debugf("binding route: %#v", route)
			warnings, err := actor.BindRouteToApp(route, config.DesiredApplication.GUID)
			allWarnings = append(allWarnings, warnings...)
			if err != nil {
				log.Errorln("binding route:", err)
				return ApplicationConfig{}, false, allWarnings, err
			}
			boundRoutes = true
		} else {
			log.Debugf("route %s already bound to app", route)
		}
	}
	log.Debug("binding routes complete")
	config.CurrentRoutes = config.DesiredRoutes

	return config, boundRoutes, allWarnings, nil
}

// GetRouteWithDefaultDomain returns a route with the host and the default org
// domain. This may be a partial route (ie no GUID) if the route does not
// exist.
func (actor Actor) GetRouteWithDefaultDomain(host string, orgGUID string, spaceGUID string, knownRoutes []v2action.Route) (v2action.Route, Warnings, error) {
	defaultDomain, warnings, err := actor.DefaultDomain(orgGUID)
	if err != nil {
		log.Errorln("could not find default domains:", err.Error())
		return v2action.Route{}, Warnings(warnings), err
	}

	defaultRoute := v2action.Route{
		Domain:    defaultDomain,
		Host:      host,
		SpaceGUID: spaceGUID,
	}

	if cachedRoute, found := actor.routeInListBySettings(defaultRoute, knownRoutes); !found {
		route, routeWarnings, err := actor.V2Actor.FindRouteBoundToSpaceWithSettings(defaultRoute)
		if _, ok := err.(v2action.RouteNotFoundError); ok {
			return defaultRoute, append(Warnings(warnings), routeWarnings...), nil
		}
		return route, append(Warnings(warnings), routeWarnings...), err
	} else {
		return cachedRoute, Warnings(warnings), nil
	}
}

func (actor Actor) BindRouteToApp(route v2action.Route, appGUID string) (v2action.Warnings, error) {
	warnings, err := actor.V2Actor.BindRouteToApplication(route.GUID, appGUID)
	if _, ok := err.(v2action.RouteInDifferentSpaceError); ok {
		return warnings, v2action.RouteInDifferentSpaceError{Route: route.String()}
	}
	return warnings, err
}

func (_ Actor) routeInListByGUID(route v2action.Route, routes []v2action.Route) bool {
	for _, r := range routes {
		if r.GUID == route.GUID {
			return true
		}
	}

	return false
}

func (_ Actor) routeInListBySettings(route v2action.Route, routes []v2action.Route) (v2action.Route, bool) {
	for _, r := range routes {
		if r.Host == route.Host && r.Path == route.Path && r.Port == route.Port &&
			r.SpaceGUID == route.SpaceGUID && r.Domain.GUID == route.Domain.GUID {
			return r, true
		}
	}

	return v2action.Route{}, false
}
