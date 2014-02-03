package maker

import "cf"

func NewServiceOffering(label string) (service cf.ServiceOffering) {
	service.Label = label
	service.Guid = serviceOfferingGuid()
	service.Description = "some service description"
	return
}

var serviceOfferingGuid  = guidGenerator("services")
