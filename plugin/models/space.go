package plugin_models

type SpaceFields struct {
	Guid string
	Name string
}

type Space struct {
	SpaceFields
	Organization OrganizationSummary
	Applications []ApplicationSummary
	//ServiceInstances []ServiceInstances
	Domains        []DomainFields
	SecurityGroups []SecurityGroupFields
	SpaceQuotaGuid string
}
