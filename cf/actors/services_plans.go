package actors

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/models"
)

type ServicePlanActor interface {
	GetSingleServicePlanForService(string, string) (models.ServicePlanFields, error)
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

func (actor ServicePlanHandler) GetSingleServicePlanForService(serviceName string, planName string) (models.ServicePlanFields, error) {
	//find service guid
	serviceOffering, err := actor.serviceRepo.FindServiceOfferingByLabel(serviceName)
	if err != nil {
		return models.ServicePlanFields{}, err
	}

	//get all service plans for the service
	servicePlans, err := actor.servicePlanRepo.Search(map[string]string{"service_guid": serviceOffering.Guid})
	if err != nil {
		return models.ServicePlanFields{}, err
	}

	//find the service plan
	var plan models.ServicePlanFields
	for _, servicePlan := range servicePlans {
		if servicePlan.Name == planName {
			plan = servicePlan
		}
	}

	return plan, nil
}
