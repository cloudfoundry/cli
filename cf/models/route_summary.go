package models

type RouteSummary struct {
	Guid   string
	Host   string
	Domain DomainFields
	Path   string
}

func (r RouteSummary) URL() string {
	return urlStringFromParts(r.Host, r.Domain.Name, r.Path)
}
