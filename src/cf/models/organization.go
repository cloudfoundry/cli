package models

type OrganizationFields struct {
	Guid            string
	Name            string
	QuotaDefinition QuotaFields
}

type Organization struct {
	OrganizationFields
	Spaces  []SpaceFields
	Domains []DomainFields
}
