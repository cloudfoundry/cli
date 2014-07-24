package actors

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/models"
)

type ServiceActor interface {
	GetAllBrokersWithDependencies() ([]models.ServiceBroker, error)
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

func (actor ServiceHandler) GetAllBrokersWithDependencies() ([]models.ServiceBroker, error) {
	brokers, err := actor.getServiceBrokers()
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

func (actor ServiceHandler) getServiceBrokers() (brokers []models.ServiceBroker, err error) {
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
