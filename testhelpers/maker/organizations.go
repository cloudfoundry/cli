package maker

import "github.com/cloudfoundry/cli/cf/models"

var orgGuid func() string

func init() {
	orgGuid = guidGenerator("org")
}

func NewOrgFields(overrides Overrides) (org models.OrganizationFields) {
	org.Name = "new-org"
	org.Guid = orgGuid()

	if overrides.Has("Guid") {
		org.Guid = overrides.Get("Guid").(string)
	}

	if overrides.Has("Name") {
		org.Name = overrides.Get("Name").(string)
	}

	return
}

func NewOrg(overrides Overrides) (org models.Organization) {
	org.OrganizationFields = NewOrgFields(overrides)
	return
}
