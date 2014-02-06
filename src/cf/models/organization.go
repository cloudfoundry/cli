package models

type OrganizationFields struct {
	BasicFields
	QuotaDefinition QuotaFields
}

type Organization struct {
	OrganizationFields
	Spaces  []SpaceFields
	Domains []DomainFields
}
