package models

type DomainFields struct {
	Guid                   string
	Name                   string
	OwningOrganizationGuid string
	Shared                 bool
}

func (model DomainFields) UrlForHostAndPath(host, path string) string {
	return urlStringFromParts(host, model.Name, path)
}
