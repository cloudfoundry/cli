package maker

import "github.com/cloudfoundry/cli/cf/models"

func NewServiceOffering(label string) models.ServiceOffering {
	return models.ServiceOffering{ServiceOfferingFields: models.ServiceOfferingFields{
		Label:       label,
		GUID:        serviceOfferingGUID(),
		Description: "some service description",
	}}
}

var serviceOfferingGUID = guidGenerator("services")
