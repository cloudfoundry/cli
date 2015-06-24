package plugin_models

type GetOrg_Model struct {
	Guid            string
	Name            string
	QuotaDefinition QuotaFields
	Spaces          []GetSpaces_Model
	Domains         []DomainFields
	SpaceQuotas     []SpaceQuotaFields
}
