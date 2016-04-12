package service_builder

import (
	"errors"

	"github.com/cloudfoundry/cli/cf/actors/plan_builder"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/models"
)

//go:generate counterfeiter . ServiceBuilder

type ServiceBuilder interface {
	GetAllServices() ([]models.ServiceOffering, error)
	GetAllServicesWithPlans() ([]models.ServiceOffering, error)

	GetServiceByNameWithPlans(string) (models.ServiceOffering, error)
	GetServiceByNameWithPlansWithOrgNames(string) (models.ServiceOffering, error)
	GetServiceByNameForSpace(string, string) (models.ServiceOffering, error)
	GetServiceByNameForSpaceWithPlans(string, string) (models.ServiceOffering, error)
	GetServicesByNameForSpaceWithPlans(string, string) (models.ServiceOfferings, error)
	GetServiceByNameForOrg(string, string) (models.ServiceOffering, error)

	GetServicesForManyBrokers([]string) ([]models.ServiceOffering, error)
	GetServicesForBroker(string) ([]models.ServiceOffering, error)

	GetServicesForSpace(string) ([]models.ServiceOffering, error)
	GetServicesForSpaceWithPlans(string) ([]models.ServiceOffering, error)

	GetServiceVisibleToOrg(string, string) (models.ServiceOffering, error)
	GetServicesVisibleToOrg(string) ([]models.ServiceOffering, error)
}

type Builder struct {
	serviceRepo api.ServiceRepository
	planBuilder plan_builder.PlanBuilder
}

func NewBuilder(service api.ServiceRepository, planBuilder plan_builder.PlanBuilder) Builder {
	return Builder{
		serviceRepo: service,
		planBuilder: planBuilder,
	}
}

func (builder Builder) GetAllServices() ([]models.ServiceOffering, error) {
	return builder.serviceRepo.GetAllServiceOfferings()
}

func (builder Builder) GetAllServicesWithPlans() ([]models.ServiceOffering, error) {
	services, err := builder.GetAllServices()
	if err != nil {
		return []models.ServiceOffering{}, err
	}
    
    
    
	return builder.getPlansForServices(services)
}

func (builder Builder) GetServicesForSpace(spaceGuid string) ([]models.ServiceOffering, error) {
	return builder.serviceRepo.GetServiceOfferingsForSpace(spaceGuid)
}

func (builder Builder) GetServicesForSpaceWithPlans(spaceGuid string) ([]models.ServiceOffering, error) {
    
	services, err := builder.GetServicesForSpace(spaceGuid)
	if err != nil {
		return []models.ServiceOffering{}, err
	}

    return builder.getPlansForServices(services)
}

func (builder Builder) GetServiceByNameWithPlans(serviceLabel string) (models.ServiceOffering, error) {
	services, err := builder.serviceRepo.FindServiceOfferingsByLabel(serviceLabel)
	if err != nil {
		return models.ServiceOffering{}, err
	}
	service := returnV2Service(services)

	service.Plans, err = builder.planBuilder.GetPlansForService(service.Guid)
	if err != nil {
		return models.ServiceOffering{}, err
	}

	return service, nil
}

func (builder Builder) GetServiceByNameForOrg(serviceLabel, orgName string) (models.ServiceOffering, error) {
	services, err := builder.serviceRepo.FindServiceOfferingsByLabel(serviceLabel)
	if err != nil {
		return models.ServiceOffering{}, err
	}

	service, err := builder.attachPlansToServiceForOrg(services[0], orgName)
	if err != nil {
		return models.ServiceOffering{}, err
	}
	return service, nil
}

func (builder Builder) GetServiceByNameForSpace(serviceLabel, spaceGuid string) (models.ServiceOffering, error) {
	offerings, err := builder.serviceRepo.FindServiceOfferingsForSpaceByLabel(spaceGuid, serviceLabel)
	if err != nil {
		return models.ServiceOffering{}, err
	}

	for _, offering := range offerings {
		if offering.Provider == "" {
			return offering, nil
		}
	}

	return models.ServiceOffering{}, errors.New("Could not find service")
}

func (builder Builder) GetServiceByNameForSpaceWithPlans(serviceLabel, spaceGuid string) (models.ServiceOffering, error) {
	offering, err := builder.GetServiceByNameForSpace(serviceLabel, spaceGuid)
	if err != nil {
		return models.ServiceOffering{}, err
	}

	offering.Plans, err = builder.planBuilder.GetPlansForService(offering.Guid)
	if err != nil {
		return models.ServiceOffering{}, err
	}

	return offering, nil
}

func (builder Builder) GetServicesByNameForSpaceWithPlans(serviceLabel, spaceGuid string) (models.ServiceOfferings, error) {
	offerings, err := builder.serviceRepo.FindServiceOfferingsForSpaceByLabel(serviceLabel, spaceGuid)
	if err != nil {
		return models.ServiceOfferings{}, err
	}

	for index, offering := range offerings {
		offerings[index].Plans, err = builder.planBuilder.GetPlansForService(offering.Guid)
		if err != nil {
			return models.ServiceOfferings{}, err
		}
	}

	return offerings, nil
}

func (builder Builder) GetServiceByNameWithPlansWithOrgNames(serviceLabel string) (models.ServiceOffering, error) {
	services, err := builder.serviceRepo.FindServiceOfferingsByLabel(serviceLabel)
	if err != nil {
		return models.ServiceOffering{}, err
	}

	service, err := builder.attachPlansToService(services[0])
	if err != nil {
		return models.ServiceOffering{}, err
	}
	return service, nil
}

