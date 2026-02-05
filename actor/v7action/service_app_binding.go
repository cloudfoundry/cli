package v7action

import (
	"code.cloudfoundry.org/cli/v9/actor/actionerror"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v9/resources"
	"code.cloudfoundry.org/cli/v9/types"
	"code.cloudfoundry.org/cli/v9/util/railway"
)

type CreateServiceAppBindingParams struct {
	SpaceGUID           string
	ServiceInstanceName string
	AppName             string
	BindingName         string
	Parameters          types.OptionalObject
	Strategy            resources.BindingStrategyType
}

type ListServiceAppBindingParams struct {
	SpaceGUID           string
	ServiceInstanceName string
	AppName             string
}

type DeleteServiceAppBindingParams struct {
	ServiceBindingGUID string
}

func (actor Actor) CreateServiceAppBinding(params CreateServiceAppBindingParams) (chan PollJobEvent, Warnings, error) {
	var (
		serviceInstance resources.ServiceInstance
		app             resources.Application
		jobURL          ccv3.JobURL
		stream          chan PollJobEvent
	)

	warnings, err := railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			serviceInstance, _, warnings, err = actor.getServiceInstanceByNameAndSpace(params.ServiceInstanceName, params.SpaceGUID)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			app, warnings, err = actor.CloudControllerClient.GetApplicationByNameAndSpace(params.AppName, params.SpaceGUID)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			jobURL, warnings, err = actor.createServiceAppBinding(serviceInstance.GUID, app.GUID, params.BindingName, params.Parameters, params.Strategy)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			stream = actor.PollJobToEventStream(jobURL)
			return
		},
	)

	switch err.(type) {
	case nil:
		return stream, Warnings(warnings), nil
	case ccerror.ApplicationNotFoundError:
		return nil, Warnings(warnings), actionerror.ApplicationNotFoundError{Name: params.AppName}
	default:
		return nil, Warnings(warnings), err
	}
}

func (actor Actor) ListServiceAppBindings(params ListServiceAppBindingParams) ([]resources.ServiceCredentialBinding, Warnings, error) {
	var (
		serviceInstance resources.ServiceInstance
		app             resources.Application
		bindings        []resources.ServiceCredentialBinding
	)

	warnings, err := railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			serviceInstance, _, warnings, err = actor.getServiceInstanceByNameAndSpace(params.ServiceInstanceName, params.SpaceGUID)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			app, warnings, err = actor.CloudControllerClient.GetApplicationByNameAndSpace(params.AppName, params.SpaceGUID)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			bindings, warnings, err = actor.getServiceAppBindings(serviceInstance.GUID, app.GUID)
			return
		},
	)

	switch err.(type) {
	case nil:
		return bindings, Warnings(warnings), nil
	case ccerror.ApplicationNotFoundError:
		return nil, Warnings(warnings), actionerror.ApplicationNotFoundError{Name: params.AppName}
	default:
		return nil, Warnings(warnings), err
	}
}

func (actor Actor) DeleteServiceAppBinding(params DeleteServiceAppBindingParams) (chan PollJobEvent, Warnings, error) {
	var (
		jobURL ccv3.JobURL
		stream chan PollJobEvent
	)

	warnings, err := railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			jobURL, warnings, err = actor.CloudControllerClient.DeleteServiceCredentialBinding(params.ServiceBindingGUID)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			stream = actor.PollJobToEventStream(jobURL)
			return
		},
	)

	switch err.(type) {
	case nil:

		return stream, Warnings(warnings), nil
	default:

		return nil, Warnings(warnings), err
	}
}

func (actor Actor) createServiceAppBinding(serviceInstanceGUID, appGUID, bindingName string, parameters types.OptionalObject, strategy resources.BindingStrategyType) (ccv3.JobURL, ccv3.Warnings, error) {
	jobURL, warnings, err := actor.CloudControllerClient.CreateServiceCredentialBinding(resources.ServiceCredentialBinding{
		Type:                resources.AppBinding,
		Name:                bindingName,
		ServiceInstanceGUID: serviceInstanceGUID,
		AppGUID:             appGUID,
		Parameters:          parameters,
		Strategy:            strategy,
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

func (actor Actor) getServiceAppBindings(serviceInstanceGUID, appGUID string) ([]resources.ServiceCredentialBinding, ccv3.Warnings, error) {
	bindings, warnings, err := actor.CloudControllerClient.GetServiceCredentialBindings(
		ccv3.Query{Key: ccv3.TypeFilter, Values: []string{"app"}},
		ccv3.Query{Key: ccv3.ServiceInstanceGUIDFilter, Values: []string{serviceInstanceGUID}},
		ccv3.Query{Key: ccv3.AppGUIDFilter, Values: []string{appGUID}},
	)

	switch {
	case err != nil:
		return []resources.ServiceCredentialBinding{}, warnings, err
	case len(bindings) == 0:
		return []resources.ServiceCredentialBinding{}, warnings, actionerror.ServiceBindingNotFoundError{
			AppGUID:             appGUID,
			ServiceInstanceGUID: serviceInstanceGUID,
		}
	default:
		return bindings, warnings, nil
	}
}
