package rpc

import (
	"time"

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

	model.DashboardUrl = serviceInstance.DashboardURL.Value

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
	if quota.Apps.TotalMemory != nil {
		model.QuotaDefinition.MemoryLimit = int64(quota.Apps.TotalMemory.Value)
	}
	if quota.Apps.InstanceMemory != nil {
		model.QuotaDefinition.InstanceMemoryLimit = int64(quota.Apps.InstanceMemory.Value)
	}

	// Map route limit
	if quota.Routes.TotalRoutes != nil {
		model.QuotaDefinition.RoutesLimit = quota.Routes.TotalRoutes.Value
	}

	// Map service limit
	if quota.Services.TotalServiceInstances != nil {
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
		if sq.Apps.TotalMemory != nil {
			spaceQuota.MemoryLimit = int64(sq.Apps.TotalMemory.Value)
		}
		if sq.Apps.InstanceMemory != nil {
			spaceQuota.InstanceMemoryLimit = int64(sq.Apps.InstanceMemory.Value)
		}
		if sq.Routes.TotalRoutes != nil {
			spaceQuota.RoutesLimit = sq.Routes.TotalRoutes.Value
		}
		if sq.Services.TotalServiceInstances != nil {
			spaceQuota.ServicesLimit = sq.Services.TotalServiceInstances.Value
		}
		if sq.Services.PaidServicePlans != nil {
			spaceQuota.NonBasicServicesAllowed = *sq.Services.PaidServicePlans
		}

		model.SpaceQuotas[i] = spaceQuota
	}

	return model
}

// populateSpaceModel maps v3 Space and related resources to plugin model
func populateSpaceModel(
	space resources.Space,
	orgGUID string,
	orgName string,
	apps []resources.Application,
	serviceInstances []v7action.ServiceInstance,
	domains []resources.Domain,
	spaceQuota resources.SpaceQuota,
	securityGroups []resources.SecurityGroup,
) plugin_models.GetSpace_Model {
	model := plugin_models.GetSpace_Model{
		GetSpaces_Model: plugin_models.GetSpaces_Model{
			Guid: space.GUID,
			Name: space.Name,
		},
		Organization: plugin_models.GetSpace_Orgs{
			Guid: orgGUID,
			Name: orgName,
		},
	}

	// Populate applications
	model.Applications = make([]plugin_models.GetSpace_Apps, len(apps))
	for i, app := range apps {
		model.Applications[i] = plugin_models.GetSpace_Apps{
			Guid: app.GUID,
			Name: app.Name,
		}
	}

	// Populate service instances
	model.ServiceInstances = make([]plugin_models.GetSpace_ServiceInstance, len(serviceInstances))
	for i, svc := range serviceInstances {
		model.ServiceInstances[i] = plugin_models.GetSpace_ServiceInstance{
			Guid: svc.GUID,
			Name: svc.Name,
		}
	}

	// Populate domains
	model.Domains = make([]plugin_models.GetSpace_Domains, len(domains))
	for i, domain := range domains {
		model.Domains[i] = plugin_models.GetSpace_Domains{
			Guid:                   domain.GUID,
			Name:                   domain.Name,
			OwningOrganizationGuid: domain.OrganizationGUID,
			Shared:                 domain.Shared(),
		}
	}

	// Populate security groups
	model.SecurityGroups = make([]plugin_models.GetSpace_SecurityGroup, len(securityGroups))
	for i, sg := range securityGroups {
		// Convert rules to map[string]interface{}
		rules := make([]map[string]interface{}, len(sg.Rules))
		for j, rule := range sg.Rules {
			ruleMap := map[string]interface{}{
				"protocol":    rule.Protocol,
				"destination": rule.Destination,
			}
			if rule.Ports != nil {
				ruleMap["ports"] = *rule.Ports
			}
			if rule.Type != nil {
				ruleMap["type"] = *rule.Type
			}
			if rule.Code != nil {
				ruleMap["code"] = *rule.Code
			}
			if rule.Description != nil {
				ruleMap["description"] = *rule.Description
			}
			if rule.Log != nil {
				ruleMap["log"] = *rule.Log
			}
			rules[j] = ruleMap
		}

		model.SecurityGroups[i] = plugin_models.GetSpace_SecurityGroup{
			Guid:  sg.GUID,
			Name:  sg.Name,
			Rules: rules,
		}
	}

	// Populate space quota (if exists)
	if spaceQuota.GUID != "" {
		model.SpaceQuota = plugin_models.GetSpace_SpaceQuota{
			Guid: spaceQuota.GUID,
			Name: spaceQuota.Name,
		}

		// Map space quota limits
		if spaceQuota.Apps.TotalMemory != nil {
			model.SpaceQuota.MemoryLimit = int64(spaceQuota.Apps.TotalMemory.Value)
		}
		if spaceQuota.Apps.InstanceMemory != nil {
			model.SpaceQuota.InstanceMemoryLimit = int64(spaceQuota.Apps.InstanceMemory.Value)
		}
		if spaceQuota.Routes.TotalRoutes != nil {
			model.SpaceQuota.RoutesLimit = spaceQuota.Routes.TotalRoutes.Value
		}
		if spaceQuota.Services.TotalServiceInstances != nil {
			model.SpaceQuota.ServicesLimit = spaceQuota.Services.TotalServiceInstances.Value
		}
		if spaceQuota.Services.PaidServicePlans != nil {
			model.SpaceQuota.NonBasicServicesAllowed = *spaceQuota.Services.PaidServicePlans
		}
	}

	return model
}

