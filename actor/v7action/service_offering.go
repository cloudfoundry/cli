package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/railway"
)

func (actor Actor) PurgeServiceOfferingByNameAndBroker(serviceOfferingName, serviceBrokerName string) (Warnings, error) {
	var serviceOffering resources.ServiceOffering

	warnings, err := railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			serviceOffering, warnings, err = actor.CloudControllerClient.GetServiceOfferingByNameAndBroker(serviceOfferingName, serviceBrokerName)
			err = actionerror.EnrichAPIErrors(err)
			return
		},
		func() (ccv3.Warnings, error) {
			return actor.CloudControllerClient.PurgeServiceOffering(serviceOffering.GUID)
		},
	)

	return Warnings(warnings), err
}
