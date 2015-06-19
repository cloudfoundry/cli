package plugin_models

type OrganizationDetails struct {
	Guid            string
	Name            string
	QuotaDefinition QuotaFields
	Spaces          []SpaceFields
	Domains         []DomainFields
	SpaceQuotas     []SpaceQuotaFields
}
