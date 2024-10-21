package v2v3action

import (
	"code.cloudfoundry.org/cli/v7/actor/v3action"
	"code.cloudfoundry.org/cli/v7/resources"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . V3Actor

type V3Actor interface {
	ManifestV3Actor
	GetApplicationSummaryByNameAndSpace(appName string, spaceGUID string, withObfuscatedValues bool) (v3action.ApplicationSummary, v3action.Warnings, error)
	GetOrganizationByName(orgName string) (v3action.Organization, v3action.Warnings, error)
	ShareServiceInstanceToSpaces(serviceInstanceGUID string, spaceGUIDs []string) (resources.RelationshipList, v3action.Warnings, error)
	UnshareServiceInstanceByServiceInstanceAndSpace(serviceInstanceGUID string, spaceGUID string) (v3action.Warnings, error)

	CloudControllerAPIVersion() string
}