// populateAppModel maps v3 DetailedApplicationSummary to plugin model
func populateAppModel(
	summary v7action.DetailedApplicationSummary,
	serviceBindings []resources.ServiceCredentialBinding,
	stack resources.Stack,
) plugin_models.GetAppModel {
	model := plugin_models.GetAppModel{
		Guid:      summary.GUID,
		Name:      summary.Name,
		State:     string(summary.State),
		SpaceGuid: summary.SpaceGUID,
	}

	// Find the web process for main app details
	var webProcess *v7action.ProcessSummary
	for i := range summary.ProcessSummaries {
		if summary.ProcessSummaries[i].Type == constant.ProcessTypeWeb {
			webProcess = &summary.ProcessSummaries[i]
			break
		}
	}

	if webProcess != nil {
		// Memory in MB
		model.Memory = int64(webProcess.MemoryInMB.Value)

		// Disk quota in MB
		model.DiskQuota = int64(webProcess.DiskInMB.Value)

		// Instance count
		model.InstanceCount = webProcess.Instances.Value

		// Running instances
		model.RunningInstances = webProcess.HealthyInstanceCount()

		// Health check timeout
		model.HealthCheckTimeout = int(webProcess.HealthCheckTimeout)

		// Command
		model.Command = webProcess.Command.Value

		// Map instances
		model.Instances = make([]plugin_models.GetApp_AppInstanceFields, len(webProcess.InstanceDetails))
		for i, inst := range webProcess.InstanceDetails {
			// Calculate actual time from uptime duration
			since := time.Now().Add(-inst.Uptime)
			model.Instances[i] = plugin_models.GetApp_AppInstanceFields{
				State:     string(inst.State),
				Details:   inst.Details,
				Since:     since,
				CpuUsage:  inst.CPU,
				DiskQuota: int64(inst.DiskQuota),
				DiskUsage: int64(inst.DiskUsage),
				MemQuota:  int64(inst.MemoryQuota),
				MemUsage:  int64(inst.MemoryUsage),
			}
		}
	}

	// Populate buildpack info from droplet
	if len(summary.CurrentDroplet.Buildpacks) > 0 {
		// Use the first buildpack as BuildpackUrl (legacy compatibility)
		firstBuildpack := summary.CurrentDroplet.Buildpacks[0]
		if firstBuildpack.Name != "" {
			model.BuildpackUrl = firstBuildpack.Name
		} else if firstBuildpack.BuildpackName != "" {
			model.BuildpackUrl = firstBuildpack.BuildpackName
		}
	}

	// Populate stack
	if stack.GUID != "" {
		model.Stack = &plugin_models.GetApp_Stack{
			Guid:        stack.GUID,
			Name:        stack.Name,
			Description: stack.Description,
		}
	}

	// Populate routes
	model.Routes = make([]plugin_models.GetApp_RouteSummary, len(summary.Routes))
	for i, route := range summary.Routes {
		model.Routes[i] = plugin_models.GetApp_RouteSummary{
			Guid: route.GUID,
			Host: route.Host,
			Path: route.Path,
			Port: route.Port,
			Domain: plugin_models.GetApp_DomainFields{
				Guid: route.DomainGUID,
				Name: route.URL, // URL contains the full domain
			},
		}
	}

	// Populate bound services
	model.Services = make([]plugin_models.GetApp_ServiceSummary, len(serviceBindings))
	for i, binding := range serviceBindings {
		model.Services[i] = plugin_models.GetApp_ServiceSummary{
			Guid: binding.ServiceInstanceGUID,
			Name: binding.Name,
		}
	}

	// Populate package state from droplet state
	model.PackageState = string(summary.CurrentDroplet.State)

	return model
}

// populateAppsModel maps v3 ApplicationSummary to plugin GetAppsModel
func populateAppsModel(summaries []v7action.ApplicationSummary) []plugin_models.GetAppsModel {
	models := make([]plugin_models.GetAppsModel, len(summaries))

	for i, summary := range summaries {
		model := plugin_models.GetAppsModel{
			Name:  summary.Name,
			Guid:  summary.GUID,
			State: string(summary.State),
		}

		// Get web process if it exists
		var webProcess *v7action.ProcessSummary
		for j := range summary.ProcessSummaries {
			if summary.ProcessSummaries[j].Type == constant.ProcessTypeWeb {
				webProcess = &summary.ProcessSummaries[j]
				break
			}
		}

		// Populate instance and resource info from web process
		if webProcess != nil {
			model.TotalInstances = webProcess.TotalInstanceCount()
			model.RunningInstances = webProcess.HealthyInstanceCount()
			model.Memory = int64(webProcess.MemoryInMB.Value)
			model.DiskQuota = int64(webProcess.DiskInMB.Value)
		}

		// Populate routes
		model.Routes = make([]plugin_models.GetAppsRouteSummary, len(summary.Routes))
		for j, route := range summary.Routes {
			model.Routes[j] = plugin_models.GetAppsRouteSummary{
				Guid: route.GUID,
				Host: route.Host,
				Domain: plugin_models.GetAppsDomainFields{
					Guid: route.DomainGUID,
					Name: route.URL,
				},
			}
		}

		models[i] = model
	}

	return models
}
