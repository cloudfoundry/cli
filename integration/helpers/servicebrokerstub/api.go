package servicebrokerstub

import (
	"time"

	"code.cloudfoundry.org/cli/integration/assets/hydrabroker/config"
)

type ServiceBrokerStub struct {
	Name, URL, GUID     string
	Username, Password  string
	Services            []config.Service
	created             bool
	registered          bool
	CatalogResponse     int
	ProvisionResponse   int
	DeprovisionResponse int
	AsyncResponseDelay  time.Duration
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

func (s *ServiceBrokerStub) Configure() *ServiceBrokerStub {
	s.configure()
	return s
}

func (s *ServiceBrokerStub) Register() *ServiceBrokerStub {
	if !s.created {
		s.Create()
	}
	s.register(false)
	s.registered = true
	return s
}

func (s *ServiceBrokerStub) RegisterSpaceScoped() *ServiceBrokerStub {
	if !s.created {
		s.Create()
	}
	s.register(true)
	s.registered = true
	return s
}

func (s *ServiceBrokerStub) RegisterViaV2() *ServiceBrokerStub {
	if !s.created {
		s.Create()
	}
	s.registerViaV2()
	s.registered = true
	return s
}

func (s *ServiceBrokerStub) EnableServiceAccess() *ServiceBrokerStub {
	if !s.registered {
		s.Register()
	}
	s.enableServiceAccess()
	return s
}

func (s *ServiceBrokerStub) WithPlans(plans int) *ServiceBrokerStub {
	for len(s.Services[0].Plans) < plans {
		s.Services[0].Plans = append(s.Services[0].Plans, newDefaultPlan())
	}
	return s
}

func (s *ServiceBrokerStub) WithName(name string) *ServiceBrokerStub {
	s.Name = name
	return s
}

func (s *ServiceBrokerStub) WithCatalogResponse(statusCode int) *ServiceBrokerStub {
	s.CatalogResponse = statusCode
	return s
}

func (s *ServiceBrokerStub) WithProvisionResponse(statusCode int) *ServiceBrokerStub {
	s.ProvisionResponse = statusCode
	return s
}

func (s *ServiceBrokerStub) WithDeprovisionResponse(statusCode int) *ServiceBrokerStub {
	s.DeprovisionResponse = statusCode
	return s
}

func (s *ServiceBrokerStub) WithAsyncDelay(delay time.Duration) *ServiceBrokerStub {
	s.AsyncResponseDelay = delay
	return s
}

func (s *ServiceBrokerStub) FirstServiceOfferingName() string {
	return s.Services[0].Name
}

func (s *ServiceBrokerStub) FirstServiceOfferingDescription() string {
	return s.Services[0].Description
}

func (s *ServiceBrokerStub) FirstServicePlanName() string {
	return s.Services[0].Plans[0].Name
}

func (s *ServiceBrokerStub) FirstServicePlanDescription() string {
	return s.Services[0].Plans[0].Description
}
