package pushaction

import (
	"fmt"
	"net/url"
	"strings"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/types"
	log "github.com/sirupsen/logrus"
)

func (actor Actor) BindRoutes(config ApplicationConfig) (ApplicationConfig, bool, Warnings, error) {
	log.Info("binding routes")

	var boundRoutes bool
	var allWarnings Warnings

	for _, route := range config.DesiredRoutes {
		if !actor.routeInListByGUID(route, config.CurrentRoutes) {
			log.Debugf("binding route: %#v", route)
			warnings, err := actor.bindRouteToApp(route, config.DesiredApplication.GUID)
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

func (actor Actor) CalculateRoutes(routes []string, orgGUID string, spaceGUID string, existingRoutes []v2action.Route) ([]v2action.Route, Warnings, error) {
	calculatedRoutes, unknownRoutes := actor.spitExistingRoutes(existingRoutes, routes)
	possibleDomains, err := actor.generatePossibleDomains(unknownRoutes)
	if err != nil {
		log.Errorln("domain breakdown:", err)
		return nil, nil, err
	}

	var allWarnings Warnings
	foundDomains, warnings, err := actor.V2Actor.GetDomainsByNameAndOrganization(possibleDomains, orgGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		log.Errorln("domain lookup:", err)
		return nil, allWarnings, err
	}
	nameToFoundDomain := map[string]v2action.Domain{}
	for _, foundDomain := range foundDomains {
		log.WithField("domain", foundDomain.Name).Debug("found domain")
		nameToFoundDomain[foundDomain.Name] = foundDomain
	}

	for _, route := range unknownRoutes {
		log.WithField("route", route).Debug("generating route")

		root, port, path, parseErr := actor.parseURL(route)
		if parseErr != nil {
			log.Errorln("parse route:", parseErr)
			return nil, allWarnings, parseErr
		}

		host, domain, domainErr := actor.calculateRoute(root, nameToFoundDomain)
		if _, ok := domainErr.(v2action.DomainNotFoundError); ok {
			log.Error("no matching domains")
			return nil, allWarnings, actionerror.NoMatchingDomainError{Route: route}
		} else if domainErr != nil {
			log.Errorln("matching domains:", domainErr)
			return nil, allWarnings, domainErr
		}

		// TODO: redo once TCP routing has been completed
		if port.IsSet && domain.IsHTTP() {
			return nil, allWarnings, actionerror.InvalidHTTPRouteSettings{Domain: domain.Name}
		}

		calculatedRoute, routeWarnings, routeErr := actor.findOrReturnPartialRouteWithSettings(v2action.Route{
			Host:      strings.Join(host, "."),
			Domain:    domain,
			Path:      path,
			Port:      port,
			SpaceGUID: spaceGUID,
		})
		allWarnings = append(allWarnings, routeWarnings...)
		if routeErr != nil {
			log.Errorln("route lookup:", routeErr)
			return nil, allWarnings, routeErr
		}

		calculatedRoutes = append(calculatedRoutes, calculatedRoute)
	}

	return calculatedRoutes, allWarnings, nil
}

func (actor Actor) CreateAndBindApplicationRoutes(orgGUID string, spaceGUID string, app v2action.Application) (Warnings, error) {
	var warnings Warnings
	defaultRoute, domainWarnings, err := actor.getDefaultRoute(orgGUID, spaceGUID, app.Name)
	warnings = append(warnings, domainWarnings...)
	if err != nil {
		return warnings, err
	}

	boundRoutes, appRouteWarnings, err := actor.V2Actor.GetApplicationRoutes(app.GUID)
	warnings = append(warnings, appRouteWarnings...)
	if err != nil {
		return warnings, err
	}

	_, routeAlreadyBound := actor.routeInListBySettings(defaultRoute, boundRoutes)
	if routeAlreadyBound {
		return warnings, err
	}

	spaceRoute, spaceRouteWarnings, err := actor.V2Actor.FindRouteBoundToSpaceWithSettings(defaultRoute)
	warnings = append(warnings, spaceRouteWarnings...)
	routeAlreadyExists := true
	if _, ok := err.(v2action.RouteNotFoundError); ok {
		routeAlreadyExists = false
	} else if err != nil {
		return warnings, err
	}

	if !routeAlreadyExists {
		var createRouteWarning v2action.Warnings
		spaceRoute, createRouteWarning, err = actor.V2Actor.CreateRoute(defaultRoute, false)
		warnings = append(warnings, createRouteWarning...)
		if err != nil {
			return warnings, err
		}
	}

	bindWarnings, err := actor.V2Actor.BindRouteToApplication(spaceRoute.GUID, app.GUID)
	warnings = append(warnings, bindWarnings...)
	return warnings, err
}

func (actor Actor) CreateRoutes(config ApplicationConfig) (ApplicationConfig, bool, Warnings, error) {
	log.Info("creating routes")

	var routes []v2action.Route
	var createdRoutes bool
	var allWarnings Warnings

	for _, route := range config.DesiredRoutes {
		if route.GUID == "" {
			log.WithField("route", route).Debug("creating route")

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

// GetRouteWithDefaultDomain returns a route with the host and the default org
// domain. This may be a partial route (ie no GUID) if the route does not
// exist.
func (actor Actor) GetRouteWithDefaultDomain(host string, orgGUID string, spaceGUID string, knownRoutes []v2action.Route) (v2action.Route, Warnings, error) {
	defaultDomain, warnings, err := actor.DefaultDomain(orgGUID)
	if err != nil {
		log.Errorln("could not find default domains:", err.Error())
		return v2action.Route{}, warnings, err
	}

	defaultRoute := v2action.Route{
		Domain:    defaultDomain,
		Host:      strings.ToLower(host),
		SpaceGUID: spaceGUID,
	}

	cachedRoute, found := actor.routeInListBySettings(defaultRoute, knownRoutes)
	if !found {
		route, routeWarnings, err := actor.V2Actor.FindRouteBoundToSpaceWithSettings(defaultRoute)
		if _, ok := err.(v2action.RouteNotFoundError); ok {
			return defaultRoute, append(warnings, routeWarnings...), nil
		}
		return route, append(warnings, routeWarnings...), err
	}
	return cachedRoute, warnings, nil
}

func (actor Actor) bindRouteToApp(route v2action.Route, appGUID string) (v2action.Warnings, error) {
	warnings, err := actor.V2Actor.BindRouteToApplication(route.GUID, appGUID)
	if _, ok := err.(v2action.RouteInDifferentSpaceError); ok {
		return warnings, v2action.RouteInDifferentSpaceError{Route: route.String()}
	}
	return warnings, err
}

func (actor Actor) calculateRoute(route string, domainCache map[string]v2action.Domain) ([]string, v2action.Domain, error) {
	host, domain := actor.splitHost(route)
	if domain, ok := domainCache[route]; ok {
		return nil, domain, nil
	}

	if host == "" {
		return nil, v2action.Domain{}, v2action.DomainNotFoundError{Name: route}
	}

	hosts, foundDomain, err := actor.calculateRoute(domain, domainCache)
	hosts = append([]string{host}, hosts...)

	return hosts, foundDomain, err
}

func (actor Actor) findOrReturnPartialRouteWithSettings(route v2action.Route) (v2action.Route, Warnings, error) {
	cachedRoute, warnings, err := actor.V2Actor.FindRouteBoundToSpaceWithSettings(route)
	if _, ok := err.(v2action.RouteNotFoundError); ok {
		return route, Warnings(warnings), nil
	}
	return cachedRoute, Warnings(warnings), err
}

func (actor Actor) generatePossibleDomains(routes []string) ([]string, error) {
	var hostnames []string
	for _, route := range routes {
		host, _, _, err := actor.parseURL(route)
		if err != nil {
			return nil, err
		}
		hostnames = append(hostnames, host)
	}

	possibleDomains := map[string]interface{}{}
	for _, route := range hostnames {
		count := strings.Count(route, ".")
		domains := strings.SplitN(route, ".", count)

		for i := range domains {
			domain := strings.Join(domains[i:], ".")
			possibleDomains[domain] = nil
		}
	}

	var domains []string
	for domain := range possibleDomains {
		domains = append(domains, domain)
	}

	log.Debugln("domain brakedown:", strings.Join(domains, ","))
	return domains, nil
}

func (actor Actor) getDefaultRoute(orgGUID string, spaceGUID string, appName string) (v2action.Route, Warnings, error) {
	defaultDomain, domainWarnings, err := actor.DefaultDomain(orgGUID)
	if err != nil {
		return v2action.Route{}, domainWarnings, err
	}

	return v2action.Route{
		Host:      appName,
		Domain:    defaultDomain,
		SpaceGUID: spaceGUID,
	}, domainWarnings, nil
}

func (Actor) parseURL(route string) (string, types.NullInt, string, error) {
	if !(strings.HasPrefix(route, "http://") || strings.HasPrefix(route, "https://")) {
		route = fmt.Sprintf("http://%s", route)
	}
	parsedURL, err := url.Parse(route)
	if err != nil {
		return "", types.NullInt{}, "", err
	}

	path := parsedURL.RequestURI()
	if path == "/" {
		path = ""
	}

	var port types.NullInt
	err = port.ParseStringValue(parsedURL.Port())
	return parsedURL.Hostname(), port, path, err
}

func (Actor) routeInListByGUID(route v2action.Route, routes []v2action.Route) bool {
	for _, r := range routes {
		if r.GUID == route.GUID {
			return true
		}
	}

	return false
}

func (Actor) routeInListByName(route string, routes []v2action.Route) (v2action.Route, bool) {
	for _, r := range routes {
		if r.String() == route {
			return r, true
		}
	}

	return v2action.Route{}, false
}
func (Actor) routeInListBySettings(route v2action.Route, routes []v2action.Route) (v2action.Route, bool) {
	for _, r := range routes {
		if r.Host == route.Host && r.Path == route.Path && r.Port == route.Port &&
			r.SpaceGUID == route.SpaceGUID && r.Domain.GUID == route.Domain.GUID {
			return r, true
		}
	}

	return v2action.Route{}, false
}

func (actor Actor) spitExistingRoutes(existingRoutes []v2action.Route, routes []string) ([]v2action.Route, []string) {
	var cachedRoutes []v2action.Route
	for _, route := range existingRoutes {
		cachedRoutes = append(cachedRoutes, route)
	}

	var unknownRoutes []string
	for _, route := range routes {
		if _, found := actor.routeInListByName(route, existingRoutes); !found {
			log.WithField("route", route).Debug("unable to find route in cache")
			unknownRoutes = append(unknownRoutes, route)
		}
	}
	return cachedRoutes, unknownRoutes
}

func (Actor) splitHost(url string) (string, string) {
	count := strings.Count(url, ".")
	if count == 1 {
		return "", url
	}

	split := strings.SplitN(url, ".", 2)
	return split[0], split[1]
}
