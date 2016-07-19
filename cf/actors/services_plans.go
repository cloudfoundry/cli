package actors

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/api/organizations"

	"code.cloudfoundry.org/cli/cf/actors/planbuilder"
	"code.cloudfoundry.org/cli/cf/actors/servicebuilder"
	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/models"
)

//go:generate counterfeiter . ServicePlanActor

type ServicePlanActor interface {
	FindServiceAccess(string, string) (ServiceAccess, error)
	UpdateAllPlansForService(string, bool) error
	UpdateOrgForService(string, string, bool) error
	UpdateSinglePlanForService(string, string, bool) error
	UpdatePlanAndOrgForService(string, string, string, bool) error
}

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
	orgRepo                   organizations.OrganizationRepository
	serviceBuilder            servicebuilder.ServiceBuilder
	planBuilder               planbuilder.PlanBuilder
}

func NewServicePlanHandler(plan api.ServicePlanRepository, vis api.ServicePlanVisibilityRepository, org organizations.OrganizationRepository, planBuilder planbuilder.PlanBuilder, serviceBuilder servicebuilder.ServiceBuilder) ServicePlanHandler {
	return ServicePlanHandler{
		servicePlanRepo:           plan,
		servicePlanVisibilityRepo: vis,
		orgRepo:                   org,
		serviceBuilder:            serviceBuilder,
		planBuilder:               planBuilder,
	}
}

func (actor ServicePlanHandler) UpdateAllPlansForService(serviceName string, setPlanVisibility bool) error {
	service, err := actor.serviceBuilder.GetServiceByNameWithPlans(serviceName)
	if err != nil {
		return err
	}

	for _, plan := range service.Plans {
		err = actor.updateSinglePlan(service, plan.Name, setPlanVisibility)
		if err != nil {
			return err
		}
	}
	return nil
}

func (actor ServicePlanHandler) UpdateOrgForService(serviceName string, orgName string, setPlanVisibility bool) error {
	service, err := actor.serviceBuilder.GetServiceByNameWithPlans(serviceName)
	if err != nil {
		return err
	}

	org, err := actor.orgRepo.FindByName(orgName)
	if err != nil {
		return err
	}

	for _, plan := range service.Plans {
		switch {
		case plan.Public:
			continue
		case setPlanVisibility:
			err = actor.servicePlanVisibilityRepo.Create(plan.GUID, org.GUID)
			if err != nil {
				return err
			}
		case !setPlanVisibility:
			err = actor.deleteServicePlanVisibilities(map[string]string{"organization_guid": org.GUID, "service_plan_guid": plan.GUID})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (actor ServicePlanHandler) UpdatePlanAndOrgForService(serviceName, planName, orgName string, setPlanVisibility bool) error {
	service, err := actor.serviceBuilder.GetServiceByNameWithPlans(serviceName)
	if err != nil {
		return err
	}

	org, err := actor.orgRepo.FindByName(orgName)
	if err != nil {
		return err
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
		return fmt.Errorf("Service plan %s not found", planName)
	}

	switch {
	case servicePlan.Public:
		return nil
	case setPlanVisibility:
		// Enable service access
		err = actor.servicePlanVisibilityRepo.Create(servicePlan.GUID, org.GUID)
	case !setPlanVisibility:
		// Disable service access
		err = actor.deleteServicePlanVisibilities(map[string]string{"organization_guid": org.GUID, "service_plan_guid": servicePlan.GUID})
	}

	return err
}

func (actor ServicePlanHandler) UpdateSinglePlanForService(serviceName string, planName string, setPlanVisibility bool) error {
	serviceOffering, err := actor.serviceBuilder.GetServiceByNameWithPlans(serviceName)
	if err != nil {
		return err
	}
	return actor.updateSinglePlan(serviceOffering, planName, setPlanVisibility)
}

func (actor ServicePlanHandler) updateSinglePlan(serviceOffering models.ServiceOffering, planName string, setPlanVisibility bool) error {
	var planToUpdate *models.ServicePlanFields

	for _, servicePlan := range serviceOffering.Plans {
		if servicePlan.Name == planName {
			planToUpdate = &servicePlan
			break
		}
	}

	if planToUpdate == nil {
		return fmt.Errorf("The plan %s could not be found for service %s", planName, serviceOffering.Label)
	}

	return actor.updateServicePlanAvailability(serviceOffering.GUID, *planToUpdate, setPlanVisibility)
}

func (actor ServicePlanHandler) deleteServicePlanVisibilities(queryParams map[string]string) error {
	visibilities, err := actor.servicePlanVisibilityRepo.Search(queryParams)
	if err != nil {
		return err
	}

	for _, visibility := range visibilities {
		err = actor.servicePlanVisibilityRepo.Delete(visibility.GUID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (actor ServicePlanHandler) updateServicePlanAvailability(serviceGUID string, servicePlan models.ServicePlanFields, setPlanVisibility bool) error {
	// We delete all service plan visibilities for the given Plan since the attribute public should function as a giant on/off
	// switch for all orgs. Thus we need to clean up any visibilities laying around so that they don't carry over.
	err := actor.deleteServicePlanVisibilities(map[string]string{"service_plan_guid": servicePlan.GUID})
	if err != nil {
		return err
	}

	if servicePlan.Public == setPlanVisibility {
		return nil
	}

	return actor.servicePlanRepo.Update(servicePlan, serviceGUID, setPlanVisibility)
}

func (actor ServicePlanHandler) FindServiceAccess(serviceName string, orgName string) (ServiceAccess, error) {
	service, err := actor.serviceBuilder.GetServiceByNameForOrg(serviceName, orgName)
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

type PlanAccess int

const (
	PlanAccessError PlanAccess = iota
	All
	Limited
	None
)

func (actor ServicePlanHandler) findPlanAccess(plan models.ServicePlanFields) PlanAccess {
	if plan.Public {
		return All
	} else if len(plan.OrgNames) > 0 {
		return Limited
	} else {
		return None
	}
}
