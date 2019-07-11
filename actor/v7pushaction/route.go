package v7pushaction

import (
	"strings"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	log "github.com/sirupsen/logrus"
)

type GenesisTechnique int

const (
	DefaultRoute GenesisTechnique = iota
	RandomRoute
)

func (actor Actor) CreateAndMapRoute(orgGUID string, spaceGUID string, app v7action.Application, gt GenesisTechnique) (Warnings, error) {
	var hostname string
	var allWarnings Warnings

	domain, domainWarnings, err := actor.V7Actor.GetDefaultDomain(orgGUID)
	allWarnings = append(allWarnings, domainWarnings...)
	if err != nil {
		return allWarnings, err
	}

	switch gt {
	case DefaultRoute:
		hostname = app.Name
	case RandomRoute:
		hostname = strings.Join(
			[]string{
				app.Name,
				actor.RandomWordGenerator.RandomAdjective(),
				actor.RandomWordGenerator.RandomNoun(),
			},
			"-",
		)
	}

	route, getRouteWarnings, err := actor.V7Actor.GetRouteByAttributes(domain.Name, domain.GUID, hostname, "")
	allWarnings = append(allWarnings, getRouteWarnings...)
	if err != nil {
		if _, ok := err.(actionerror.RouteNotFoundError); !ok {
			return allWarnings, err
		}
		var createRouteWarnings v7action.Warnings
		route, createRouteWarnings, err = actor.V7Actor.CreateRoute(
			spaceGUID,
			domain.Name,
			hostname,
			"",
		)
		allWarnings = append(allWarnings, createRouteWarnings...)
		if err != nil {
			return allWarnings, err
		}
	}

	log.Debug("mapping route", route.URL)
	mapWarnings, err := actor.V7Actor.MapRoute(route.GUID, app.GUID)
	allWarnings = append(allWarnings, mapWarnings...)
	return allWarnings, err
}
