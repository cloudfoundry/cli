package actors

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/models"
)

type ServiceActor interface {
	GetAllBrokersWithDependencies() ([]models.ServiceBroker, error)
	GetBrokerWithDependencies(string) ([]models.ServiceBroker, error)
	GetBrokerWithSingleService(string) ([]models.ServiceBroker, error)
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

func (actor ServiceHandler) GetBrokerWithSingleService(serviceLabel string) ([]models.ServiceBroker, error) {
	service, err := actor.serviceRepo.FindServiceOfferingByLabel(serviceLabel)
	if err != nil {
		return nil, err
	}

	broker, err := actor.brokerRepo.FindByGuid(service.BrokerGuid)
	if err != nil {
		println("ERROR IN ACTOR IS THIS:" + err.Error())
		return nil, err
	}

	broker.Services = []models.ServiceOffering{service}
	brokers := []models.ServiceBroker{broker}

	brokers, err = actor.getServicePlans(brokers)
	if err != nil {
		return nil, err
	}

	return actor.getOrgs(brokers)
}

func (actor ServiceHandler) GetBrokerWithDependencies(brokerName string) ([]models.ServiceBroker, error) {
	broker, err := actor.brokerRepo.FindByName(brokerName)
	if err != nil {
		return nil, err
	}
	brokers := []models.ServiceBroker{broker}
	brokers, err = actor.getServices(brokers)
	if err != nil {
		return nil, err
	}

	brokers, err = actor.getServicePlans(brokers)
	if err != nil {
		return nil, err
	}

	return actor.getOrgs(brokers)
}

func (actor ServiceHandler) GetAllBrokersWithDependencies() ([]models.ServiceBroker, error) {
	brokers, err := actor.getAllServiceBrokers()
	if err != nil {
		return nil, err
	}

	brokers, err = actor.getServices(brokers)
	if err != nil {
		return nil, err
	}

	brokers, err = actor.getServicePlans(brokers)
	if err != nil {
		return nil, err
	}
	return actor.getOrgs(brokers)
}

func (actor ServiceHandler) getAllServiceBrokers() (brokers []models.ServiceBroker, err error) {
	err = actor.brokerRepo.ListServiceBrokers(func(broker models.ServiceBroker) bool {
		brokers = append(brokers, broker)
		return true
	})
	return
}

func (actor ServiceHandler) getServices(brokers []models.ServiceBroker) ([]models.ServiceBroker, error) {
	var err error
	for index, _ := range brokers {
		brokers[index].Services, err = actor.serviceRepo.ListServicesFromBroker(brokers[index].Guid)
		if err != nil {
			return nil, err
		}
	}
	return brokers, nil
}

func (actor ServiceHandler) getServicePlans(brokers []models.ServiceBroker) ([]models.ServiceBroker, error) {
	var err error
	//Is there a cleaner way to do this?
	for brokerIndex, _ := range brokers {
		broker := &brokers[brokerIndex]
		for serviceIndex, _ := range broker.Services {
			service := &broker.Services[serviceIndex]
			service.Plans, err = actor.servicePlanRepo.Search(map[string]string{"service_guid": service.Guid})
			if err != nil {
				return nil, err
			}
		}
	}
	return brokers, nil
}

func (actor ServiceHandler) getOrgs(brokers []models.ServiceBroker) ([]models.ServiceBroker, error) {
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

	//Is there a cleaner way to do this?
	for brokerIndex, _ := range brokers {
		broker := &brokers[brokerIndex]
		for serviceIndex, _ := range broker.Services {
			service := &broker.Services[serviceIndex]
			for planIndex, _ := range service.Plans {
				plan := &service.Plans[planIndex]
				plan.OrgNames = visMap[plan.Guid]
			}
		}
	}
	return brokers, nil
}
