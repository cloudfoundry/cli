package broker_builder

import (
	"github.com/cloudfoundry/cli/cf/actors/service_builder"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/models"
)

type BrokerBuilder interface {
	AttachBrokersToServices([]models.ServiceOffering) ([]models.ServiceBroker, error)
	AttachSpecificBrokerToServices(string, []models.ServiceOffering) (models.ServiceBroker, error)
	GetAllServiceBrokers() ([]models.ServiceBroker, error)
	GetBrokerWithAllServices(brokerName string) (models.ServiceBroker, error)
	GetBrokerWithSpecifiedService(serviceName string) (models.ServiceBroker, error)
}

type Builder struct {
	brokerRepo     api.ServiceBrokerRepository
	serviceBuilder service_builder.ServiceBuilder
}

func NewBuilder(broker api.ServiceBrokerRepository, serviceBuilder service_builder.ServiceBuilder) Builder {
	return Builder{
		brokerRepo:     broker,
		serviceBuilder: serviceBuilder,
	}
}

func (builder Builder) AttachBrokersToServices(services []models.ServiceOffering) ([]models.ServiceBroker, error) {
	var brokers []models.ServiceBroker
	brokersMap := make(map[string]models.ServiceBroker)

	for _, service := range services {
		if service.BrokerGuid == "" {
			continue
		}

		if broker, ok := brokersMap[service.BrokerGuid]; ok {
			broker.Services = append(broker.Services, service)
			brokersMap[broker.Guid] = broker
		} else {
			broker, err := builder.brokerRepo.FindByGuid(service.BrokerGuid)
			if err != nil {
				return nil, err
			}
			broker.Services = append(broker.Services, service)
			brokersMap[service.BrokerGuid] = broker
		}
	}

	for _, broker := range brokersMap {
		brokers = append(brokers, broker)
	}

	return brokers, nil
}

func (builder Builder) AttachSpecificBrokerToServices(brokerName string, services []models.ServiceOffering) (models.ServiceBroker, error) {
	broker, err := builder.brokerRepo.FindByName(brokerName)
	if err != nil {
		return models.ServiceBroker{}, err
	}

	for _, service := range services {
		if service.BrokerGuid == broker.Guid {
			broker.Services = append(broker.Services, service)
		}
	}

	return broker, nil
}

func (builder Builder) GetAllServiceBrokers() ([]models.ServiceBroker, error) {
	brokers := []models.ServiceBroker{}
	var err error
	var services models.ServiceOfferings

	err = builder.brokerRepo.ListServiceBrokers(func(broker models.ServiceBroker) bool {
		brokers = append(brokers, broker)
		return true
	})

	for index, broker := range brokers {
		services, err = builder.serviceBuilder.GetServicesForBroker(broker.Guid)
		if err != nil {
			return nil, err
		}

		brokers[index].Services = services
	}
	return brokers, err
}

func (builder Builder) GetBrokerWithAllServices(brokerName string) (models.ServiceBroker, error) {
	broker, err := builder.brokerRepo.FindByName(brokerName)
	if err != nil {
		return models.ServiceBroker{}, err
	}
	services, err := builder.serviceBuilder.GetServicesForBroker(broker.Guid)
	if err != nil {
		return models.ServiceBroker{}, err
	}
	broker.Services = services

	return broker, nil
}

func (builder Builder) GetBrokerWithSpecifiedService(serviceName string) (models.ServiceBroker, error) {
	service, err := builder.serviceBuilder.GetServiceByName(serviceName)
	if err != nil {
		return models.ServiceBroker{}, err
	}
	brokers, err := builder.AttachBrokersToServices([]models.ServiceOffering{service})
	if err != nil || len(brokers) == 0 {
		return models.ServiceBroker{}, err
	}
	return brokers[0], err
}
