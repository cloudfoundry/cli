package plugin_models

type OrganizationFields struct {
	Guid            string
	Name            string
	QuotaDefinition QuotaFields
}

type Organization struct {
	OrganizationFields
}
