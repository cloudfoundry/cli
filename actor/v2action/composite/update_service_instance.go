package composite

import (
	"code.cloudfoundry.org/cli/actor/v2action"
)

//go:generate counterfeiter . GetServiceInstanceActor

type GetServiceInstanceActor interface {
	GetServiceInstanceByNameAndSpace(name string, spaceGUID string) (v2action.ServiceInstance, v2action.Warnings, error)
}

//go:generate counterfeiter . GetServicePlanActor

type GetServicePlanActor interface {
	GetServicePlan(servicePlanGUID string) (v2action.ServicePlan, v2action.Warnings, error)
}

//go:generate counterfeiter . UpdateServiceInstanceMaintenanceInfoActor

type UpdateServiceInstanceMaintenanceInfoActor interface {
	UpdateServiceInstanceMaintenanceInfo(serviceInsrtanceGUID string, maintenanceInfo v2action.MaintenanceInfo) (v2action.Warnings, error)
}

//go:generate counterfeiter . GetAPIVersionActor

type GetAPIVersionActor interface {
	CloudControllerAPIVersion() string
}