func (builder Builder) GetServicesForManyBrokers(brokerGuids []string) ([]models.ServiceOffering, error) {
	services, err := builder.serviceRepo.ListServicesFromManyBrokers(brokerGuids)
	if err != nil {
		return nil, err
	}
	return builder.populateServicesWithPlansAndOrgs(services)
}

func (builder Builder) GetServicesForBroker(brokerGuid string) ([]models.ServiceOffering, error) {
	services, err := builder.serviceRepo.ListServicesFromBroker(brokerGuid)
	if err != nil {
		return nil, err
	}
	return builder.populateServicesWithPlansAndOrgs(services)
}

func (builder Builder) populateServicesWithPlansAndOrgs(services []models.ServiceOffering) ([]models.ServiceOffering, error) {
	serviceGuids := []string{}
	for _, service := range services {
		serviceGuids = append(serviceGuids, service.Guid)
	}

	plans, err := builder.planBuilder.GetPlansForManyServicesWithOrgs(serviceGuids)
	if err != nil {
		return nil, err
	}
	return builder.attachPlansToManyServices(services, plans)
}

func (builder Builder) GetServiceVisibleToOrg(serviceName string, orgName string) (models.ServiceOffering, error) {
	visiblePlans, err := builder.planBuilder.GetPlansVisibleToOrg(orgName)
	if err != nil {
		return models.ServiceOffering{}, err
	}

	if len(visiblePlans) == 0 {
		return models.ServiceOffering{}, nil
	}

	return builder.attachSpecificServiceToPlans(serviceName, visiblePlans)
}

func (builder Builder) GetServicesVisibleToOrg(orgName string) ([]models.ServiceOffering, error) {
	visiblePlans, err := builder.planBuilder.GetPlansVisibleToOrg(orgName)
	if err != nil {
		return nil, err
	}

	if len(visiblePlans) == 0 {
		return nil, nil
	}

	return builder.attachServicesToPlans(visiblePlans)
}

func (builder Builder) attachPlansToServiceForOrg(service models.ServiceOffering, orgName string) (models.ServiceOffering, error) {
	plans, err := builder.planBuilder.GetPlansForServiceForOrg(service.Guid, orgName)
	if err != nil {
		return models.ServiceOffering{}, err
	}

	service.Plans = plans
	return service, nil
}

func (builder Builder) attachPlansToManyServices(services []models.ServiceOffering, plans []models.ServicePlanFields) ([]models.ServiceOffering, error) {
	for _, plan := range plans {
		for index, service := range services {
			if service.Guid == plan.ServiceOfferingGuid {
				services[index].Plans = append(service.Plans, plan)
				break
			}
		}
	}
	return services, nil
}

func (builder Builder) attachPlansToService(service models.ServiceOffering) (models.ServiceOffering, error) {
	plans, err := builder.planBuilder.GetPlansForServiceWithOrgs(service.Guid)
	if err != nil {
		return models.ServiceOffering{}, err
	}

	service.Plans = plans
	return service, nil
}

func (builder Builder) attachServicesToPlans(plans []models.ServicePlanFields) ([]models.ServiceOffering, error) {
	var services []models.ServiceOffering
	servicesMap := make(map[string]models.ServiceOffering)

	for _, plan := range plans {
		if plan.ServiceOfferingGuid == "" {
			continue
		}

		if service, ok := servicesMap[plan.ServiceOfferingGuid]; ok {
			service.Plans = append(service.Plans, plan)
			servicesMap[service.Guid] = service
		} else {
			service, err := builder.serviceRepo.GetServiceOfferingByGuid(plan.ServiceOfferingGuid)
			if err != nil {
				return nil, err
			}
			service.Plans = append(service.Plans, plan)
			servicesMap[service.Guid] = service
		}
	}

	for _, service := range servicesMap {
		services = append(services, service)
	}

	return services, nil
}

func (builder Builder) attachSpecificServiceToPlans(serviceName string, plans []models.ServicePlanFields) (models.ServiceOffering, error) {
	services, err := builder.serviceRepo.FindServiceOfferingsByLabel(serviceName)
	if err != nil {
		return models.ServiceOffering{}, err
	}

	service := services[0]
	for _, plan := range plans {
		if plan.ServiceOfferingGuid == service.Guid {
			service.Plans = append(service.Plans, plan)
		}
	}

	return service, nil
}

func returnV2Service(services models.ServiceOfferings) models.ServiceOffering {
	for _, service := range services {
		if service.Provider == "" {
			return service
		}
	}

	return models.ServiceOffering{}
}

func (builder Builder) getPlansForServices(services []models.ServiceOffering) ([]models.ServiceOffering, error) {
    type planResponse struct {
        index int
        plans []models.ServicePlanFields
        err error
    }
    
    resChan := make(chan *planResponse);
    
	for index, service := range services {
        go func(resChan chan *planResponse, guid string, index int) {
            res := &planResponse{
                index: index,
            }
            res.plans, res.err = builder.planBuilder.GetPlansForService(guid)
            resChan <- res                
        }(resChan, service.Guid, index)
	}
    
    for responseCount := 0; responseCount < len(services); {
        res := <- resChan
        if res.err != nil {
            return services, res.err
        }
        services[res.index].Plans = res.plans
        responseCount++
    }
    
	return services, nil
} 