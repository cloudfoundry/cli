package service_builder

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/models"
)

type ServiceBuilder interface {
	AttachPlansToService(models.ServiceOffering) (models.ServiceOffering, error)

	GetServiceByName(string) ([]models.ServiceOffering, error)
	GetServicesForBroker(string) ([]models.ServiceOffering, error)

	GetServiceVisibleToOrg(string, string) ([]models.ServiceOffering, error)
	GetServicesVisibleToOrg(string) ([]models.ServiceOffering, error)
}

type Builder struct {
	serviceRepo               api.ServiceRepository
	servicePlanRepo           api.ServicePlanRepository
	servicePlanVisibilityRepo api.ServicePlanVisibilityRepository
	orgRepo                   api.OrganizationRepository
}

func NewBuilder(service api.ServiceRepository, plan api.ServicePlanRepository, vis api.ServicePlanVisibilityRepository, org api.OrganizationRepository) Builder {
	return Builder{
		serviceRepo:               service,
		servicePlanRepo:           plan,
		servicePlanVisibilityRepo: vis,
		orgRepo:                   org,
	}
}

func (builder Builder) AttachPlansToService(service models.ServiceOffering) (models.ServiceOffering, error) {
	plans, err := builder.servicePlanRepo.Search(map[string]string{"service_guid": service.Guid})
	if err != nil {
		return models.ServiceOffering{}, err
	}

	service.Plans, err = builder.attachOrgsToPlans(plans)
	if err != nil {
		return models.ServiceOffering{}, err
	}
	return service, nil
}

func (builder Builder) GetServiceByName(serviceLabel string) ([]models.ServiceOffering, error) {
	service, err := builder.serviceRepo.FindServiceOfferingByLabel(serviceLabel)
	if err != nil {
		return nil, err
	}
	service, err = builder.AttachPlansToService(service)
	if err != nil {
		return nil, err
	}
	return []models.ServiceOffering{service}, nil
}

func (builder Builder) GetServicesForBroker(brokerGuid string) ([]models.ServiceOffering, error) {
	services, err := builder.serviceRepo.ListServicesFromBroker(brokerGuid)
	if err != nil {
		return nil, err
	}
	for index, service := range services {
		services[index], err = builder.AttachPlansToService(service)
		if err != nil {
			return nil, err
		}
	}
	return services, nil
}

func (builder Builder) GetServiceVisibleToOrg(serviceName string, orgName string) ([]models.ServiceOffering, error) {
	serviceToVisiblePlansMap, err := builder.buildServicesToVisiblePlansMap(orgName)
	if err != nil {
		return nil, err
	}

	service, err := builder.serviceRepo.FindServiceOfferingByLabel(serviceName)
	if err != nil {
		return nil, err
	}

	plans, ok := serviceToVisiblePlansMap[service.Guid]
	if !ok {
		// Service is not visible to Org
		return nil, nil
	}

	service.Plans = plans
	return []models.ServiceOffering{service}, nil
}

func (builder Builder) GetServicesVisibleToOrg(orgName string) ([]models.ServiceOffering, error) {
	var services []models.ServiceOffering

	serviceToVisiblePlansMap, err := builder.buildServicesToVisiblePlansMap(orgName)
	if err != nil {
		return nil, err
	}

	for serviceGuid, plans := range serviceToVisiblePlansMap {
		service, err := builder.serviceRepo.GetServiceOfferingByGuid(serviceGuid)
		if err != nil {
			return nil, err
		}
		service.Plans = plans
		services = append(services, service)
	}

	return services, nil
}

func (builder Builder) attachOrgsToPlans(plans []models.ServicePlanFields) ([]models.ServicePlanFields, error) {
	visMap, err := builder.buildPlanToOrgsVisibilityMap()
	if err != nil {
		return nil, err
	}
	for planIndex, _ := range plans {
		plan := &plans[planIndex]
		plan.OrgNames = visMap[plan.Guid]
	}

	return plans, nil
}

func (builder Builder) buildOrgToPlansVisibilityMap(planToOrgsMap map[string][]string) map[string][]string {
	visMap := make(map[string][]string)
	for planGuid, orgNames := range planToOrgsMap {
		for _, orgName := range orgNames {
			visMap[orgName] = append(visMap[orgName], planGuid)
		}
	}

	return visMap
}

func (builder Builder) buildPlanToOrgsVisibilityMap() (map[string][]string, error) {
	//WE PROBABLY HAVE A TERRIBLE PERFORMANCE PROBLEM HERE
	//WE SHOULD MEMOIZE THESE MAPS
	orgLookup := make(map[string]string)
	builder.orgRepo.ListOrgs(func(org models.Organization) bool {
		orgLookup[org.Guid] = org.Name
		return true
	})

	visibilities, err := builder.servicePlanVisibilityRepo.List()
	if err != nil {
		return nil, err
	}

	visMap := make(map[string][]string)
	for _, vis := range visibilities {
		visMap[vis.ServicePlanGuid] = append(visMap[vis.ServicePlanGuid], orgLookup[vis.OrganizationGuid])
	}

	return visMap, nil
}

func (builder Builder) buildServicesToVisiblePlansMap(orgName string) (map[string][]models.ServicePlanFields, error) {
	allPlans, err := builder.servicePlanRepo.Search(nil)
	if err != nil {
		return nil, err
	}

	servicesToPlansMap := make(map[string][]models.ServicePlanFields)
	PlanToOrgsVisMap, err := builder.buildPlanToOrgsVisibilityMap()
	if err != nil {
		return nil, err
	}
	OrgToPlansVisMap := builder.buildOrgToPlansVisibilityMap(PlanToOrgsVisMap)
	filterOrgPlans := OrgToPlansVisMap[orgName]

	for _, plan := range allPlans {
		if builder.containsGuid(filterOrgPlans, plan.Guid) {
			plan.OrgNames = PlanToOrgsVisMap[plan.Guid]
			servicesToPlansMap[plan.ServiceOfferingGuid] = append(servicesToPlansMap[plan.ServiceOfferingGuid], plan)
		} else if plan.Public {
			servicesToPlansMap[plan.ServiceOfferingGuid] = append(servicesToPlansMap[plan.ServiceOfferingGuid], plan)
		}
	}

	return servicesToPlansMap, nil
}

func (builder Builder) containsGuid(guidSlice []string, guid string) bool {
	for _, g := range guidSlice {
		if g == guid {
			return true
		}
	}
	return false
}
