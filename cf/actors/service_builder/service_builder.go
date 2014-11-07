package service_builder

import (
	"github.com/cloudfoundry/cli/cf/actors/plan_builder"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/models"
)

type ServiceBuilder interface {
	GetAllServices() ([]models.ServiceOffering, error)
	GetAllServicesWithPlans() ([]models.ServiceOffering, error)

	GetServiceByName(string) (models.ServiceOffering, error)
	GetServiceByNameForSpace(string, string) (models.ServiceOffering, error)
	GetServiceByNameForOrg(string, string) (models.ServiceOffering, error)

	GetServicesForBroker(string) ([]models.ServiceOffering, error)
	GetServicesForSpace(string) ([]models.ServiceOffering, error)

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

func (builder Builder) GetServiceByNameForOrg(serviceLabel, orgName string) (models.ServiceOffering, error) {
	service, err := builder.serviceRepo.FindServiceOfferingByLabel(serviceLabel)
	if err != nil {
		return models.ServiceOffering{}, err
	}
	service, err = builder.attachPlansToServiceForOrg(service, orgName)
	if err != nil {
		return models.ServiceOffering{}, err
	}
	return service, nil
}

func (builder Builder) GetAllServices() ([]models.ServiceOffering, error) {
	return builder.serviceRepo.GetAllServiceOfferings()
}

func (builder Builder) GetAllServicesWithPlans() ([]models.ServiceOffering, error) {
	services, err := builder.GetAllServices()
	if err != nil {
		return []models.ServiceOffering{}, err
	}

	var plans []models.ServicePlanFields
	for index, service := range services {
		plans, err = builder.planBuilder.GetPlansForService(service.Guid)
		if err != nil {
			return []models.ServiceOffering{}, err
		}
		services[index].Plans = plans
	}

	return services, err
}

func (builder Builder) GetServicesForSpace(spaceGuid string) ([]models.ServiceOffering, error) {
	return builder.serviceRepo.GetServiceOfferingsForSpace(spaceGuid)
}

func (builder Builder) GetServiceByNameForSpace(serviceLabel, spaceGuid string) (models.ServiceOffering, error) {
	serviceOfferings, err := builder.serviceRepo.GetServiceOfferingsForSpace(spaceGuid)
	if err != nil {
		return models.ServiceOffering{}, err
	}
	for _, offering := range serviceOfferings {
		if offering.Label == serviceLabel {
			return offering, nil
		}
	}
	return models.ServiceOffering{}, nil
}

func (builder Builder) GetServiceByName(serviceLabel string) (models.ServiceOffering, error) {
	service, err := builder.serviceRepo.FindServiceOfferingByLabel(serviceLabel)
	if err != nil {
		return models.ServiceOffering{}, err
	}
	service, err = builder.attachPlansToService(service)
	if err != nil {
		return models.ServiceOffering{}, err
	}
	return service, nil
}

func (builder Builder) GetServicesForBroker(brokerGuid string) ([]models.ServiceOffering, error) {
	services, err := builder.serviceRepo.ListServicesFromBroker(brokerGuid)
	if err != nil {
		return nil, err
	}
	for index, service := range services {
		services[index], err = builder.attachPlansToService(service)
		if err != nil {
			return nil, err
		}
	}
	return services, nil
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
	service, err := builder.serviceRepo.FindServiceOfferingByLabel(serviceName)
	if err != nil {
		return models.ServiceOffering{}, err
	}

	for _, plan := range plans {
		if plan.ServiceOfferingGuid == service.Guid {
			service.Plans = append(service.Plans, plan)
		}
	}

	return service, nil
}
