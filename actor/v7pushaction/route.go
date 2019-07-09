package v7pushaction

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	log "github.com/sirupsen/logrus"
)

func (actor Actor) CreateAndMapDefaultApplicationRoute(orgGUID string, spaceGUID string, app v7action.Application) (Warnings, error) {
	log.Info("default route creation/mapping")
	var warnings Warnings
	defaultRoute, domainWarnings, err := actor.getDefaultRoute(orgGUID, spaceGUID, app.Name)
	warnings = append(warnings, domainWarnings...)
	if err != nil {
		log.Errorln("getting default route:", err)
		return warnings, err
	}
	log.WithField("defaultRoute", defaultRoute.String()).Debug("calculated default route")

	route, getRouteWarnings, err := actor.V7Actor.GetRouteByAttributes(defaultRoute.DomainName, defaultRoute.DomainGUID, defaultRoute.Host, defaultRoute.Path)
	warnings = append(warnings, getRouteWarnings...)
	if err != nil {
		if _, ok := err.(actionerror.RouteNotFoundError); !ok {
			return warnings, err
		}
		var createRouteWarnings v7action.Warnings
		route, createRouteWarnings, err = actor.V7Actor.CreateRoute(
			spaceGUID,
			defaultRoute.DomainName,
			defaultRoute.Host,
			defaultRoute.Path,
		)
		warnings = append(warnings, createRouteWarnings...)
		if err != nil {
			return warnings, err
		}
	}

	log.Debug("mapping default route")
	mapWarnings, err := actor.V7Actor.MapRoute(route.GUID, app.GUID)
	warnings = append(warnings, mapWarnings...)
	return warnings, err
}

func (actor Actor) getDefaultRoute(orgGUID string, spaceGUID string, appName string) (v7action.Route, Warnings, error) {
	domain, v7domainWarnings, err := actor.V7Actor.GetDefaultDomain(orgGUID)

	domainWarnings := append(Warnings{}, v7domainWarnings...)
	if err != nil {
		return v7action.Route{}, domainWarnings, err
	}

	return v7action.Route{
		Host:       appName,
		DomainName: domain.Name,
		DomainGUID: domain.GUID,
		SpaceGUID:  spaceGUID,
	}, domainWarnings, nil
}
