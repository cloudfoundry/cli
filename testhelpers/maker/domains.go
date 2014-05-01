package maker

import "github.com/cloudfoundry/cli/cf/models"

var domainGuid func() string = guidGenerator("domain")

func NewSharedDomainFields(overrides Overrides) (domain models.DomainFields) {
	domain.Name = "new-domain"
	domain.Guid = domainGuid()
	domain.Shared = true

	if overrides.Has("Name") {
		domain.Name = overrides.Get("Name").(string)
	}
	if overrides.Has("Guid") {
		domain.Guid = overrides.Get("Guid").(string)
	}
	return
}

func NewPrivateDomainFields(overrides Overrides) (domain models.DomainFields) {
	domain.Name = "new-domain"
	domain.Guid = domainGuid()
	domain.Shared = false

	if overrides.Has("Name") {
		domain.Name = overrides.Get("Name").(string)
	}
	if overrides.Has("Guid") {
		domain.Guid = overrides.Get("Guid").(string)
	}
	return
}
