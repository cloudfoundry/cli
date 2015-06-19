package plugin_models

type Space struct {
	SpaceSummary
	Organization     OrganizationSummary
	Applications     []ApplicationSummary
	ServiceInstances []ServiceInstanceSummary
	Domains          []DomainFields
	SecurityGroups   []SecurityGroupFields
	SpaceQuota       SpaceQuotaFields
}
