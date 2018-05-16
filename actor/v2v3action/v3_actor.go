package v2v3action

import (
	"code.cloudfoundry.org/cli/actor/v3action"
)

//go:generate counterfeiter . V3Actor

type V3Actor interface {
	ManifestV3Actor
	GetOrganizationByName(orgName string) (v3action.Organization, v3action.Warnings, error)
	ShareServiceInstanceToSpaces(serviceInstanceGUID string, spaceGUIDs []string) (v3action.RelationshipList, v3action.Warnings, error)
	UnshareServiceInstanceByServiceInstanceAndSpace(serviceInstanceGUID string, spaceGUID string) (v3action.Warnings, error)

	CloudControllerAPIVersion() string
}
