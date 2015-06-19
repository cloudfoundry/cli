package plugin_models

type Organization struct {
	Guid            string
	Name            string
	QuotaDefinition QuotaFields
	Spaces          []SpaceFields
	Domains         []DomainFields
	SpaceQuotas     []SpaceQuotaFields
}
