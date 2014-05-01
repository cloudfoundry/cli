package maker

import "github.com/cloudfoundry/cli/cf/models"

var spaceGuid func() string

func init() {
	spaceGuid = guidGenerator("space")
}

func NewSpaceFields(overrides Overrides) (space models.SpaceFields) {
	space.Name = "new-space"
	space.Guid = spaceGuid()

	if overrides.Has("Guid") {
		space.Guid = overrides.Get("Guid").(string)
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
