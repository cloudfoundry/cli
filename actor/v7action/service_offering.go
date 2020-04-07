package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

type ServiceOffering ccv3.ServiceOffering

func (actor Actor) PurgeServiceOfferingByNameAndBroker(serviceOfferingName, serviceBrokerName string) (Warnings, error) {
	serviceOffering, warnings, err := actor.CloudControllerClient.GetServiceOfferingByNameAndBroker(serviceOfferingName, serviceBrokerName)
	allWarnings := Warnings(warnings)
	if err != nil {
		return allWarnings, actionerror.EnrichAPIErrors(err)
	}

	warnings, err = actor.CloudControllerClient.PurgeServiceOffering(serviceOffering.GUID)
	allWarnings = append(allWarnings, warnings...)
	return allWarnings, err
}
