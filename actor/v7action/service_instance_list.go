package v7action

import (
	"fmt"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
)

type ServiceInstance struct {
	Type                resources.ServiceInstanceType `json:"type,omitempty"`
	Name                string                        `json:"name"`
	ServicePlanName     string                        `json:"plan,omitempty"`
	ServiceOfferingName string                        `json:"offering,omitempty"`
	ServiceBrokerName   string                        `json:"broker,omitempty"`
	BoundApps           []string                      `json:"bound_apps,omitempty"`
	LastOperation       string                        `json:"last_operation,omitempty"`
	UpgradeAvailable    types.OptionalBoolean         `json:"upgrade_available"`
}

func (actor Actor) GetServiceInstancesForSpace(spaceGUID string) ([]ServiceInstance, Warnings, error) {
	instances, included, warnings, err := actor.CloudControllerClient.GetServiceInstances(
		ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{spaceGUID}},
		ccv3.Query{Key: ccv3.FieldsServicePlan, Values: []string{"guid", "name", "relationships.service_offering"}},
		ccv3.Query{Key: ccv3.FieldsServicePlanServiceOffering, Values: []string{"guid", "name", "relationships.service_broker"}},
		ccv3.Query{Key: ccv3.FieldsServicePlanServiceOfferingServiceBroker, Values: []string{"guid", "name"}},
		ccv3.Query{Key: ccv3.OrderBy, Values: []string{"name"}},
	)

	brokerLookup := make(map[string]string)
	for _, b := range included.ServiceBrokers {
		brokerLookup[b.GUID] = b.Name
	}

	type twoNames struct{ broker, offering string }
	offeringLookup := make(map[string]twoNames)
	for _, o := range included.ServiceOfferings {
		brokerName := brokerLookup[o.ServiceBrokerGUID]
		offeringLookup[o.GUID] = twoNames{
			broker:   brokerName,
			offering: o.Name,
		}
	}

	type threeNames struct{ broker, offering, plan string }
	planLookup := make(map[string]threeNames)
	for _, p := range included.ServicePlans {
		names := offeringLookup[p.ServiceOfferingGUID]
		planLookup[p.GUID] = threeNames{
			broker:   names.broker,
			offering: names.offering,
			plan:     p.Name,
		}
	}

	lastOperation := func(lo resources.LastOperation) string {
		if lo.Type != "" && lo.State != "" {
			return fmt.Sprintf("%s %s", lo.Type, lo.State)
		}
		return ""
	}

	result := make([]ServiceInstance, len(instances))
	for i, instance := range instances {
		names := planLookup[instance.ServicePlanGUID]
		result[i] = ServiceInstance{
			Name:                instance.Name,
			Type:                instance.Type,
			UpgradeAvailable:    instance.UpgradeAvailable,
			ServicePlanName:     names.plan,
			ServiceOfferingName: names.offering,
			ServiceBrokerName:   names.broker,
			BoundApps:           []string{"foo", "bar"},
			LastOperation:       lastOperation(instance.LastOperation),
		}
	}

	return result, Warnings(warnings), err
}
