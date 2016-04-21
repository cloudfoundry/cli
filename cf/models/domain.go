package models

type DomainFields struct {
	GUID                   string
	Name                   string
	OwningOrganizationGUID string
	RouterGroupGUID        string
	RouterGroupType        string
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
