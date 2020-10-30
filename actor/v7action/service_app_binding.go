package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/railway"
)

type CreateServiceAppBindingParams struct {
	SpaceGUID           string
	ServiceInstanceName string
	AppName             string
	BindingName         string
	Parameters          types.OptionalObject
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
			jobURL, warnings, err = actor.createServiceAppBinding(serviceInstance.GUID, app.GUID, params.BindingName, params.Parameters)
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

func (actor Actor) createServiceAppBinding(serviceInstanceGUID, appGUID, bindingName string, parameters types.OptionalObject) (ccv3.JobURL, ccv3.Warnings, error) {
	jobURL, warnings, err := actor.CloudControllerClient.CreateServiceCredentialBinding(resources.ServiceCredentialBinding{
		Type:                resources.AppBinding,
		Name:                bindingName,
		ServiceInstanceGUID: serviceInstanceGUID,
		AppGUID:             appGUID,
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
