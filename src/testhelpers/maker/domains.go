package maker

import "cf/models"

var domainGuid func() string = guidGenerator("domain")

func NewSharedDomain(overrides Overrides) (domain models.Domain) {
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

func NewPrivateDomain(overrides Overrides) (domain models.Domain) {
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
