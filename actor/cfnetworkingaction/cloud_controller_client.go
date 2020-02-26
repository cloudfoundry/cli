package cfnetworkingaction

import "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"

//go:generate counterfeiter . CloudControllerClient

type CloudControllerClient interface {
	GetApplicationByNameAndSpace(appName string, spaceGUID string) (ccv3.Application, ccv3.Warnings, error)
	GetApplications(query ...ccv3.Query) ([]ccv3.Application, ccv3.Warnings, error)
	GetOrganizations(query ...ccv3.Query) ([]ccv3.Organization, ccv3.Warnings, error)
	GetSpaces(query ...ccv3.Query) ([]ccv3.Space, ccv3.Warnings, error)
}
