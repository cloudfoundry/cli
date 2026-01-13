package v7action

import (
	"fmt"

	"code.cloudfoundry.org/cli/v9/actor/actionerror"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v9/resources"
	"code.cloudfoundry.org/cli/v9/types"
	"code.cloudfoundry.org/cli/v9/util/railway"
)

type CreateServiceKeyParams struct {
	SpaceGUID           string
	ServiceInstanceName string
	ServiceKeyName      string
	Parameters          types.OptionalObject
}

func (actor Actor) CreateServiceKey(params CreateServiceKeyParams) (chan PollJobEvent, Warnings, error) {
	var (
		serviceInstance resources.ServiceInstance
		jobURL          ccv3.JobURL
		stream          chan PollJobEvent
	)

	warnings, err := railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			serviceInstance, _, warnings, err = actor.getServiceInstanceByNameAndSpace(params.ServiceInstanceName, params.SpaceGUID)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			jobURL, warnings, err = actor.createServiceKey(serviceInstance.GUID, params.ServiceKeyName, params.Parameters)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			stream = actor.PollJobToEventStream(jobURL)
			return
		},
	)

	if err != nil {
		return nil, Warnings(warnings), err
	}

	return stream, Warnings(warnings), nil
}

func (actor Actor) GetServiceKeysByServiceInstance(serviceInstanceName, spaceGUID string) ([]resources.ServiceCredentialBinding, Warnings, error) {
	var (
		serviceInstance resources.ServiceInstance
		keys            []resources.ServiceCredentialBinding
	)

	warnings, err := railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			serviceInstance, _, warnings, err = actor.getServiceInstanceByNameAndSpace(serviceInstanceName, spaceGUID)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			keys, warnings, err = actor.CloudControllerClient.GetServiceCredentialBindings(
				ccv3.Query{Key: ccv3.ServiceInstanceGUIDFilter, Values: []string{serviceInstance.GUID}},
				ccv3.Query{Key: ccv3.TypeFilter, Values: []string{"key"}},
				ccv3.Query{Key: ccv3.PerPage, Values: []string{ccv3.MaxPerPage}},
			)
			return
		},
	)

	return keys, Warnings(warnings), err
}

func (actor Actor) GetServiceKeyByServiceInstanceAndName(serviceInstanceName, serviceKeyName, spaceGUID string) (resources.ServiceCredentialBinding, Warnings, error) {
	key, warnings, err := actor.getServiceKeyByServiceInstanceAndName(serviceInstanceName, serviceKeyName, spaceGUID)
	return key, Warnings(warnings), err
}

func (actor Actor) GetServiceKeyDetailsByServiceInstanceAndName(serviceInstanceName, serviceKeyName, spaceGUID string) (resources.ServiceCredentialBindingDetails, Warnings, error) {
	var (
		key     resources.ServiceCredentialBinding
		details resources.ServiceCredentialBindingDetails
	)

	warnings, err := railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			key, warnings, err = actor.getServiceKeyByServiceInstanceAndName(serviceInstanceName, serviceKeyName, spaceGUID)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			details, warnings, err = actor.CloudControllerClient.GetServiceCredentialBindingDetails(key.GUID)
			return
		},
	)

	return details, Warnings(warnings), err
}

func (actor Actor) DeleteServiceKeyByServiceInstanceAndName(serviceInstanceName, serviceKeyName, spaceGUID string) (chan PollJobEvent, Warnings, error) {
	var (
		key    resources.ServiceCredentialBinding
		jobURL ccv3.JobURL
		stream chan PollJobEvent
	)

	warnings, err := railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			key, warnings, err = actor.getServiceKeyByServiceInstanceAndName(serviceInstanceName, serviceKeyName, spaceGUID)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			jobURL, warnings, err = actor.CloudControllerClient.DeleteServiceCredentialBinding(key.GUID)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			stream = actor.PollJobToEventStream(jobURL)
			return
		},
	)

	return stream, Warnings(warnings), err
}

func (actor Actor) createServiceKey(serviceInstanceGUID, serviceKeyName string, parameters types.OptionalObject) (ccv3.JobURL, ccv3.Warnings, error) {
	jobURL, warnings, err := actor.CloudControllerClient.CreateServiceCredentialBinding(resources.ServiceCredentialBinding{
		Type:                resources.KeyBinding,
		Name:                serviceKeyName,
		ServiceInstanceGUID: serviceInstanceGUID,
		Parameters:          parameters,
	})
	switch err.(type) {
	case nil:
		return jobURL, warnings, nil
	case ccerror.ServiceKeyTakenError:
		return "", warnings, actionerror.ResourceAlreadyExistsError{
			Message: fmt.Sprintf("Service key %s already exists", serviceKeyName),
		}
	default:
		return "", warnings, err
	}
}

func (actor Actor) getServiceKeyByServiceInstanceAndName(serviceInstanceName, serviceKeyName, spaceGUID string) (resources.ServiceCredentialBinding, ccv3.Warnings, error) {
	var (
		serviceInstance resources.ServiceInstance
		keys            []resources.ServiceCredentialBinding
	)

	warnings, err := railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			serviceInstance, _, warnings, err = actor.getServiceInstanceByNameAndSpace(serviceInstanceName, spaceGUID)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			keys, warnings, err = actor.CloudControllerClient.GetServiceCredentialBindings(
				ccv3.Query{Key: ccv3.ServiceInstanceGUIDFilter, Values: []string{serviceInstance.GUID}},
				ccv3.Query{Key: ccv3.TypeFilter, Values: []string{"key"}},
				ccv3.Query{Key: ccv3.NameFilter, Values: []string{serviceKeyName}},
				ccv3.Query{Key: ccv3.PerPage, Values: []string{"1"}},
				ccv3.Query{Key: ccv3.Page, Values: []string{"1"}},
			)
			return
		},
	)
	switch {
	case err != nil:
		return resources.ServiceCredentialBinding{}, warnings, err
	case len(keys) == 0:
		return resources.ServiceCredentialBinding{}, warnings, actionerror.NewServiceKeyNotFoundError(serviceKeyName, serviceInstanceName)
	default:
		return keys[0], warnings, nil
	}
}
