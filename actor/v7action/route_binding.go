package v7action

import (
	"code.cloudfoundry.org/cli/v9/actor/actionerror"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v9/resources"
	"code.cloudfoundry.org/cli/v9/types"
	"code.cloudfoundry.org/cli/v9/util/railway"
)

type CreateRouteBindingParams struct {
	SpaceGUID           string
	ServiceInstanceName string
	DomainName          string
	Hostname            string
	Path                string
	Parameters          types.OptionalObject
}

type DeleteRouteBindingParams struct {
	SpaceGUID           string
	ServiceInstanceName string
	DomainName          string
	Hostname            string
	Path                string
}

type getRouteForBindingParams struct {
	SpaceGUID  string
	DomainName string
	Hostname   string
	Path       string
}

func (actor Actor) CreateRouteBinding(params CreateRouteBindingParams) (chan PollJobEvent, Warnings, error) {
	var (
		serviceInstance resources.ServiceInstance
		route           resources.Route
		jobURL          ccv3.JobURL
		stream          chan PollJobEvent
	)

	warnings, err := railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			serviceInstance, _, warnings, err = actor.getServiceInstanceByNameAndSpace(params.ServiceInstanceName, params.SpaceGUID)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			route, warnings, err = actor.getRouteForBinding(getRouteForBindingParams{
				SpaceGUID:  params.SpaceGUID,
				DomainName: params.DomainName,
				Hostname:   params.Hostname,
				Path:       params.Path,
			})
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

func (actor Actor) DeleteRouteBinding(params DeleteRouteBindingParams) (chan PollJobEvent, Warnings, error) {
	var (
		serviceInstance resources.ServiceInstance
		route           resources.Route
		binding         resources.RouteBinding
		jobURL          ccv3.JobURL
		stream          chan PollJobEvent
	)

	warnings, err := railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			serviceInstance, _, warnings, err = actor.getServiceInstanceByNameAndSpace(params.ServiceInstanceName, params.SpaceGUID)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			route, warnings, err = actor.getRouteForBinding(getRouteForBindingParams{
				SpaceGUID:  params.SpaceGUID,
				DomainName: params.DomainName,
				Hostname:   params.Hostname,
				Path:       params.Path,
			})
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			binding, warnings, err = actor.getRouteBinding(serviceInstance.GUID, route.GUID)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			jobURL, warnings, err = actor.CloudControllerClient.DeleteRouteBinding(binding.GUID)
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

func (actor Actor) getRouteBinding(serviceInstanceGUID, routeGUID string) (resources.RouteBinding, ccv3.Warnings, error) {
	bindings, _, warnings, err := actor.CloudControllerClient.GetRouteBindings(
		ccv3.Query{Key: ccv3.RouteGUIDFilter, Values: []string{routeGUID}},
		ccv3.Query{Key: ccv3.ServiceInstanceGUIDFilter, Values: []string{serviceInstanceGUID}},
		ccv3.Query{Key: ccv3.PerPage, Values: []string{"1"}},
		ccv3.Query{Key: ccv3.Page, Values: []string{"1"}},
	)

	switch {
	case err != nil:
		return resources.RouteBinding{}, warnings, err
	case len(bindings) == 0:
		return resources.RouteBinding{}, warnings, actionerror.RouteBindingNotFoundError{}
	default:
		return bindings[0], warnings, nil
	}
}

func (actor Actor) getRouteForBinding(params getRouteForBindingParams) (resources.Route, ccv3.Warnings, error) {
	var (
		domain resources.Domain
		routes []resources.Route
	)

	warnings, err := railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			domain, warnings, err = actor.getDomainByName(params.DomainName)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			routes, warnings, err = actor.CloudControllerClient.GetRoutes(
				ccv3.Query{Key: ccv3.DomainGUIDFilter, Values: []string{domain.GUID}},
				ccv3.Query{Key: ccv3.HostsFilter, Values: []string{params.Hostname}},
				ccv3.Query{Key: ccv3.PathsFilter, Values: []string{params.Path}},
				ccv3.Query{Key: ccv3.PerPage, Values: []string{"1"}},
				ccv3.Query{Key: ccv3.Page, Values: []string{"1"}},
			)
			return
		},
	)

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
