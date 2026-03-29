package rpc

import (
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3"
	plugin_models "code.cloudfoundry.org/cli/v8/plugin/models"
	"code.cloudfoundry.org/cli/v8/resources"
)

// populateServiceModel maps v3 API resources to the plugin model
func populateServiceModel(
	model *plugin_models.GetService_Model,
	serviceInstance resources.ServiceInstance,
	includedResources ccv3.IncludedResources,
) {
	model.Guid = serviceInstance.GUID
	model.Name = serviceInstance.Name
	model.IsUserProvided = serviceInstance.Type == resources.UserProvidedServiceInstance

	if serviceInstance.DashboardURL.IsSet {
		model.DashboardUrl = serviceInstance.DashboardURL.Value
	}

	// Populate service plan
	if len(includedResources.ServicePlans) > 0 {
		plan := includedResources.ServicePlans[0]
		model.ServicePlan.Name = plan.Name
		model.ServicePlan.Guid = plan.GUID
	}

	// Populate service offering
	if len(includedResources.ServiceOfferings) > 0 {
		offering := includedResources.ServiceOfferings[0]
		model.ServiceOffering.Name = offering.Name
		model.ServiceOffering.DocumentationUrl = offering.DocumentationURL
	}

	// Populate last operation
	model.LastOperation.Type = string(serviceInstance.LastOperation.Type)
	model.LastOperation.State = string(serviceInstance.LastOperation.State)
	model.LastOperation.Description = serviceInstance.LastOperation.Description
	model.LastOperation.CreatedAt = serviceInstance.LastOperation.CreatedAt
	model.LastOperation.UpdatedAt = serviceInstance.LastOperation.UpdatedAt
}
