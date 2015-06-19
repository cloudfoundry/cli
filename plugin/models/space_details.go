package plugin_models

type SpaceDetails struct {
	SpaceFields
	Organization     OrganizationSummary
	Applications     []ApplicationSummary
	ServiceInstances []ServiceInstanceSummary
	Domains          []DomainFields
	SecurityGroups   []SecurityGroupFields
	SpaceQuota       SpaceQuotaFields
}
