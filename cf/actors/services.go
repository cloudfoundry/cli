package actors

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/models"
)

type ServiceActor interface {
	FilterBrokers(brokerFlag string, serviceFlag string, orgFlag string) ([]models.ServiceBroker, error)
}

type ServiceHandler struct {
	brokerRepo                api.ServiceBrokerRepository
	serviceRepo               api.ServiceRepository
	servicePlanRepo           api.ServicePlanRepository
	servicePlanVisibilityRepo api.ServicePlanVisibilityRepository
	orgRepo                   api.OrganizationRepository
}

func NewServiceHandler(broker api.ServiceBrokerRepository, service api.ServiceRepository, plan api.ServicePlanRepository, vis api.ServicePlanVisibilityRepository, org api.OrganizationRepository) ServiceHandler {
	return ServiceHandler{
		brokerRepo:                broker,
		serviceRepo:               service,
		servicePlanRepo:           plan,
		servicePlanVisibilityRepo: vis,
		orgRepo:                   org,
	}
}

func (actor ServiceHandler) FilterBrokers(brokerFlag string, serviceFlag string, orgFlag string) ([]models.ServiceBroker, error) {
	if orgFlag == "" {
		return actor.buildBrokerTreeFromTop(brokerFlag, serviceFlag)
	} else {
		_, err := actor.orgRepo.FindByName(orgFlag)
		if err != nil {
			return nil, err
		}
		return actor.buildBrokerTreeFromBottom(brokerFlag, serviceFlag, orgFlag)
	}
}

func (actor ServiceHandler) buildBrokerTreeFromTop(brokerFlag string, serviceFlag string) ([]models.ServiceBroker, error) {
	var brokers []models.ServiceBroker
	var err error
	var broker models.ServiceBroker

	if brokerFlag != "" {
		broker, err = actor.brokerRepo.FindByName(brokerFlag)
		if err != nil {
			return nil, err
		}
		brokers = []models.ServiceBroker{broker}
	} else {
		brokers, err = actor.getAllServiceBrokers()
		if err != nil {
			return nil, err
		}
	}

	brokers, err = actor.attachServicesToBrokers(brokers, serviceFlag)
	if err != nil {
		return nil, err
	}

	//Prune brokers with no services
	brokersToReturn := []models.ServiceBroker{}
	for index, _ := range brokers {
		if len(brokers[index].Services) > 0 {
			brokersToReturn = append(brokersToReturn, brokers[index])
		}
	}

	return brokersToReturn, nil
}

func (actor ServiceHandler) attachServicesToBrokers(brokers []models.ServiceBroker, serviceFlag string) ([]models.ServiceBroker, error) {
	var err error
	var service models.ServiceOffering

	serviceFlagEnabled := serviceFlag != ""

	if serviceFlagEnabled {
		service, err = actor.serviceRepo.FindServiceOfferingByLabel(serviceFlag)
		if err != nil {
			return nil, err
		}
	}

	for index, _ := range brokers {
		if serviceFlagEnabled {
			if brokers[index].Guid == service.BrokerGuid {
				//check to see if its guid is contained in the list of the broker's service guid
				brokers[index].Services, err = actor.attachPlansToServices([]models.ServiceOffering{service})
				if err != nil {
					return nil, err
				}
				return brokers, nil
			} else {
				continue
			}
		}
		services, err := actor.serviceRepo.ListServicesFromBroker(brokers[index].Guid)
		if err != nil {
			return nil, err
		}
		brokers[index].Services, err = actor.attachPlansToServices(services)
		if err != nil {
			return nil, err
		}
	}
	return brokers, nil
}

func (actor ServiceHandler) attachPlansToServices(services []models.ServiceOffering) ([]models.ServiceOffering, error) {
	for serviceIndex, _ := range services {
		service := &services[serviceIndex]
		plans, err := actor.servicePlanRepo.Search(map[string]string{"service_guid": service.Guid})
		if err != nil {
			return nil, err
		}
		service.Plans, err = actor.attachOrgsToPlans(plans)
		if err != nil {
			return nil, err
		}
	}
	return services, nil
}

func (actor ServiceHandler) attachOrgsToPlans(plans []models.ServicePlanFields) ([]models.ServicePlanFields, error) {
	visMap, err := actor.buildPlanToOrgsVisibilityMap()
	if err != nil {
		return nil, err
	}
	for planIndex, _ := range plans {
		plan := &plans[planIndex]
		plan.OrgNames = visMap[plan.Guid]
	}

	return plans, nil
}

