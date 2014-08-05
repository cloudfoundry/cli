package actors

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/models"
)

type ServicePlanActor interface {
	UpdateAllPlansForService(string) (bool, error)
	UpdateSinglePlanForService(string, string, bool) (Access, error)
	UpdatePlanAndOrgForService(string, string, string, bool) (Access, error)
}

type Access int

const (
	Error Access = iota
	All
	Limited
	None
)

type ServicePlanHandler struct {
	serviceRepo               api.ServiceRepository
	servicePlanRepo           api.ServicePlanRepository
	servicePlanVisibilityRepo api.ServicePlanVisibilityRepository
	orgRepo                   api.OrganizationRepository
	serviceHandler            ServiceActor
}

func NewServicePlanHandler(service api.ServiceRepository, plan api.ServicePlanRepository, vis api.ServicePlanVisibilityRepository, org api.OrganizationRepository) ServicePlanHandler {
	serviceHandler := NewServiceHandler(nil, service, plan, vis, org)
	return ServicePlanHandler{
		serviceRepo:               service,
		servicePlanRepo:           plan,
		servicePlanVisibilityRepo: vis,
		orgRepo:                   org,
		serviceHandler:            serviceHandler,
	}
}

func (actor ServicePlanHandler) UpdateAllPlansForService(serviceName string) (bool, error) {
	service, err := actor.serviceRepo.FindServiceOfferingByLabel(serviceName)
	if err != nil {
		return false, err
	}

	service, err = actor.serviceHandler.AttachPlansToService(service)
	allPlansWerePublic := true
	for _, plan := range service.Plans {
		planAccess, err := actor.updateSinglePlan(service, plan.Name, true)
		if err != nil {
			return false, err
		}
		allPlansWerePublic = allPlansWerePublic && (planAccess == All)
	}
	return allPlansWerePublic, nil
}

func (actor ServicePlanHandler) UpdatePlanAndOrgForService(serviceName, planName, orgName string, setPlanVisibility bool) (Access, error) {
	serviceOffering, err := actor.serviceRepo.FindServiceOfferingByLabel(serviceName)
	if err != nil {
		return Error, err
	}

	org, err := actor.orgRepo.FindByName(orgName)
	if err != nil {
		return Error, err
	}

	servicePlans, err := actor.servicePlanRepo.Search(map[string]string{"service_guid": serviceOffering.Guid})
	if err != nil {
		return Error, err
	}

	servicePlans, err = actor.serviceHandler.AttachOrgsToPlans(servicePlans)
	if err != nil {
		return Error, err
	}

	found := false
	var servicePlan models.ServicePlanFields
	for i, val := range servicePlans {
		if val.Name == planName {
			found = true
			servicePlan = servicePlans[i]
		}
	}
	if !found {
		return Error, errors.New(fmt.Sprintf("Service plan %s not found", planName))
	}

	if !servicePlan.Public && setPlanVisibility {
		// Enable service access
		err = actor.servicePlanVisibilityRepo.Create(servicePlan.Guid, org.Guid)
		if err != nil {
			return Error, err
		}
	} else if !servicePlan.Public {
		// Disable service access
		if actor.checkPlanForOrgVisibility(servicePlan, org.Name) {
			err = actor.deleteServicePlanVisibility(servicePlan, org)
			if err != nil {
				return Error, err
			}
		}
	}

	access := actor.findPlanAccess(servicePlan)
	return access, nil
}

func (actor ServicePlanHandler) UpdateSinglePlanForService(serviceName string, planName string, setPlanVisibility bool) (Access, error) {
	serviceOffering, err := actor.serviceRepo.FindServiceOfferingByLabel(serviceName)
	if err != nil {
		return Error, err
	}
	return actor.updateSinglePlan(serviceOffering, planName, setPlanVisibility)
}

func (actor ServicePlanHandler) updateSinglePlan(serviceOffering models.ServiceOffering, planName string, setPlanVisibility bool) (Access, error) {
	var servicePlan models.ServicePlanFields

	//get all service plans for the one specific service
	//if there are no plans for the service it returns an empty set
	servicePlans, err := actor.servicePlanRepo.Search(map[string]string{"service_guid": serviceOffering.Guid})
	if err != nil {
		return Error, err
	}

	//find the service plan and set it as the only service plan for update
	serviceOffering.Plans = nil //set it to nil initially

	for _, servicePlan := range servicePlans {
		if servicePlan.Name == planName {
			serviceOffering.Plans = []models.ServicePlanFields{servicePlan} //he has the orgs inside him!!!
			break
		}
	}

	if serviceOffering.Plans == nil {
		return Error, errors.New(fmt.Sprintf("The plan %s could not be found for service %s", planName, serviceOffering.Label))
	} else {
		servicePlan = serviceOffering.Plans[0]
	}

	err = actor.updateServicePlanAvailability(serviceOffering.Guid, servicePlan, setPlanVisibility)
	if err != nil {
		return Error, err
	}

	access := actor.findPlanAccess(servicePlan)
	return access, nil
}

func (actor ServicePlanHandler) deleteServicePlanVisibility(servicePlan models.ServicePlanFields, org models.Organization) error {
	vis, err := actor.findServicePlanVisibility(servicePlan, org)
	if err != nil {
		return err
	}

	err = actor.servicePlanVisibilityRepo.Delete(vis.Guid)
	if err != nil {
		return err
	}
	return nil
}

func (actor ServicePlanHandler) checkPlanForOrgVisibility(servicePlan models.ServicePlanFields, orgName string) bool {
	for _, org := range servicePlan.OrgNames {
		if org == orgName {
			return true
		}
	}
	return false
}

func (actor ServicePlanHandler) updateServicePlanAvailability(serviceGuid string, servicePlan models.ServicePlanFields, setPlanVisibility bool) error {
	//delete service_plan_visibility guids[] and public: true
	err := actor.removeServicePlanVisibilities(servicePlan.Guid)
	if err != nil {
		return err
	}

	if servicePlan.Public == setPlanVisibility {
		return nil
	}

	return actor.servicePlanRepo.Update(servicePlan, serviceGuid, setPlanVisibility)
}

func (actor ServicePlanHandler) removeServicePlanVisibilities(servicePlanGuid string) error {
	planVisibilities, err := actor.servicePlanVisibilityRepo.List()
	if err != nil {
		return err
	}

	for _, planVisibility := range planVisibilities {
		if planVisibility.ServicePlanGuid == servicePlanGuid {
			err := actor.servicePlanVisibilityRepo.Delete(planVisibility.Guid)
			if err != nil {
				return err
			}
		}
		//else we simply skip it
	}

	return nil
}

func (actor ServicePlanHandler) findServicePlanVisibility(servicePlan models.ServicePlanFields, org models.Organization) (models.ServicePlanVisibilityFields, error) {
	visibilities, err := actor.servicePlanVisibilityRepo.List()
	if err != nil {
		return models.ServicePlanVisibilityFields{}, err
	}

	for _, vis := range visibilities {
		if vis.ServicePlanGuid == servicePlan.Guid && vis.OrganizationGuid == org.Guid {
			return vis, nil
		}
	}
	// We should never get here since we call checkPlanForOrgVisibility first.
	return models.ServicePlanVisibilityFields{}, nil
}

func (actor ServicePlanHandler) findPlanAccess(plan models.ServicePlanFields) Access {
	if plan.Public {
		return All
	} else if len(plan.OrgNames) > 0 {
		return Limited
	} else {
		return None
	}
}
