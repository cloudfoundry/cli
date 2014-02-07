package maker

import "cf/models"

var domainGuid func() string = guidGenerator("domain")

func NewSharedDomainFields(overrides Overrides) (domain models.DomainFields) {
	domain.Name = "new-domain"
	domain.Guid = domainGuid()
	domain.Shared = true

	if overrides.Has("name") {
		domain.Name = overrides.Get("name").(string)
	}
	if overrides.Has("guid") {
		domain.Guid = overrides.Get("guid").(string)
	}
	return
}

func NewPrivateDomainFields(overrides Overrides) (domain models.DomainFields) {
	domain.Name = "new-domain"
	domain.Guid = domainGuid()
	domain.Shared = false

	if overrides.Has("name") {
		domain.Name = overrides.Get("name").(string)
	}
	if overrides.Has("guid") {
		domain.Guid = overrides.Get("guid").(string)
	}
	return
}
