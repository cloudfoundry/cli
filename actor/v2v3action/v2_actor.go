package v2v3action

import "code.cloudfoundry.org/cli/v7/actor/v2action"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . V2Actor

type V2Actor interface {
	ManifestV2Actor
	GetApplicationInstancesWithStatsByApplication(guid string) ([]v2action.ApplicationInstanceWithStats, v2action.Warnings, error)
	GetApplicationRoutes(appGUID string) (v2action.Routes, v2action.Warnings, error)
	GetFeatureFlags() ([]v2action.FeatureFlag, v2action.Warnings, error)
	GetService(serviceGUID string) (v2action.Service, v2action.Warnings, error)
	GetServiceInstanceByNameAndSpace(serviceInstanceName string, spaceGUID string) (v2action.ServiceInstance, v2action.Warnings, error)
	GetServiceInstanceSharedTosByServiceInstance(serviceInstanceGUID string) ([]v2action.ServiceInstanceSharedTo, v2action.Warnings, error)
	GetSpaceByOrganizationAndName(orgGUID string, spaceName string) (v2action.Space, v2action.Warnings, error)
}
