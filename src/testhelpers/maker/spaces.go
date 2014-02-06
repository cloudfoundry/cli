package maker

import "cf/models"

var spaceGuid func() string

func init() {
	spaceGuid = guidGenerator("space")
}

func NewSpaceFields(overrides Overrides) (space models.SpaceFields) {
	space.Name = "new-space"
	space.Guid = spaceGuid()

	if overrides.Has("guid") {
		space.Guid = overrides.Get("guid").(string)
	}

	if overrides.Has("name") {
		space.Name = overrides.Get("name").(string)
	}

	return
}

func NewSpace(overrides Overrides) (space models.Space) {
	space.SpaceFields = NewSpaceFields(overrides)
	return
}
