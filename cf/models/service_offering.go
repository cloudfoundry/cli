package models

type ServiceOfferingFields struct {
	Guid             string
	Label            string
	Provider         string
	Version          string
	Description      string
	DocumentationUrl string
}

type ServiceOfferings []ServiceOffering
type ServiceOffering struct {
	ServiceOfferingFields
	Plans []ServicePlanFields
}

func (s ServiceOfferings) Len() int {
	return len(s)
}

func (s ServiceOfferings) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ServiceOfferings) Less(i, j int) bool {
	return s[i].Label < s[j].Label
}
