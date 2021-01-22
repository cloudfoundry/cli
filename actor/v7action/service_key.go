package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/extract"
	"code.cloudfoundry.org/cli/util/railway"
)

func (actor Actor) GetServiceKeysByServiceInstance(serviceInstanceName, spaceGUID string) ([]string, Warnings, error) {
	var (
		serviceInstance resources.ServiceInstance
		keys            []resources.ServiceCredentialBinding
	)

	warnings, err := railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			serviceInstance, _, warnings, err = actor.getManagedServiceInstanceByNameAndSpace(serviceInstanceName, spaceGUID)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			keys, warnings, err = actor.CloudControllerClient.GetServiceCredentialBindings(
				ccv3.Query{Key: ccv3.ServiceInstanceGUIDFilter, Values: []string{serviceInstance.GUID}},
				ccv3.Query{Key: ccv3.TypeFilter, Values: []string{"key"}},
			)
			return
		},
	)

	return extract.List("Name", keys), Warnings(warnings), err
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

func (actor Actor) getServiceKeyByServiceInstanceAndName(serviceInstanceName, serviceKeyName, spaceGUID string) (resources.ServiceCredentialBinding, ccv3.Warnings, error) {
	var (
		serviceInstance resources.ServiceInstance
		keys            []resources.ServiceCredentialBinding
	)

	warnings, err := railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			serviceInstance, _, warnings, err = actor.getManagedServiceInstanceByNameAndSpace(serviceInstanceName, spaceGUID)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			keys, warnings, err = actor.CloudControllerClient.GetServiceCredentialBindings(
				ccv3.Query{Key: ccv3.ServiceInstanceGUIDFilter, Values: []string{serviceInstance.GUID}},
				ccv3.Query{Key: ccv3.TypeFilter, Values: []string{"key"}},
				ccv3.Query{Key: ccv3.NameFilter, Values: []string{serviceKeyName}},
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
