package models

type SpaceFields struct {
	Guid string
	Name string
}

type Space struct {
	SpaceFields
	Organization     OrganizationFields
	Applications     []ApplicationFields
	ServiceInstances []ServiceInstanceFields
	Domains          []DomainFields
}
