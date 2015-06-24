package plugin_models

type Space struct {
	SpaceSummary
	Organization     OrganizationSummary
	Applications     []GetAppsModel
	ServiceInstances []ServiceInstanceSummary
	Domains          []DomainFields
	SecurityGroups   []SecurityGroupFields
	SpaceQuota       SpaceQuotaFields
}
