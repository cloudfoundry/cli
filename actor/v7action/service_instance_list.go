package v7action

import (
	"fmt"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/batcher"
	"code.cloudfoundry.org/cli/util/railway"
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
	var (
		instances []resources.ServiceInstance
		bindings  []resources.ServiceCredentialBinding
		included  ccv3.IncludedResources
	)

	warnings, err := railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			instances, included, warnings, err = actor.CloudControllerClient.GetServiceInstances(
				ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{spaceGUID}},
				ccv3.Query{Key: ccv3.FieldsServicePlan, Values: []string{"guid", "name", "relationships.service_offering"}},
				ccv3.Query{Key: ccv3.FieldsServicePlanServiceOffering, Values: []string{"guid", "name", "relationships.service_broker"}},
				ccv3.Query{Key: ccv3.FieldsServicePlanServiceOfferingServiceBroker, Values: []string{"guid", "name"}},
				ccv3.Query{Key: ccv3.OrderBy, Values: []string{"name"}},
			)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			if !omitApps {
				return batcher.RequestByGUID(instanceGUIDS(instances), func(guids []string) (ccv3.Warnings, error) {
					batch, warnings, err := actor.CloudControllerClient.GetServiceCredentialBindings(
						ccv3.Query{Key: ccv3.ServiceInstanceGUIDFilter, Values: guids},
						ccv3.Query{Key: ccv3.Include, Values: []string{"app"}},
					)
					bindings = append(bindings, batch...)
					return warnings, err
				})
			}
			return
		},
	)
	if err != nil {
		return nil, Warnings(warnings), err
	}

	planDetailsFromPlanGUIDLookup := buildPlanDetailsLookup(included)
	boundAppsNamesFromInstanceGUIDLookup := buildBoundAppsLookup(bindings, spaceGUID)

	result := make([]ServiceInstance, len(instances))
	for i, instance := range instances {
		names := planDetailsFromPlanGUIDLookup[instance.ServicePlanGUID]
		result[i] = ServiceInstance{
			Name:                instance.Name,
			Type:                instance.Type,
			UpgradeAvailable:    instance.UpgradeAvailable,
			ServicePlanName:     names.plan,
			ServiceOfferingName: names.offering,
			ServiceBrokerName:   names.broker,
			BoundApps:           boundAppsNamesFromInstanceGUIDLookup[instance.GUID],
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

func buildBoundAppsLookup(bindings []resources.ServiceCredentialBinding, spaceGUID string) map[string][]string {
	appsBoundLookup := make(map[string][]string)
	for _, binding := range bindings {
		if binding.Type == resources.AppBinding && binding.AppSpaceGUID == spaceGUID {
			appsBoundLookup[binding.ServiceInstanceGUID] = append(appsBoundLookup[binding.ServiceInstanceGUID], binding.AppName)
		}
	}
	return appsBoundLookup
}
