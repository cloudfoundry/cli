package maker

import "github.com/cloudfoundry/cli/cf/models"

var domainGUID func() string = guidGenerator("domain")

func NewSharedDomainFields(overrides Overrides) (domain models.DomainFields) {
	domain.Name = "new-domain"
	domain.GUID = domainGUID()
	domain.Shared = true

	if overrides.Has("Name") {
		domain.Name = overrides.Get("Name").(string)
	}
	if overrides.Has("GUID") {
		domain.GUID = overrides.Get("GUID").(string)
	}
	return
}

func NewPrivateDomainFields(overrides Overrides) (domain models.DomainFields) {
	domain.Name = "new-domain"
	domain.GUID = domainGUID()
	domain.Shared = false

	if overrides.Has("Name") {
		domain.Name = overrides.Get("Name").(string)
	}
	if overrides.Has("GUID") {
		domain.GUID = overrides.Get("GUID").(string)
	}
	return
}
