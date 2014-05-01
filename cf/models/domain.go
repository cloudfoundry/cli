package models

import "fmt"

type DomainFields struct {
	Guid                   string
	Name                   string
	OwningOrganizationGuid string
	Shared                 bool
}

func (model DomainFields) UrlForHost(host string) string {
	if host == "" {
		return model.Name
	}
	return fmt.Sprintf("%s.%s", host, model.Name)
}
