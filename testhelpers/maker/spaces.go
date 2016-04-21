package maker

import "github.com/cloudfoundry/cli/cf/models"

var spaceGUID func() string

func init() {
	spaceGUID = guidGenerator("space")
}

func NewSpaceFields(overrides Overrides) (space models.SpaceFields) {
	space.Name = "new-space"
	space.GUID = spaceGUID()

	if overrides.Has("GUID") {
		space.GUID = overrides.Get("GUID").(string)
	}

	if overrides.Has("Name") {
		space.Name = overrides.Get("Name").(string)
	}

	return
}

func NewSpace(overrides Overrides) (space models.Space) {
	space.SpaceFields = NewSpaceFields(overrides)

	if overrides.Has("Organization") {
		space.Organization = overrides.Get("Organization").(models.OrganizationFields)
	}

	return
}
