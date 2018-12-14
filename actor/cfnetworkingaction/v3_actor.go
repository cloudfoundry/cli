package cfnetworkingaction

import "code.cloudfoundry.org/cli/actor/v3action"

//go:generate counterfeiter . V3Actor
type V3Actor interface {
	GetApplicationByNameAndSpace(appName string, spaceGUID string) (v3action.Application, v3action.Warnings, error)
	GetApplicationsBySpace(spaceGUID string) ([]v3action.Application, v3action.Warnings, error)
	GetOrganizationByName(name string) (v3action.Organization, v3action.Warnings, error)
	GetSpaceByNameAndOrganization(spaceName string, orgGUID string) (v3action.Space, v3action.Warnings, error)
	GetApplicationsByGUIDs(appGUIDs ...string) ([]v3action.Application, v3action.Warnings, error)
	GetSpacesByGUIDs(spaceGUIDs ...string) ([]v3action.Space, v3action.Warnings, error)
	GetOrganizationsByGUIDs(orgGUIDs ...string) ([]v3action.Organization, v3action.Warnings, error)
}
