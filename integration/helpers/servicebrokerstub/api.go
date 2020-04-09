package servicebrokerstub

import (
	"code.cloudfoundry.org/cli/integration/assets/hydrabroker/config"
	"code.cloudfoundry.org/cli/integration/helpers"
)

type ServiceBrokerStub struct {
	Name, URL, GUID    string
	Username, Password string
	Services           []config.Service
	registered         bool
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
	if s.registered {
		s.deregister()
	}
	s.forget()
}

func (s *ServiceBrokerStub) Register() *ServiceBrokerStub {
	s.register()
	s.registered = true
	return s
}

func (s *ServiceBrokerStub) RegisterViaV2() *ServiceBrokerStub {
	s.registerViaV2()
	s.registered = true
	return s
}

func (s *ServiceBrokerStub) EnableServiceAccess() *ServiceBrokerStub {
	s.enableServiceAccess()
	return s
}

func (s *ServiceBrokerStub) WithPlans(plans int) *ServiceBrokerStub {
	for len(s.Services[0].Plans) < plans {
		s.Services[0].Plans = append(s.Services[0].Plans, config.Plan{Name: helpers.PrefixedRandomName("INTEGRATION-PLAN")})
	}
	return s
}

func (s *ServiceBrokerStub) FirstServiceOfferingName() string {
	return s.Services[0].Name
}

func (s *ServiceBrokerStub) FirstServicePlanName() string {
	return s.Services[0].Plans[0].Name
}
