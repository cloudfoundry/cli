package v7action

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/railway"
)

type RouteBindingSummary struct {
	URL           string
	LastOperation resources.LastOperation
}

type BindingList struct {
	App   []resources.ServiceCredentialBinding
	Key   []resources.ServiceCredentialBinding
	Route []RouteBindingSummary
}

type BindingListParameters struct {
	ServiceInstanceName string
	SpaceGUID           string
	GetAppBindings      bool
	GetServiceKeys      bool
	GetRouteBindings    bool
}

func (actor Actor) GetBindingsByServiceInstance(params BindingListParameters) (BindingList, Warnings, error) {
	var (
		routeBindings      []resources.RouteBinding
		credentialBindings []resources.ServiceCredentialBinding
		serviceInstance    resources.ServiceInstance
		includedResources  ccv3.IncludedResources
		result             BindingList
	)

	warnings, err := railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			serviceInstance, _, warnings, err = actor.getServiceInstanceByNameAndSpace(params.ServiceInstanceName, params.SpaceGUID)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			if params.GetRouteBindings {
				routeBindings, includedResources, warnings, err = actor.CloudControllerClient.GetRouteBindings(
					ccv3.Query{Key: ccv3.ServiceInstanceGUIDFilter, Values: []string{serviceInstance.GUID}},
					ccv3.Query{Key: ccv3.Include, Values: []string{"route"}},
				)
			}
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			query := []ccv3.Query{
				{Key: ccv3.ServiceInstanceGUIDFilter, Values: []string{serviceInstance.GUID}},
				{Key: ccv3.Include, Values: []string{"app"}},
			}

			switch {
			case params.GetServiceKeys && params.GetAppBindings:
			case params.GetServiceKeys:
				query = append(query, ccv3.Query{Key: ccv3.TypeFilter, Values: []string{"key"}})
			case params.GetAppBindings:
				query = append(query, ccv3.Query{Key: ccv3.TypeFilter, Values: []string{"app"}})
			}

			if params.GetAppBindings || params.GetServiceKeys {
				credentialBindings, warnings, err = actor.CloudControllerClient.GetServiceCredentialBindings(query...)
			}
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			result = partitionBindingData(routeBindings, includedResources.Routes, credentialBindings)
			return
		},
	)
	return result, Warnings(warnings), err
}

func partitionBindingData(routeBindings []resources.RouteBinding, routes []resources.Route, credentialBindings []resources.ServiceCredentialBinding) BindingList {
	routeGUIDLookup := make(map[string]string)
	for _, r := range routes {
		routeGUIDLookup[r.GUID] = r.URL
	}

	result := BindingList{}
	for _, rb := range routeBindings {
		result.Route = append(result.Route, RouteBindingSummary{
			URL:           routeGUIDLookup[rb.RouteGUID],
			LastOperation: rb.LastOperation,
		})
	}

	for _, b := range credentialBindings {
		switch b.Type {
		case resources.KeyBinding:
			result.Key = append(result.Key, b)
		case resources.AppBinding:
			result.App = append(result.App, b)
		}
	}

	return result
}
