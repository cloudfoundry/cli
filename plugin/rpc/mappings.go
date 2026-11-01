package rpc

import (
	"strings"
	"time"

	"code.cloudfoundry.org/cli/v8/actor/v7action"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3/constant"
	plugin_models "code.cloudfoundry.org/cli/v8/plugin/models"
	"code.cloudfoundry.org/cli/v8/resources"
	"code.cloudfoundry.org/cli/v8/types"
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
			Guid: svc.ServicePlanGUID,
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

// legacyRoleNames maps v3 role types to the role names legacy plugins received
// (cf/models.Role.ToString). space_supporter has no legacy equivalent.
var legacyRoleNames = map[constant.RoleType]string{
	constant.OrgUserRole:           "RoleOrgUser",
	constant.OrgManagerRole:        "RoleOrgManager",
	constant.OrgBillingManagerRole: "RoleBillingManager",
	constant.OrgAuditorRole:        "RoleOrgAuditor",
	constant.SpaceManagerRole:      "RoleSpaceManager",
	constant.SpaceDeveloperRole:    "RoleSpaceDeveloper",
	constant.SpaceAuditorRole:      "RoleSpaceAuditor",
}

// Role selections and order of the legacy org-users / space-users commands.
var (
	orgUsersDefaultRoles = []constant.RoleType{constant.OrgManagerRole, constant.OrgBillingManagerRole, constant.OrgAuditorRole}
	orgUsersAllFlagRoles = []constant.RoleType{constant.OrgUserRole}
	spaceUsersRoles      = []constant.RoleType{constant.SpaceManagerRole, constant.SpaceDeveloperRole, constant.SpaceAuditorRole}
)

type userWithRoles struct {
	user  resources.User
	roles []string
}

// collectUsersByRoles walks the given roles in order and their users in API
// order, returning each user once with its legacy role names in that order.
func collectUsersByRoles(usersByRole map[constant.RoleType][]resources.User, roles []constant.RoleType) []userWithRoles {
	var result []userWithRoles
	indexByGUID := make(map[string]int)

	for _, roleType := range roles {
		roleName := legacyRoleNames[roleType]
		for _, user := range usersByRole[roleType] {
			if i, found := indexByGUID[user.GUID]; found {
				result[i].roles = append(result[i].roles, roleName)
				continue
			}
			indexByGUID[user.GUID] = len(result)
			result = append(result, userWithRoles{user: user, roles: []string{roleName}})
		}
	}
	return result
}

// populateOrgUsersModel maps v3 users of the given org roles to plugin models
func populateOrgUsersModel(usersByRole map[constant.RoleType][]resources.User, roles []constant.RoleType) []plugin_models.GetOrgUsers_Model {
	users := collectUsersByRoles(usersByRole, roles)
	result := make([]plugin_models.GetOrgUsers_Model, len(users))
	for i, u := range users {
		result[i] = plugin_models.GetOrgUsers_Model{
			Guid:     u.user.GUID,
			Username: u.user.Username,
			IsAdmin:  false, // as in legacy, which listed users without querying UAA
			Roles:    u.roles,
		}
	}
	return result
}

// populateSpaceUsersModel maps v3 users of the legacy space roles to plugin models
func populateSpaceUsersModel(usersByRole map[constant.RoleType][]resources.User) []plugin_models.GetSpaceUsers_Model {
	users := collectUsersByRoles(usersByRole, spaceUsersRoles)
	result := make([]plugin_models.GetSpaceUsers_Model, len(users))
	for i, u := range users {
		result[i] = plugin_models.GetSpaceUsers_Model{
			Guid:     u.user.GUID,
			Username: u.user.Username,
			IsAdmin:  false, // as in legacy, which listed users without querying UAA
			Roles:    u.roles,
		}
	}
	return result
}

// quotaLimit maps a v3 quota limit to the legacy plugin value: CAPI v3 reports an
// unlimited quota as null (unset), which legacy (v2) plugins received as -1.
func quotaLimit(limit *types.NullInt) int64 {
	if limit == nil || !limit.IsSet {
		return -1
	}
	return int64(limit.Value)
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

	model.QuotaDefinition.MemoryLimit = quotaLimit(quota.Apps.TotalMemory)
	model.QuotaDefinition.InstanceMemoryLimit = quotaLimit(quota.Apps.InstanceMemory)
	model.QuotaDefinition.RoutesLimit = int(quotaLimit(quota.Routes.TotalRoutes))
	model.QuotaDefinition.ServicesLimit = int(quotaLimit(quota.Services.TotalServiceInstances))

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

		spaceQuota.MemoryLimit = quotaLimit(sq.Apps.TotalMemory)
		spaceQuota.InstanceMemoryLimit = quotaLimit(sq.Apps.InstanceMemory)
		spaceQuota.RoutesLimit = int(quotaLimit(sq.Routes.TotalRoutes))
		spaceQuota.ServicesLimit = int(quotaLimit(sq.Services.TotalServiceInstances))
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

		model.SpaceQuota.MemoryLimit = quotaLimit(spaceQuota.Apps.TotalMemory)
		model.SpaceQuota.InstanceMemoryLimit = quotaLimit(spaceQuota.Apps.InstanceMemory)
		model.SpaceQuota.RoutesLimit = int(quotaLimit(spaceQuota.Routes.TotalRoutes))
		model.SpaceQuota.ServicesLimit = int(quotaLimit(spaceQuota.Services.TotalServiceInstances))
		if spaceQuota.Services.PaidServicePlans != nil {
			model.SpaceQuota.NonBasicServicesAllowed = *spaceQuota.Services.PaidServicePlans
		}
	}

	return model
}

