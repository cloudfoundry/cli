package v7pushaction

import (
	"code.cloudfoundry.org/cli/actor/v2action"
)

//go:generate counterfeiter . V2Actor

type V2Actor interface {
	MapRouteToApplication(routeGUID string, appGUID string) (v2action.Warnings, error)
	CreateRoute(route v2action.Route, generatePort bool) (v2action.Route, v2action.Warnings, error)
	FindRouteBoundToSpaceWithSettings(route v2action.Route) (v2action.Route, v2action.Warnings, error)
	GetApplicationRoutes(applicationGUID string) (v2action.Routes, v2action.Warnings, error)
	GetOrganizationDomains(orgGUID string) ([]v2action.Domain, v2action.Warnings, error)
}
