package rpc

import (
	"code.cloudfoundry.org/cli/actor/v7action"
)

//go:generate counterfeiter . PluginActor

type PluginActor interface {
	GetApplicationsBySpace(spaceGUID string) ([]v7action.Application, v7action.Warnings, error)
	GetDetailedAppSummary(appName string, spaceGUID string, withObfuscatedValues bool) (v7action.DetailedApplicationSummary, v7action.Warnings, error)
	GetOrganizationByName(string) (v7action.Organization, v7action.Warnings, error)
	GetOrganizationSpaces(string) ([]v7action.Space, v7action.Warnings, error)
	GetSpaceByNameAndOrganization(spaceName string, orgGUID string) (v7action.Space, v7action.Warnings, error)
	RefreshAccessToken() (accessToken string, err error)
}
