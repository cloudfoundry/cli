package actors

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/models"
)

type ServicePlanActor interface {
	GetServiceWithSinglePlan(string, string) (models.ServiceOffering, error)
	UpdateServicePlanAvailability(models.ServiceOffering, bool) error
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

//service
func (actor ServicePlanHandler) GetServiceWithAllPlans(serviceName string) (models.ServiceOffering, error) {
	return models.ServiceOffering{}, nil
}

// service -p
func (actor ServicePlanHandler) GetServiceWithSinglePlan(serviceName string, planName string) (models.ServiceOffering, error) {
	//find service guid
	serviceOffering, err := actor.serviceRepo.FindServiceOfferingByLabel(serviceName)
	if err != nil {
		return models.ServiceOffering{}, err
	}

	//get all service plans for the one specific service
	servicePlans, err := actor.servicePlanRepo.Search(map[string]string{"service_guid": serviceOffering.Guid})
	if err != nil {
		return models.ServiceOffering{}, err
	}

	//find the service plan and replace it
	for _, servicePlan := range servicePlans {
		if servicePlan.Name == planName {
			serviceOffering.Plans = []models.ServicePlanFields{servicePlan} //he has the org inside him!!!
		}
	}

	return serviceOffering, nil
}

//service -p -o
func (actor ServicePlanHandler) GetServiceWithSinglePlanAndOrg(serviceName, planName, orgName string) (models.ServiceOffering, error) {
	return models.ServiceOffering{}, nil
}

//service -o
func (actor ServicePlanHandler) GeServiceWithAllPlansSingleOrg(serviceName, orgName string) (models.ServiceOffering, error) {
	return models.ServiceOffering{}, nil
}

func (actor ServicePlanHandler) UpdateServicePlanAvailability(service models.ServiceOffering, public bool) error {
	//post to service_plan guids [] and public: true
	return actor.servicePlanRepo.Update(service.Plans[0], service.Guid, public)
}

func (actor ServicePlanHandler) RemoveServicePlanVisabilities(service models.ServiceOffering) error {
	planVisibilities := actor.servicePlanVisibilityRepo.List()

	for _, planVisability := range planVisibilities {
		if planVisability.Guid == service.Plan[0].Guid {
			err := actor.servicePlanVisibilityRepo.Delete(planVisability.Guid)
			if err != nil {
				return err
			}
		}
		//else we simply skip it
	}

	return nil
}

/*
func (actor ServicePlanHandler) GetSingleServicePlan(serviceName string, planName string) (models.ServicePlanFields, error) {
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
	var plan models.ServicePlanFields //he has the orgs inside him!!!!
	for _, servicePlan := range servicePlans {
		if servicePlan.Name == planName {

			serviceOffering.Plan = plan

			//plan = servicePlan
		}
	}

	return serviceOffering, nil
	//return plan, nil
}
*/
