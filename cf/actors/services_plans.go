package actors

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/cli/cf/actors/plan_builder"
	"github.com/cloudfoundry/cli/cf/actors/service_builder"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/models"
)

type ServicePlanActor interface {
	FindServiceAccess(string) (ServiceAccess, error)
	UpdateAllPlansForService(string, bool) (bool, error)
	UpdateOrgForService(string, string, bool) (bool, error)
	UpdateSinglePlanForService(string, string, bool) (PlanAccess, error)
	UpdatePlanAndOrgForService(string, string, string, bool) (PlanAccess, error)
}

type PlanAccess int

const (
	PlanAccessError PlanAccess = iota
	All
	Limited
	None
)

type ServiceAccess int

const (
	ServiceAccessError ServiceAccess = iota
	AllPlansArePublic
	AllPlansArePrivate
	AllPlansAreLimited
	SomePlansArePublicSomeAreLimited
	SomePlansArePublicSomeArePrivate
	SomePlansAreLimitedSomeArePrivate
	SomePlansArePublicSomeAreLimitedSomeArePrivate
)

type ServicePlanHandler struct {
	servicePlanRepo           api.ServicePlanRepository
	servicePlanVisibilityRepo api.ServicePlanVisibilityRepository
	orgRepo                   api.OrganizationRepository
	serviceBuilder            service_builder.ServiceBuilder
	planBuilder               plan_builder.PlanBuilder
}

func NewServicePlanHandler(plan api.ServicePlanRepository, vis api.ServicePlanVisibilityRepository, org api.OrganizationRepository, planBuilder plan_builder.PlanBuilder, serviceBuilder service_builder.ServiceBuilder) ServicePlanHandler {
	return ServicePlanHandler{
		servicePlanRepo:           plan,
		servicePlanVisibilityRepo: vis,
		orgRepo:                   org,
		serviceBuilder:            serviceBuilder,
		planBuilder:               planBuilder,
	}
}

func (actor ServicePlanHandler) UpdateAllPlansForService(serviceName string, setPlanVisibility bool) (bool, error) {
	service, err := actor.serviceBuilder.GetServiceByName(serviceName)
	if err != nil {
		return false, err
	}

	allPlansWereSet := true
	for _, plan := range service.Plans {
		planAccess, err := actor.updateSinglePlan(service, plan.Name, setPlanVisibility)
		if err != nil {
			return false, err
		}
		// If any plan is Limited we know that we have to change the visibility.
		planAlreadySet := ((planAccess == All) == setPlanVisibility) && planAccess != Limited
		allPlansWereSet = allPlansWereSet && planAlreadySet
	}
	return allPlansWereSet, nil
}

func (actor ServicePlanHandler) UpdateOrgForService(serviceName string, orgName string, setPlanVisibility bool) (bool, error) {
	var err error
	var service models.ServiceOffering

	service, err = actor.serviceBuilder.GetServiceByName(serviceName)
	if err != nil {
		return false, err
	}

	org, err := actor.orgRepo.FindByName(orgName)
	if err != nil {
		return false, err
	}

	allPlansWereSet := true
	for _, plan := range service.Plans {
		visibilityExists := plan.OrgHasVisibility(org.Name)
		if plan.Public || visibilityExists == setPlanVisibility {
			continue
		} else if visibilityExists && !setPlanVisibility {
			actor.deleteServicePlanVisibilities(map[string]string{"org_guid": org.Guid, "plan_guid": plan.Guid})
		} else if !visibilityExists && setPlanVisibility {
			err = actor.servicePlanVisibilityRepo.Create(plan.Guid, org.Guid)
			if err != nil {
				return false, err
			}
		}
		// We only get here once we have already updated a plan.
		allPlansWereSet = false
	}
	return allPlansWereSet, nil
}

func (actor ServicePlanHandler) UpdatePlanAndOrgForService(serviceName, planName, orgName string, setPlanVisibility bool) (PlanAccess, error) {
	service, err := actor.serviceBuilder.GetServiceByName(serviceName)
	if err != nil {
		return PlanAccessError, err
	}

	org, err := actor.orgRepo.FindByName(orgName)
	if err != nil {
		return PlanAccessError, err
	}

	found := false
	var servicePlan models.ServicePlanFields
	for i, val := range service.Plans {
		if val.Name == planName {
			found = true
			servicePlan = service.Plans[i]
		}
	}
	if !found {
		return PlanAccessError, errors.New(fmt.Sprintf("Service plan %s not found", planName))
	}

	if !servicePlan.Public && setPlanVisibility {
		if servicePlan.OrgHasVisibility(orgName) {
			return Limited, nil
		}

		// Enable service access
		err = actor.servicePlanVisibilityRepo.Create(servicePlan.Guid, org.Guid)
		if err != nil {
			return PlanAccessError, err
		}
	} else if !servicePlan.Public && !setPlanVisibility {
		// Disable service access
		if servicePlan.OrgHasVisibility(org.Name) {
			err = actor.deleteServicePlanVisibilities(map[string]string{"organization_guid": org.Guid, "service_plan_guid": servicePlan.Guid})
			if err != nil {
				return PlanAccessError, err
			}
		}
	}

	access := actor.findPlanAccess(servicePlan)
	return access, nil
}

