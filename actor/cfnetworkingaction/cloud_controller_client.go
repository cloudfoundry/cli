package cfnetworkingaction

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
)

//go:generate counterfeiter . CloudControllerClient

type CloudControllerClient interface {
	GetApplicationByNameAndSpace(appName string, spaceGUID string) (resources.Application, ccv3.Warnings, error)
	GetApplications(query ...ccv3.Query) ([]resources.Application, ccv3.Warnings, error)
	GetOrganizations(query ...ccv3.Query) ([]ccv3.Organization, ccv3.Warnings, error)
	GetSpaces(query ...ccv3.Query) ([]ccv3.Space, ccv3.IncludedResources, ccv3.Warnings, error)
}
