package v7action

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	"fmt"
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

type planDetails struct {
	plan, offering, broker string
}

func (actor Actor) GetServiceInstancesForSpace(spaceGUID string, omitApps bool) ([]ServiceInstance, Warnings, error) {
	instances, included, warnings, err := actor.CloudControllerClient.GetServiceInstances(
		ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{spaceGUID}},
		ccv3.Query{Key: ccv3.FieldsServicePlan, Values: []string{"guid", "name", "relationships.service_offering"}},
		ccv3.Query{Key: ccv3.FieldsServicePlanServiceOffering, Values: []string{"guid", "name", "relationships.service_broker"}},
		ccv3.Query{Key: ccv3.FieldsServicePlanServiceOfferingServiceBroker, Values: []string{"guid", "name"}},
		ccv3.Query{Key: ccv3.OrderBy, Values: []string{"name"}},
	)
	if err != nil {
		return nil, Warnings(warnings), err
	}

	planDetailsLookup := buildPlanDetailsLookup(included)
	var boundAppsLookup map[string][]string
	if !omitApps {
		var bindingsWarnings ccv3.Warnings
		bindings, included, bindingsWarnings, err := actor.CloudControllerClient.GetServiceCredentialBindings(
			ccv3.Query{Key: ccv3.ServiceInstanceGUIDFilter, Values: instanceGUIDS(instances)},
			ccv3.Query{Key: ccv3.Include, Values: []string{"app"}},
		)
		warnings = append(warnings, bindingsWarnings...)
		if err != nil {
			return nil, Warnings(warnings), err
		}
		boundAppsLookup = buildBoundAppsLookup(bindings, included)
	}

	result := make([]ServiceInstance, len(instances))
	for i, instance := range instances {
		names := planDetailsLookup[instance.ServicePlanGUID]
		result[i] = ServiceInstance{
			Name:                instance.Name,
			Type:                instance.Type,
			UpgradeAvailable:    instance.UpgradeAvailable,
			ServicePlanName:     names.plan,
			ServiceOfferingName: names.offering,
			ServiceBrokerName:   names.broker,
			BoundApps:           boundAppsLookup[instance.GUID],
			LastOperation:       lastOperation(instance.LastOperation),
		}
	}

	return result, Warnings(warnings), nil
}

func lastOperation(lo resources.LastOperation) string {
	if lo.Type != "" && lo.State != "" {
		return fmt.Sprintf("%s %s", lo.Type, lo.State)
	}
	return ""
}

func instanceGUIDS(instances []resources.ServiceInstance) []string {
	serviceInstanceGUIDS := make([]string, len(instances))
	for i, instance := range instances {
		serviceInstanceGUIDS[i] = instance.GUID
	}
	return serviceInstanceGUIDS
}

func buildPlanDetailsLookup(included ccv3.IncludedResources) map[string]planDetails {
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

	planLookup := make(map[string]planDetails)
	for _, p := range included.ServicePlans {
		names := offeringLookup[p.ServiceOfferingGUID]
		planLookup[p.GUID] = planDetails{
			broker:   names.broker,
			offering: names.offering,
			plan:     p.Name,
		}
	}
	return planLookup
}

func buildBoundAppsLookup(bindings []resources.ServiceCredentialBinding, included ccv3.IncludedResources) map[string][]string {
	appLookup := make(map[string]string, len(included.Apps))
	for _, app := range included.Apps {
		appLookup[app.GUID] = app.Name
	}
	appsBoundLookup := make(map[string][]string)
	for _, binding := range bindings {
		if binding.Type == resources.AppBinding {
			appsBoundLookup[binding.ServiceInstanceGUID] = append(appsBoundLookup[binding.ServiceInstanceGUID], appLookup[binding.AppGUID])
		}
	}
	return appsBoundLookup
}
