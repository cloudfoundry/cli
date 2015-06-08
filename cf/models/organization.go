package models

type OrganizationFields struct {
	Guid            string
	Name            string
	QuotaDefinition QuotaFields
}

type Organization struct {
	OrganizationFields
	Spaces      []SpaceFields
	Domains     []DomainFields
	SpaceQuotas []SpaceQuota
}

type Organizations []Organization

func (o Organizations) Len() int {
	return len(o)
}

func (o Organizations) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}

func (o Organizations) Less(i, j int) bool {
	return o[i].Name < o[j].Name
}
