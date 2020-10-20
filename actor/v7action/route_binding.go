package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/railway"
)

type CreateRouteBindingParams struct {
	SpaceGUID           string
	ServiceInstanceName string
	DomainName          string
	Hostname            string
	Path                string
	Parameters          types.OptionalObject
}

func (actor Actor) CreateRouteBinding(params CreateRouteBindingParams) (chan PollJobEvent, Warnings, error) {
	var (
		serviceInstance resources.ServiceInstance
		domain          resources.Domain
		route           resources.Route
		jobURL          ccv3.JobURL
		stream          chan PollJobEvent
	)

	warnings, err := railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			serviceInstance, warnings, err = actor.getServiceInstanceByNameAndSpace(params.ServiceInstanceName, params.SpaceGUID)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			domain, warnings, err = actor.getDomainByName(params.DomainName)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			route, warnings, err = actor.getRouteForBinding(params, domain)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			jobURL, warnings, err = actor.createRouteBinding(serviceInstance.GUID, route.GUID, params.Parameters)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			stream = actor.PollJobToEventStream(jobURL)
			return
		},
	)

	return stream, Warnings(warnings), err
}

func (actor Actor) createRouteBinding(serviceInstanceGUID, routeGUID string, parameters types.OptionalObject) (ccv3.JobURL, ccv3.Warnings, error) {
	jobURL, warnings, err := actor.CloudControllerClient.CreateRouteBinding(resources.RouteBinding{
		ServiceInstanceGUID: serviceInstanceGUID,
		RouteGUID:           routeGUID,
		Parameters:          parameters,
	})
	switch err.(type) {
	case nil:
		return jobURL, warnings, nil
	case ccerror.ResourceAlreadyExistsError:
		return "", warnings, actionerror.ResourceAlreadyExistsError{Message: err.Error()}
	default:
		return "", warnings, err
	}
}

func (actor Actor) getRouteForBinding(params CreateRouteBindingParams, domain resources.Domain) (resources.Route, ccv3.Warnings, error) {
	query := []ccv3.Query{{Key: ccv3.DomainGUIDFilter, Values: []string{domain.GUID}}}
	if params.Hostname != "" {
		query = append(query, ccv3.Query{Key: ccv3.HostsFilter, Values: []string{params.Hostname}})
	}
	if params.Path != "" {
		query = append(query, ccv3.Query{Key: ccv3.PathsFilter, Values: []string{params.Path}})
	}

	routes, warnings, err := actor.CloudControllerClient.GetRoutes(query...)
	switch {
	case err != nil:
		return resources.Route{}, warnings, err
	case len(routes) == 0:
		return resources.Route{}, warnings, actionerror.RouteNotFoundError{
			Host:       params.Hostname,
			DomainName: domain.Name,
			Path:       params.Path,
		}
	default:
		return routes[0], warnings, nil
	}
}
