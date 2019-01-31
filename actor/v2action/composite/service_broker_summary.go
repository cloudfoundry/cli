package composite

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
)

//go:generate counterfeiter . ServiceActor

type ServiceActor interface {
	GetServicesWithPlans(filters ...v2action.Filter) (v2action.ServicesWithPlans, v2action.Warnings, error)
	ServiceExistsWithName(serviceName string) (bool, v2action.Warnings, error)
}

//go:generate counterfeiter . BrokerActor

type BrokerActor interface {
	GetServiceBrokerByName(brokerName string) (v2action.ServiceBroker, v2action.Warnings, error)
	GetServiceBrokers() ([]v2action.ServiceBroker, v2action.Warnings, error)
}

//go:generate counterfeiter . OrganizationActor

type OrganizationActor interface {
	GetOrganization(organizationGUID string) (v2action.Organization, v2action.Warnings, error)
	OrganizationExistsWithName(organizationName string) (bool, v2action.Warnings, error)
}

//go:generate counterfeiter . VisibilityActor

type VisibilityActor interface {
	GetServicePlanVisibilities(planGUID string) ([]v2action.ServicePlanVisibility, v2action.Warnings, error)
}

type ServiceBrokerSummaryCompositeActor struct {
	ServiceActor    ServiceActor
	BrokerActor     BrokerActor
	OrgActor        OrganizationActor
	VisibilityActor VisibilityActor
}

// GetServiceBrokerSummaries returns summaries for service brokers that match the arguments passed.
// An error will be returned if any of the options are invalid (i.e. there is no broker, service or
// organization with the given names). Consider the structure of Service Brokers, Services,
// and Plans as a tree (a Broker may have many Services, a Service may have many Plans) for the purpose
// of this explanation. Each of the provided arguments will act as a filtering mechanism,
// with the expectation that the caller does not want matches for "parent" concepts, if they have no
// "children" matching a filtered argument.
//
// For example, given a Broker "Foo", only containing a Service "Bar":
//
// `GetServiceBrokerSummaries("Foo", "NotBar", "")` will return a slice of broker summaries that does not
// include the Broker "Foo".
//
// Similarly, given a broker "Foo", containing a Service "Bar", that has plans available only in Organization "Baz":
//
// `GetServiceBrokerSummaries("Foo", "Bar", "NotBaz") will recurse upwards resulting in a slice of broker
// summaries that does not include the Broker "Foo" either.
func (c *ServiceBrokerSummaryCompositeActor) GetServiceBrokerSummaries(brokerName string, serviceName string, organizationName string) ([]v2action.ServiceBrokerSummary, v2action.Warnings, error) {
	var warnings v2action.Warnings

	if organizationName != "" {
		orgExistsWarnings, err := c.ensureOrgExists(organizationName)
		warnings = append(warnings, orgExistsWarnings...)
		if err != nil {
			return nil, warnings, err
		}
	}

	if serviceName != "" {
		serviceExistsWarnings, err := c.ensureServiceExists(serviceName)
		warnings = append(warnings, serviceExistsWarnings...)
		if err != nil {
			return nil, warnings, err
		}
	}

	brokers, brokerWarnings, err := c.getServiceBrokers(brokerName)
	warnings = append(warnings, brokerWarnings...)
	if err != nil {
		return nil, warnings, err
	}

	brokerSummaries, brokerSummaryWarnings, err := c.fetchBrokerSummaries(brokers, serviceName, organizationName)
	warnings = append(warnings, brokerSummaryWarnings...)
	if err != nil {
		return nil, warnings, err
	}

	if organizationName != "" || serviceName != "" {
		brokerSummaries = pruneEmptyLeaves(brokerSummaries)
	}

	return brokerSummaries, warnings, nil
}

func (c *ServiceBrokerSummaryCompositeActor) ensureOrgExists(organizationName string) (v2action.Warnings, error) {
	organizationExists, warnings, err := c.OrgActor.OrganizationExistsWithName(organizationName)
	if err != nil {
		return warnings, err
	}

	if !organizationExists {
		return warnings, actionerror.OrganizationNotFoundError{Name: organizationName}
	}

	return warnings, nil
}

func (c *ServiceBrokerSummaryCompositeActor) ensureServiceExists(serviceName string) (v2action.Warnings, error) {
	serviceExists, warnings, err := c.ServiceActor.ServiceExistsWithName(serviceName)
	if err != nil {
		return warnings, err
	}

	if !serviceExists {
		return warnings, actionerror.ServiceNotFoundError{Name: serviceName}
	}

	return warnings, nil
}

func (c *ServiceBrokerSummaryCompositeActor) getServiceBrokers(brokerName string) ([]v2action.ServiceBroker, v2action.Warnings, error) {
	if brokerName != "" {
		broker, brokerWarnings, err := c.BrokerActor.GetServiceBrokerByName(brokerName)
		return []v2action.ServiceBroker{broker}, brokerWarnings, err
	}
	return c.BrokerActor.GetServiceBrokers()
}

