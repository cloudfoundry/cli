package rpc

import (
	"code.cloudfoundry.org/cli/v8/actor/v7action"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3/constant"
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

// populateOrgsModel maps v3 Organization resources to plugin models
func populateOrgsModel(orgs []resources.Organization) []plugin_models.GetOrgs_Model {
	models := make([]plugin_models.GetOrgs_Model, len(orgs))
	for i, org := range orgs {
		models[i] = plugin_models.GetOrgs_Model{
			Guid: org.GUID,
			Name: org.Name,
		}
	}
	return models
}

// populateSpacesModel maps v3 Space resources to plugin models
func populateSpacesModel(spaces []resources.Space) []plugin_models.GetSpaces_Model {
	models := make([]plugin_models.GetSpaces_Model, len(spaces))
	for i, space := range spaces {
		models[i] = plugin_models.GetSpaces_Model{
			Guid: space.GUID,
			Name: space.Name,
		}
	}
	return models
}

// populateServicesModel maps v7action ServiceInstance list to plugin models
func populateServicesModel(services []v7action.ServiceInstance) []plugin_models.GetServices_Model {
	models := make([]plugin_models.GetServices_Model, len(services))
	for i, svc := range services {
		models[i] = plugin_models.GetServices_Model{
			Name:             svc.Name,
			Guid:             svc.GUID,
			IsUserProvided:   svc.Type == resources.UserProvidedServiceInstance,
			ApplicationNames: svc.BoundApps,
		}

		// Populate service plan
		models[i].ServicePlan = plugin_models.GetServices_ServicePlan{
			Name: svc.ServicePlanName,
		}

		// Populate service fields
		models[i].Service = plugin_models.GetServices_ServiceFields{
			Name: svc.ServiceOfferingName,
		}

		// Populate last operation with separate type and state
		models[i].LastOperation = plugin_models.GetServices_LastOperation{
			Type:  string(svc.LastOperationType),
			State: string(svc.LastOperationState),
		}
	}
	return models
}

// populateOrgUsersModel maps v3 User resources by role type to plugin models
func populateOrgUsersModel(usersByRole map[constant.RoleType][]resources.User) []plugin_models.GetOrgUsers_Model {
	// Create a map to deduplicate users by GUID
	userMap := make(map[string]*plugin_models.GetOrgUsers_Model)

	// Iterate through each role type and its users
	for roleType, users := range usersByRole {
		roleName := string(roleType)
		for _, user := range users {
			if existing, found := userMap[user.GUID]; found {
				// User already exists, add the role
				existing.Roles = append(existing.Roles, roleName)
			} else {
				// New user, create entry
				userMap[user.GUID] = &plugin_models.GetOrgUsers_Model{
					Guid:     user.GUID,
					Username: user.Username,
					IsAdmin:  false, // v3 API doesn't provide this info in role listing
					Roles:    []string{roleName},
				}
			}
		}
	}

	// Convert map to slice
	result := make([]plugin_models.GetOrgUsers_Model, 0, len(userMap))
	for _, user := range userMap {
		result = append(result, *user)
	}
	return result
}

// populateSpaceUsersModel maps v3 User resources by role type to plugin models
func populateSpaceUsersModel(usersByRole map[constant.RoleType][]resources.User) []plugin_models.GetSpaceUsers_Model {
	// Create a map to deduplicate users by GUID
	userMap := make(map[string]*plugin_models.GetSpaceUsers_Model)

	// Iterate through each role type and its users
	for roleType, users := range usersByRole {
		roleName := string(roleType)
		for _, user := range users {
			if existing, found := userMap[user.GUID]; found {
				// User already exists, add the role
				existing.Roles = append(existing.Roles, roleName)
			} else {
				// New user, create entry
				userMap[user.GUID] = &plugin_models.GetSpaceUsers_Model{
					Guid:     user.GUID,
					Username: user.Username,
					IsAdmin:  false, // v3 API doesn't provide this info in role listing
					Roles:    []string{roleName},
				}
			}
		}
	}

	// Convert map to slice
	result := make([]plugin_models.GetSpaceUsers_Model, 0, len(userMap))
	for _, user := range userMap {
		result = append(result, *user)
	}
	return result
}

// populateOrgModel maps v3 Organization and related resources to plugin model
func populateOrgModel(
	org resources.Organization,
	quota resources.OrganizationQuota,
	spaces []resources.Space,
	domains []resources.Domain,
	spaceQuotas []resources.SpaceQuota,
) plugin_models.GetOrg_Model {
	model := plugin_models.GetOrg_Model{
		Guid: org.GUID,
		Name: org.Name,
	}

	// Populate quota definition
	model.QuotaDefinition = plugin_models.QuotaFields{
		Guid: quota.GUID,
		Name: quota.Name,
	}

	// Map memory limits (convert from NullInt)
	if quota.Apps.TotalMemory != nil && quota.Apps.TotalMemory.IsSet {
		model.QuotaDefinition.MemoryLimit = int64(quota.Apps.TotalMemory.Value)
	}
	if quota.Apps.InstanceMemory != nil && quota.Apps.InstanceMemory.IsSet {
		model.QuotaDefinition.InstanceMemoryLimit = int64(quota.Apps.InstanceMemory.Value)
	}

	// Map route limit
	if quota.Routes.TotalRoutes != nil && quota.Routes.TotalRoutes.IsSet {
		model.QuotaDefinition.RoutesLimit = quota.Routes.TotalRoutes.Value
	}

	// Map service limit
	if quota.Services.TotalServiceInstances != nil && quota.Services.TotalServiceInstances.IsSet {
		model.QuotaDefinition.ServicesLimit = quota.Services.TotalServiceInstances.Value
	}

	// Map paid services allowed
	if quota.Services.PaidServicePlans != nil {
		model.QuotaDefinition.NonBasicServicesAllowed = *quota.Services.PaidServicePlans
	}

	// Populate spaces
	model.Spaces = make([]plugin_models.GetOrg_Space, len(spaces))
	for i, space := range spaces {
		model.Spaces[i] = plugin_models.GetOrg_Space{
			Guid: space.GUID,
			Name: space.Name,
		}
	}

	// Populate domains
	model.Domains = make([]plugin_models.GetOrg_Domains, len(domains))
	for i, domain := range domains {
		model.Domains[i] = plugin_models.GetOrg_Domains{
			Guid:                   domain.GUID,
			Name:                   domain.Name,
			OwningOrganizationGuid: domain.OrganizationGUID,
			Shared:                 domain.Shared(),
		}
	}

	// Populate space quotas
	model.SpaceQuotas = make([]plugin_models.GetOrg_SpaceQuota, len(spaceQuotas))
	for i, sq := range spaceQuotas {
		spaceQuota := plugin_models.GetOrg_SpaceQuota{
			Guid: sq.GUID,
			Name: sq.Name,
		}

		// Map space quota limits
		if sq.Apps.TotalMemory != nil && sq.Apps.TotalMemory.IsSet {
			spaceQuota.MemoryLimit = int64(sq.Apps.TotalMemory.Value)
		}
		if sq.Apps.InstanceMemory != nil && sq.Apps.InstanceMemory.IsSet {
			spaceQuota.InstanceMemoryLimit = int64(sq.Apps.InstanceMemory.Value)
		}
		if sq.Routes.TotalRoutes != nil && sq.Routes.TotalRoutes.IsSet {
			spaceQuota.RoutesLimit = sq.Routes.TotalRoutes.Value
		}
		if sq.Services.TotalServiceInstances != nil && sq.Services.TotalServiceInstances.IsSet {
			spaceQuota.ServicesLimit = sq.Services.TotalServiceInstances.Value
		}
		if sq.Services.PaidServicePlans != nil {
			spaceQuota.NonBasicServicesAllowed = *sq.Services.PaidServicePlans
		}

		model.SpaceQuotas[i] = spaceQuota
	}

	return model
}
