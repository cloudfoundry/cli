package servicebrokerstub

import (
	"time"

	"code.cloudfoundry.org/cli/integration/helpers"

	"code.cloudfoundry.org/cli/integration/assets/hydrabroker/config"
)

type ServiceBrokerStub struct {
	Name, URL, GUID     string
	Username, Password  string
	Services            []config.Service
	created             bool
	registered          bool
	catalogResponse     int
	provisionResponse   int
	deprovisionResponse int
	asyncResponseDelay  time.Duration
}

func New() *ServiceBrokerStub {
	return newDefaultConfig()
}

func Create() *ServiceBrokerStub {
	return New().Create()
}

func Register() *ServiceBrokerStub {
	return New().Register()
}

func EnableServiceAccess() *ServiceBrokerStub {
	return New().EnableServiceAccess()
}

func (s *ServiceBrokerStub) Create() *ServiceBrokerStub {
	ensureAppIsDeployed()
	s.requestNewBrokerRoute()
	s.created = true
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

func (s *ServiceBrokerStub) WithName(name string) *ServiceBrokerStub {
	s.Name = name
	return s
}

func (s *ServiceBrokerStub) WithServiceOfferings(services int) *ServiceBrokerStub {
	previousName := s.Services[0].Name
	for len(s.Services) < services {
		ser := newDefaultServiceOffering()
		ser.Name = helpers.GenerateHigherName(helpers.NewServiceOfferingName, previousName)
		previousName = ser.Name
		s.Services = append(s.Services, ser)
	}
	return s
}

func (s *ServiceBrokerStub) WithPlans(plans int) *ServiceBrokerStub {
	previousName := s.Services[0].Plans[0].Name
	for len(s.Services[0].Plans) < plans {
		p := newDefaultPlan()
		p.Name = helpers.GenerateHigherName(helpers.NewPlanName, previousName)
		previousName = p.Name
		s.Services[0].Plans = append(s.Services[0].Plans, p)
	}
	return s
}

func (s *ServiceBrokerStub) WithHigherNameThan(o *ServiceBrokerStub) *ServiceBrokerStub {
	higherName := helpers.GenerateHigherName(helpers.NewServiceBrokerName, o.Name)
	return s.WithName(higherName)
}

func (s *ServiceBrokerStub) WithCatalogResponse(statusCode int) *ServiceBrokerStub {
	s.catalogResponse = statusCode
	return s
}

func (s *ServiceBrokerStub) WithProvisionResponse(statusCode int) *ServiceBrokerStub {
	s.provisionResponse = statusCode
	return s
}

func (s *ServiceBrokerStub) WithDeprovisionResponse(statusCode int) *ServiceBrokerStub {
	s.deprovisionResponse = statusCode
	return s
}

func (s *ServiceBrokerStub) WithAsyncDelay(delay time.Duration) *ServiceBrokerStub {
	s.asyncResponseDelay = delay
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