func (actor ServiceHandler) buildBrokerTreeFromBottom(brokerFlag string, serviceFlag string, orgFlag string) ([]models.ServiceBroker, error) {
	var err error
	var service models.ServiceOffering

	serviceToVisiblePlansMap, err := actor.createMapOfServicesToVisiblePlans(orgFlag)
	if serviceFlag != "" {
		service, err = actor.serviceRepo.FindServiceOfferingByLabel(serviceFlag)
		if err != nil {
			return nil, err
		}
		serviceToFilter, ok := serviceToVisiblePlansMap[service.Guid]
		if !ok {
			// Service is not visible to Org
			return nil, nil
		}
		serviceMap := make(map[string][]models.ServicePlanFields)
		serviceMap[service.Guid] = serviceToFilter
		serviceToVisiblePlansMap = serviceMap
	}

	brokers, err := actor.getAllBrokersFromServicesMap(serviceToVisiblePlansMap)
	if err != nil {
		return nil, err
	}

	if brokerFlag != "" {
		for brokerIndex, _ := range brokers {
			broker := &brokers[brokerIndex]
			if broker.Name == brokerFlag {
				return []models.ServiceBroker{brokers[brokerIndex]}, nil
			}
		}
		// Could not find brokerFlag in visible brokers.
		return nil, nil
	}

	return brokers, nil
}

func (actor ServiceHandler) createMapOfServicesToVisiblePlans(orgName string) (map[string][]models.ServicePlanFields, error) {
	allPlans, err := actor.servicePlanRepo.Search(nil)
	if err != nil {
		return nil, err
	}

	servicesToPlansMap := make(map[string][]models.ServicePlanFields)
	PlanToOrgsVisMap, err := actor.buildPlanToOrgsVisibilityMap()
	if err != nil {
		return nil, err
	}
	OrgToPlansVisMap := actor.buildOrgToPlansVisibilityMap(PlanToOrgsVisMap)
	filterOrgPlans := OrgToPlansVisMap[orgName]

	for _, plan := range allPlans {
		if actor.containsGuid(filterOrgPlans, plan.Guid) {
			plan.OrgNames = PlanToOrgsVisMap[plan.Guid]
			servicesToPlansMap[plan.ServiceOfferingGuid] = append(servicesToPlansMap[plan.ServiceOfferingGuid], plan)
		} else if plan.Public {
			servicesToPlansMap[plan.ServiceOfferingGuid] = append(servicesToPlansMap[plan.ServiceOfferingGuid], plan)
		}
	}

	return servicesToPlansMap, nil
}

func (actor ServiceHandler) getAllBrokersFromServicesMap(serviceMap map[string][]models.ServicePlanFields) ([]models.ServiceBroker, error) {
	var brokers []models.ServiceBroker
	brokersToServices := make(map[string][]models.ServiceOffering)

	for serviceGuid, plans := range serviceMap {
		service, err := actor.serviceRepo.GetServiceOfferingByGuid(serviceGuid)
		if err != nil {
			return nil, err
		}
		service.Plans = plans
		brokersToServices[service.BrokerGuid] = append(brokersToServices[service.BrokerGuid], service)
	}

	for brokerGuid, services := range brokersToServices {
		if brokerGuid == "" {
			continue
		}
		broker, err := actor.brokerRepo.FindByGuid(brokerGuid)
		if err != nil {
			return nil, err
		}

		broker.Services = services
		brokers = append(brokers, broker)
	}

	return brokers, nil
}

func (actor ServiceHandler) containsGuid(guidSlice []string, guid string) bool {
	for _, g := range guidSlice {
		if g == guid {
			return true
		}
	}
	return false
}

func (actor ServiceHandler) getAllServiceBrokers() (brokers []models.ServiceBroker, err error) {
	err = actor.brokerRepo.ListServiceBrokers(func(broker models.ServiceBroker) bool {
		brokers = append(brokers, broker)
		return true
	})
	return
}

func (actor ServiceHandler) buildPlanToOrgsVisibilityMap() (map[string][]string, error) {
	orgLookup := make(map[string]string)
	actor.orgRepo.ListOrgs(func(org models.Organization) bool {
		orgLookup[org.Guid] = org.Name
		return true
	})

	visibilities, err := actor.servicePlanVisibilityRepo.List()
	if err != nil {
		return nil, err
	}

	visMap := make(map[string][]string)
	for _, vis := range visibilities {
		visMap[vis.ServicePlanGuid] = append(visMap[vis.ServicePlanGuid], orgLookup[vis.OrganizationGuid])
	}

	return visMap, nil
}

func (actor ServiceHandler) buildOrgToPlansVisibilityMap(planToOrgsMap map[string][]string) map[string][]string {
	visMap := make(map[string][]string)
	for planGuid, orgNames := range planToOrgsMap {
		for _, orgName := range orgNames {
			visMap[orgName] = append(visMap[orgName], planGuid)
		}
	}

	return visMap
}
