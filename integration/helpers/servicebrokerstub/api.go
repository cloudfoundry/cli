package servicebrokerstub

import (
	"code.cloudfoundry.org/cli/integration/assets/hydrabroker/config"
)

type ServiceBrokerStub struct {
	Name, URL, GUID    string
	Username, Password string
	Services           []config.Service
}

func New() *ServiceBrokerStub {
	return newDefaultConfig()
}

func Create() *ServiceBrokerStub {
	return New().Create()
}

func Register() *ServiceBrokerStub {
	return New().Create().Register()
}

func EnableServiceAccess() *ServiceBrokerStub {
	return New().Create().Register().EnableServiceAccess()
}

func (s *ServiceBrokerStub) Create() *ServiceBrokerStub {
	ensureAppIsDeployed()
	s.requestNewBrokerRoute()
	return s
}

func (s *ServiceBrokerStub) Forget() {
	s.forget()
}

func (s *ServiceBrokerStub) Register() *ServiceBrokerStub {
	s.register()
	return s
}

func (s *ServiceBrokerStub) RegisterViaV2() *ServiceBrokerStub {
	s.registerViaV2()
	return s
}

func (s *ServiceBrokerStub) EnableServiceAccess() *ServiceBrokerStub {
	s.enableServiceAccess()
	return s
}

func (s *ServiceBrokerStub) FirstServiceOfferingName() string {
	return s.Services[0].Name
}

func (s *ServiceBrokerStub) FirstServicePlanName() string {
	return s.Services[0].Plans[0].Name
}
