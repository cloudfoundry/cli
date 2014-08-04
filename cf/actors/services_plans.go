package actors

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/cli/cf/actors/service_builder"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/models"
)

type ServicePlanActor interface {
	UpdateAllPlansForService(string, bool) (bool, error)
	UpdateSinglePlanForService(string, string, bool) (bool, error)
}

type ServicePlanHandler struct {
	serviceRepo               api.ServiceRepository
	servicePlanRepo           api.ServicePlanRepository
	servicePlanVisibilityRepo api.ServicePlanVisibilityRepository
	orgRepo                   api.OrganizationRepository
	serviceBuilder            service_builder.ServiceBuilder
}

func NewServicePlanHandler(service api.ServiceRepository, plan api.ServicePlanRepository, vis api.ServicePlanVisibilityRepository, org api.OrganizationRepository) ServicePlanHandler {
	serviceBuilder := service_builder.NewBuilder(service, plan, vis, org)
	return ServicePlanHandler{
		serviceRepo:               service,
		servicePlanRepo:           plan,
		servicePlanVisibilityRepo: vis,
		orgRepo:                   org,
		serviceBuilder:            serviceBuilder,
	}
}

func (actor ServicePlanHandler) UpdateAllPlansForService(serviceName string, setPlanVisibility bool) (bool, error) {
	service, err := actor.serviceRepo.FindServiceOfferingByLabel(serviceName)
	if err != nil {
		return false, err
	}

	service, err = actor.serviceBuilder.AttachPlansToService(service)
	allPlansWerePublic := true
	for _, plan := range service.Plans {
		planWasPublic, err := actor.updateSinglePlan(service, plan.Name, setPlanVisibility)
		if err != nil {
			return false, err
		}
		allPlansWerePublic = allPlansWerePublic && planWasPublic
	}
	return allPlansWerePublic, nil
}

func (actor ServicePlanHandler) UpdateSinglePlanForService(serviceName string, planName string, setPlanVisibility bool) (bool, error) {
	serviceOffering, err := actor.serviceRepo.FindServiceOfferingByLabel(serviceName)
	if err != nil {
		return false, err
	}
	return actor.updateSinglePlan(serviceOffering, planName, setPlanVisibility)
}

func (actor ServicePlanHandler) updateSinglePlan(serviceOffering models.ServiceOffering, planName string, setPlanVisibility bool) (bool, error) {
	var servicePlan models.ServicePlanFields

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
		return false, errors.New(fmt.Sprintf("The plan %s could not be found for service %s", planName, serviceOffering.Label))
	} else {
		servicePlan = serviceOffering.Plans[0]
	}

	err = actor.updateServicePlanAvailability(serviceOffering.Guid, servicePlan, setPlanVisibility)
	if err != nil {
		return false, err
	}

	return servicePlan.Public, nil
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
