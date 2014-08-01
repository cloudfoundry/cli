package actors

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/models"
)

type ServicePlanActor interface {
	UpdateSinglePlanForService(string, string) (bool, error)
}

type ServicePlanHandler struct {
	serviceRepo               api.ServiceRepository
	servicePlanRepo           api.ServicePlanRepository
	servicePlanVisibilityRepo api.ServicePlanVisibilityRepository
	orgRepo                   api.OrganizationRepository
}

func NewServicePlanHandler(service api.ServiceRepository, plan api.ServicePlanRepository, vis api.ServicePlanVisibilityRepository, org api.OrganizationRepository) ServicePlanHandler {
	return ServicePlanHandler{
		serviceRepo:               service,
		servicePlanRepo:           plan,
		servicePlanVisibilityRepo: vis,
		orgRepo:                   org,
	}
}

func (actor ServicePlanHandler) UpdateSinglePlanForService(serviceName string, planName string) (bool, error) {
	var servicePlan models.ServicePlanFields

	serviceOffering, err := actor.serviceRepo.FindServiceOfferingByLabel(serviceName)
	if err != nil {
		return false, err
	}

	//get all service plans for the one specific service
	//if there are no plans for the service it returns an empty set
	servicePlans, err := actor.servicePlanRepo.Search(map[string]string{"service_guid": serviceOffering.Guid})
	if err != nil {
		return false, err
	}

	//find the service plan and set it as the only service plan for update
	serviceOffering.Plans = nil //set it to nil initialy
	for _, servicePlan := range servicePlans {
		if servicePlan.Name == planName {
			serviceOffering.Plans = []models.ServicePlanFields{servicePlan} //he has the orgs inside him!!!
			break
		}
	}

	if serviceOffering.Plans == nil {
		return false, errors.New(fmt.Sprintf("The plan %s could not be found for service %s", planName, serviceName))
	} else {
		servicePlan = serviceOffering.Plans[0]
	}

	err = actor.updateServicePlanAvailability(serviceOffering.Guid, servicePlan, true)
	if err != nil {
		return false, err
	}

	return servicePlan.Public, nil
}

func (actor ServicePlanHandler) updateServicePlanAvailability(serviceGuid string, servicePlan models.ServicePlanFields, public bool) error {
	//delete service_plan_visibility guids[] and public: true
	err := actor.removeServicePlanVisibilities(servicePlan.Guid)
	if err != nil {
		return err
	}

	if servicePlan.Public {
		return nil
	}

	return actor.servicePlanRepo.Update(servicePlan, serviceGuid, public)
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
