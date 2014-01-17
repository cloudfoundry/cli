package maker

import "cf"

var orgGuid func() string

func init() {
	orgGuid = guidGenerator("org")
}

func NewOrgFields(overrides Overrides) (org cf.OrganizationFields) {
	org.Name = "new-org"
	org.Guid = orgGuid()

	if overrides.Has("guid") {
		org.Guid = overrides.Get("guid").(string)
	}

	if overrides.Has("name") {
		org.Name = overrides.Get("name").(string)
	}

	return
}

func NewOrg(overrides Overrides) (org cf.Organization) {
	org.OrganizationFields = NewOrgFields(overrides)
	return
}