func (actor ServicePlanHandler) UpdateSinglePlanForService(serviceName string, planName string, setPlanVisibility bool) (PlanAccess, error) {
	serviceOffering, err := actor.serviceBuilder.GetServiceByName(serviceName)
	if err != nil {
		return PlanAccessError, err
	}
	return actor.updateSinglePlan(serviceOffering, planName, setPlanVisibility)
}

func (actor ServicePlanHandler) updateSinglePlan(serviceOffering models.ServiceOffering, planName string, setPlanVisibility bool) (PlanAccess, error) {
	var planToUpdate *models.ServicePlanFields

	//find the service plan and set it as the only service plan for update
	for _, servicePlan := range serviceOffering.Plans {
		if servicePlan.Name == planName {
			planToUpdate = &servicePlan //he has the orgs inside him!!!
			break
		}
	}

	if planToUpdate == nil {
		return PlanAccessError, errors.New(fmt.Sprintf("The plan %s could not be found for service %s", planName, serviceOffering.Label))
	}

	err := actor.updateServicePlanAvailability(serviceOffering.Guid, *planToUpdate, setPlanVisibility)
	if err != nil {
		return PlanAccessError, err
	}

	access := actor.findPlanAccess(*planToUpdate)
	return access, nil
}

func (actor ServicePlanHandler) deleteServicePlanVisibilities(queryParams map[string]string) error {
	visibilities, err := actor.servicePlanVisibilityRepo.Search(queryParams)
	if err != nil {
		return err
	}
	for _, visibility := range visibilities {
		err = actor.servicePlanVisibilityRepo.Delete(visibility.Guid)
		if err != nil {
			return err
		}
	}

	return nil
}

func (actor ServicePlanHandler) updateServicePlanAvailability(serviceGuid string, servicePlan models.ServicePlanFields, setPlanVisibility bool) error {
	// We delete all service plan visibilities for the given Plan since the attribute public should function as a giant on/off
	// switch for all orgs. Thus we need to clean up any visibilities laying around so that they don't carry over.
	err := actor.deleteServicePlanVisibilities(map[string]string{"plan_guid": servicePlan.Guid})
	if err != nil {
		return err
	}

	if servicePlan.Public == setPlanVisibility {
		return nil
	}

	return actor.servicePlanRepo.Update(servicePlan, serviceGuid, setPlanVisibility)
}

func (actor ServicePlanHandler) FindServiceAccess(serviceName string) (ServiceAccess, error) {
	service, err := actor.serviceBuilder.GetServiceByName(serviceName)
	if err != nil {
		return ServiceAccessError, err
	}

	publicBucket, limitedBucket, privateBucket := 0, 0, 0

	for _, plan := range service.Plans {
		if plan.Public {
			publicBucket++
		} else if len(plan.OrgNames) > 0 {
			limitedBucket++
		} else {
			privateBucket++
		}
	}

	if publicBucket > 0 && limitedBucket == 0 && privateBucket == 0 {
		return AllPlansArePublic, nil
	}
	if publicBucket > 0 && limitedBucket > 0 && privateBucket == 0 {
		return SomePlansArePublicSomeAreLimited, nil
	}
	if publicBucket > 0 && privateBucket > 0 && limitedBucket == 0 {
		return SomePlansArePublicSomeArePrivate, nil
	}

	if limitedBucket > 0 && publicBucket == 0 && privateBucket == 0 {
		return AllPlansAreLimited, nil
	}
	if privateBucket > 0 && publicBucket == 0 && privateBucket == 0 {
		return AllPlansArePrivate, nil
	}
	if limitedBucket > 0 && privateBucket > 0 && publicBucket == 0 {
		return SomePlansAreLimitedSomeArePrivate, nil
	}
	return SomePlansArePublicSomeAreLimitedSomeArePrivate, nil
}

func (actor ServicePlanHandler) findPlanAccess(plan models.ServicePlanFields) PlanAccess {
	if plan.Public {
		return All
	} else if len(plan.OrgNames) > 0 {
		return Limited
	} else {
		return None
	}
}
