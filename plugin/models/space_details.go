package plugin_models

type SpaceDetails struct {
	SpaceFields
	Organization OrganizationSummary
	Applications []ApplicationSummary
	//ServiceInstances []ServiceInstances
	Domains        []DomainFields
	SecurityGroups []SecurityGroupFields
	SpaceQuotaGuid string
}
