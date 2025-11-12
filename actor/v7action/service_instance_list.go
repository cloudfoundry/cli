package v7action

import (
	"fmt"

	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v9/resources"
	"code.cloudfoundry.org/cli/v9/types"
	"code.cloudfoundry.org/cli/v9/util/batcher"
	"code.cloudfoundry.org/cli/v9/util/extract"
	"code.cloudfoundry.org/cli/v9/util/lookuptable"
	"code.cloudfoundry.org/cli/v9/util/railway"
)

type ServiceInstance struct {
	Type                resources.ServiceInstanceType
	Name                string
	ServicePlanName     string
	ServiceOfferingName string
	ServiceBrokerName   string
	BoundApps           []string
	LastOperation       string
	UpgradeAvailable    types.OptionalBoolean
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
				ccv3.Query{Key: ccv3.OrderBy, Values: []string{ccv3.NameOrder}},
				ccv3.Query{Key: ccv3.PerPage, Values: []string{ccv3.MaxPerPage}},
			)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			if !omitApps {
				return batcher.RequestByGUID(
					extract.UniqueList("GUID", instances),
					func(guids []string) (ccv3.Warnings, error) {
						batch, warnings, err := actor.CloudControllerClient.GetServiceCredentialBindings(
							ccv3.Query{Key: ccv3.ServiceInstanceGUIDFilter, Values: guids},
							ccv3.Query{Key: ccv3.Include, Values: []string{"app"}},
						)
						bindings = append(bindings, batch...)
						return warnings, err
					},
				)
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

func buildPlanDetailsLookup(included ccv3.IncludedResources) map[string]planDetails {
	brokerLookup := lookuptable.NameFromGUID(included.ServiceBrokers)

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
