package maker

import "cf"

var spaceGuid func() string

func init() {
	spaceGuid = guidGenerator("space")
}

func NewSpaceFields(overrides Overrides) (space cf.SpaceFields) {
	space.Name = "new-space"
	space.Guid = spaceGuid()

	if overrides.Has("guid"){
		space.Guid = overrides.Get("guid").(string)
	}

	if overrides.Has("name") {
		space.Name = overrides.Get("name").(string)
	}
	
	return
}

func NewSpace(overrides Overrides) (space cf.Space) {
	space.SpaceFields = NewSpaceFields(overrides)
	return
}
