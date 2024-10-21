package cfnetworkingaction

import (
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v9/resources"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . CloudControllerClient

type CloudControllerClient interface {
	GetApplicationByNameAndSpace(appName string, spaceGUID string) (resources.Application, ccv3.Warnings, error)
	GetApplications(query ...ccv3.Query) ([]resources.Application, ccv3.Warnings, error)
	GetOrganizations(query ...ccv3.Query) ([]resources.Organization, ccv3.Warnings, error)
	GetSpaces(query ...ccv3.Query) ([]resources.Space, ccv3.IncludedResources, ccv3.Warnings, error)
}
