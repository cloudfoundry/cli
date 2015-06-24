package plugin_models

type Organization struct {
	Guid            string
	Name            string
	QuotaDefinition QuotaFields
	Spaces          []GetSpaces_Model
	Domains         []DomainFields
	SpaceQuotas     []SpaceQuotaFields
}
