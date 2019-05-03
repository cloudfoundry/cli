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

type UpdateServiceInstanceCompositeActor struct {
	GetServiceInstanceActor                   GetServiceInstanceActor
	GetServicePlanActor                       GetServicePlanActor
	UpdateServiceInstanceMaintenanceInfoActor UpdateServiceInstanceMaintenanceInfoActor
}

// UpgradeServiceInstance requests update on the service instance with the `maintenance_info` available on the plan
func (c UpdateServiceInstanceCompositeActor) UpgradeServiceInstance(serviceInstanceGUID, servicePlanGUID string) (v2action.Warnings, error) {
	servicePlan, warnings, err := c.GetServicePlanActor.GetServicePlan(servicePlanGUID)
	if err != nil {
		return warnings, err
	}
	updateWarnings, err := c.UpdateServiceInstanceMaintenanceInfoActor.UpdateServiceInstanceMaintenanceInfo(
		serviceInstanceGUID,
		v2action.MaintenanceInfo(servicePlan.MaintenanceInfo),
	)
	return append(warnings, updateWarnings...), err
}

// GetServiceInstanceByNameAndSpace gets the service instance by name and space guid provided
func (c UpdateServiceInstanceCompositeActor) GetServiceInstanceByNameAndSpace(name string, spaceGUID string) (v2action.ServiceInstance, v2action.Warnings, error) {
	return c.GetServiceInstanceActor.GetServiceInstanceByNameAndSpace(name, spaceGUID)
}