func (c *ServiceBrokerSummaryCompositeActor) fetchBrokerSummaries(brokers []v2action.ServiceBroker, serviceName, organizationName string) ([]v2action.ServiceBrokerSummary, v2action.Warnings, error) {
	var (
		brokerSummaries []v2action.ServiceBrokerSummary
		warnings        v2action.Warnings
	)

	for _, broker := range brokers {
		brokerSummary, brokerWarnings, err := c.fetchBrokerSummary(v2action.ServiceBroker(broker), serviceName, organizationName)
		warnings = append(warnings, brokerWarnings...)
		if err != nil {
			return nil, warnings, err
		}
		brokerSummaries = append(brokerSummaries, brokerSummary)
	}

	return brokerSummaries, warnings, nil
}

func (c *ServiceBrokerSummaryCompositeActor) fetchBrokerSummary(broker v2action.ServiceBroker, serviceName, organizationName string) (v2action.ServiceBrokerSummary, v2action.Warnings, error) {
	filters := []v2action.Filter{filterByBrokerGUID(broker.GUID)}
	if serviceName != "" {
		filters = append(filters, filterByServiceName(serviceName))
	}

	servicesWithPlans, warnings, err := c.ServiceActor.GetServicesWithPlans(filters...)
	if err != nil {
		return v2action.ServiceBrokerSummary{}, warnings, err
	}

	var services []v2action.ServiceSummary
	for service, servicePlans := range servicesWithPlans {
		serviceSummary, serviceWarnings, err := c.fetchServiceSummary(service, servicePlans, organizationName)
		warnings = append(warnings, serviceWarnings...)
		if err != nil {
			return v2action.ServiceBrokerSummary{}, warnings, err
		}
		services = append(services, serviceSummary)
	}

	return v2action.ServiceBrokerSummary{
		ServiceBroker: v2action.ServiceBroker(broker),
		Services:      services,
	}, warnings, nil
}

func (c *ServiceBrokerSummaryCompositeActor) fetchServiceSummary(service v2action.Service, servicePlans []v2action.ServicePlan, organizationName string) (v2action.ServiceSummary, v2action.Warnings, error) {
	var warnings v2action.Warnings
	var servicePlanSummaries []v2action.ServicePlanSummary
	for _, plan := range servicePlans {
		var visibleTo []string
		if !plan.Public {
			serviceVisibilities, visibilityWarnings, visErr := c.VisibilityActor.GetServicePlanVisibilities(plan.GUID)
			warnings = append(warnings, visibilityWarnings...)
			if visErr != nil {
				return v2action.ServiceSummary{}, v2action.Warnings(warnings), visErr
			}

			for _, serviceVisibility := range serviceVisibilities {
				org, orgWarnings, orgsErr := c.OrgActor.GetOrganization(serviceVisibility.OrganizationGUID)
				warnings = append(warnings, orgWarnings...)
				if orgsErr != nil {
					return v2action.ServiceSummary{}, v2action.Warnings(warnings), orgsErr
				}

				visibleTo = append(visibleTo, org.Name)
			}
		}

		if isPlanVisibleToOrg(organizationName, visibleTo, plan.Public) {
			servicePlanSummaries = append(servicePlanSummaries,
				v2action.ServicePlanSummary{
					ServicePlan: v2action.ServicePlan(plan),
					VisibleTo:   visibleTo,
				})
		}
	}

	return v2action.ServiceSummary{
		Service: v2action.Service(service),
		Plans:   servicePlanSummaries,
	}, v2action.Warnings(warnings), nil
}

func isPlanVisibleToOrg(organizationName string, visibleTo []string, isPlanPublic bool) bool {
	if organizationName == "" || isPlanPublic {
		return true
	}

	for _, visibleOrgName := range visibleTo {
		if visibleOrgName == organizationName {
			return true
		}
	}
	return false
}

func pruneEmptyLeaves(brokerSummaries []v2action.ServiceBrokerSummary) []v2action.ServiceBrokerSummary {
	filteredBrokerSummaries := []v2action.ServiceBrokerSummary{}

	for _, brokerSummary := range brokerSummaries {
		filteredServiceSummaries := []v2action.ServiceSummary{}
		for _, serviceSummary := range brokerSummary.Services {
			if len(serviceSummary.Plans) != 0 {
				filteredServiceSummaries = append(filteredServiceSummaries, serviceSummary)
			}
		}

		if len(filteredServiceSummaries) != 0 {
			filteredBrokerSummaries = append(filteredBrokerSummaries, brokerSummary)
		}
	}

	return filteredBrokerSummaries
}

func filterByServiceName(serviceName string) v2action.Filter {
	return v2action.Filter{
		Type:     constant.LabelFilter,
		Operator: constant.EqualOperator,
		Values:   []string{serviceName},
	}
}
func filterByBrokerGUID(brokerGUID string) v2action.Filter {
	return v2action.Filter{
		Type:     constant.ServiceBrokerGUIDFilter,
		Operator: constant.EqualOperator,
		Values:   []string{brokerGUID},
	}
}
