package models

type RouteSummary struct {
	GUID    string
	Host    string
	Domain  DomainFields
	Path    string
	Port    int
	Options map[string]string
}

func (r RouteSummary) URL() string {
	return (&RoutePresenter{
		Host:    r.Host,
		Domain:  r.Domain.Name,
		Path:    r.Path,
		Port:    r.Port,
		Options: r.Options,
	}).URL()
}
