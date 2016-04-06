package models

type DomainFields struct {
	Guid                   string
	Name                   string
	OwningOrganizationGuid string
	RouterGroupGuid        string
	RouterGroupTypes       []string
	Shared                 bool
}

func (model DomainFields) UrlForHostAndPath(host, path string, port int) string {
	return (&RoutePresenter{
		Host:   host,
		Domain: model.Name,
		Path:   path,
		Port:   port,
	}).URL()
}
