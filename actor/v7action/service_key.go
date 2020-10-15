package v7action

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/railway"
)

func (actor Actor) GetServiceKeysByServiceInstance(serviceInstanceName, spaceGUID string) ([]string, Warnings, error) {
	var (
		serviceInstance resources.ServiceInstance
		keys            []resources.ServiceCredentialBinding
	)

	warnings, err := railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			serviceInstance, warnings, err = actor.getServiceInstanceByNameAndSpace(serviceInstanceName, spaceGUID)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			err = assertServiceInstanceType(resources.ManagedServiceInstance, serviceInstance)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			keys, warnings, err = actor.CloudControllerClient.GetServiceCredentialBindings([]ccv3.Query{
				{Key: ccv3.ServiceInstanceGUIDFilter, Values: []string{serviceInstance.GUID}},
				{Key: ccv3.TypeFilter, Values: []string{"key"}},
			}...)
			return
		},
	)

	var result []string
	for _, k := range keys {
		result = append(result, k.Name)
	}

	return result, Warnings(warnings), err
}
