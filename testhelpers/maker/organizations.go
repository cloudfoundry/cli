package maker

import "github.com/cloudfoundry/cli/cf/models"

var orgGUID func() string

func init() {
	orgGUID = guidGenerator("org")
}

func NewOrgFields(overrides Overrides) (org models.OrganizationFields) {
	org.Name = "new-org"
	org.GUID = orgGUID()

	if overrides.Has("GUID") {
		org.GUID = overrides.Get("GUID").(string)
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
