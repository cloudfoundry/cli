package resources

import "code.cloudfoundry.org/cli/v9/cf/models"

type BuildpackResource struct {
	Resource
	Entity BuildpackEntity
}

type BuildpackEntity struct {
	Name     string `json:"name"`
	Stack    string `json:"stack,omitempty"`
	Position *int   `json:"position,omitempty"`
	Enabled  *bool  `json:"enabled,omitempty"`
	Key      string `json:"key,omitempty"`
	Filename string `json:"filename,omitempty"`
	Locked   *bool  `json:"locked,omitempty"`
}

func (resource BuildpackResource) ToFields() models.Buildpack {
	return models.Buildpack{
		GUID:     resource.Metadata.GUID,
		Name:     resource.Entity.Name,
		Stack:    resource.Entity.Stack,
		Position: resource.Entity.Position,
		Enabled:  resource.Entity.Enabled,
		Key:      resource.Entity.Key,
		Filename: resource.Entity.Filename,
		Locked:   resource.Entity.Locked,
	}
}
