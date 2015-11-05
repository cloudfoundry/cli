package models

type DomainFields struct {
	Guid                   string
	Name                   string
	OwningOrganizationGuid string
	Shared                 bool
}

func (model DomainFields) UrlForHostAndPath(host string, port int, path string) string {
	return urlStringFromParts(host, port, model.Name, path)
}