// primaryProcess returns the process that represents the app in the legacy
// plugin models: the web process, or the first process when there is no web
// process, or nil when the app has no processes.
func primaryProcess(processes v7action.ProcessSummaries) *v7action.ProcessSummary {
	for i := range processes {
		if processes[i].Type == constant.ProcessTypeWeb {
			return &processes[i]
		}
	}
	if len(processes) > 0 {
		return &processes[0]
	}
	return nil
}

// populateAppModel maps v3 DetailedApplicationSummary to plugin model
func populateAppModel(
	summary v7action.DetailedApplicationSummary,
	serviceBindings []resources.ServiceCredentialBinding,
	stack resources.Stack,
	envVars map[string]interface{},
) plugin_models.GetAppModel {
	model := plugin_models.GetAppModel{
		EnvironmentVars: envVars,
		Guid:            summary.GUID,
		Name:            summary.Name,
		State:           strings.ToLower(string(summary.State)),
		SpaceGuid:       summary.SpaceGUID,
	}

	process := primaryProcess(summary.ProcessSummaries)
	if process != nil {
		// Memory in MB
		model.Memory = int64(process.MemoryInMB.Value)

		// Disk quota in MB
		model.DiskQuota = int64(process.DiskInMB.Value)

		// Instance count
		model.InstanceCount = process.Instances.Value

		// Running instances
		model.RunningInstances = process.HealthyInstanceCount()

		// Health check timeout
		model.HealthCheckTimeout = int(process.HealthCheckTimeout)

		// Command
		model.Command = process.Command.Value

		// Map instances
		model.Instances = make([]plugin_models.GetApp_AppInstanceFields, len(process.InstanceDetails))
		for i, inst := range process.InstanceDetails {
			// Calculate actual time from uptime duration
			since := time.Now().Add(-inst.Uptime)
			model.Instances[i] = plugin_models.GetApp_AppInstanceFields{
				State:     strings.ToLower(string(inst.State)),
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
				Name: v7action.DomainNameFromRouteURL(route.URL, route.Host, route.Path, route.Port),
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
			State: strings.ToLower(string(summary.State)),
		}

		process := primaryProcess(summary.ProcessSummaries)
		if process != nil {
			model.TotalInstances = process.Instances.Value // desired scale, as legacy reported
			model.RunningInstances = process.HealthyInstanceCount()
			model.Memory = int64(process.MemoryInMB.Value)
			model.DiskQuota = int64(process.DiskInMB.Value)
		}

		// Populate routes
		model.Routes = make([]plugin_models.GetAppsRouteSummary, len(summary.Routes))
		for j, route := range summary.Routes {
			model.Routes[j] = plugin_models.GetAppsRouteSummary{
				Guid: route.GUID,
				Host: route.Host,
				Domain: plugin_models.GetAppsDomainFields{
					Guid: route.DomainGUID,
					Name: v7action.DomainNameFromRouteURL(route.URL, route.Host, route.Path, route.Port),
				},
			}
		}

		models[i] = model
	}

	return models
}
