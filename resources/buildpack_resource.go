package resources

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/types"
)

// Buildpack represents a Cloud Controller V3 buildpack.
type Buildpack struct {
	// Enabled is true when the buildpack can be used for staging.
	Enabled types.NullBool
	// Filename is the uploaded filename of the buildpack.
	Filename string
	// GUID is the unique identifier for the buildpack.
	GUID string
	// Locked is true when the buildpack cannot be updated.
	Locked types.NullBool
	// Name is the name of the buildpack. To be used by app buildpack field.
	// (only alphanumeric characters)
	Name string
	// Position is the order in which the buildpacks are checked during buildpack
	// auto-detection.
	Position types.NullInt
	// Stack is the name of the stack that the buildpack will use.
	Stack string
	// State is the current state of the buildpack.
	State string
	// Links are links to related resources.
	Links APILinks
	// Metadata is used for custom tagging of API resources
	Metadata *Metadata
	// Lifecycle is the lifecycle used with this buildpack
	Lifecycle string
}

// MarshalJSON converts a Package into a Cloud Controller Package.
func (buildpack Buildpack) MarshalJSON() ([]byte, error) {
	ccBuildpack := struct {
		Name      string    `json:"name,omitempty"`
		Stack     string    `json:"stack,omitempty"`
		Position  *int      `json:"position,omitempty"`
		Enabled   *bool     `json:"enabled,omitempty"`
		Locked    *bool     `json:"locked,omitempty"`
		Metadata  *Metadata `json:"metadata,omitempty"`
		Lifecycle string    `json:"lifecycle,omitempty"`
	}{
		Name:      buildpack.Name,
		Stack:     buildpack.Stack,
		Lifecycle: buildpack.Lifecycle,
	}

	if buildpack.Position.IsSet {
		ccBuildpack.Position = &buildpack.Position.Value
	}
	if buildpack.Enabled.IsSet {
		ccBuildpack.Enabled = &buildpack.Enabled.Value
	}
	if buildpack.Locked.IsSet {
		ccBuildpack.Locked = &buildpack.Locked.Value
	}

	return json.Marshal(ccBuildpack)
}

func (buildpack *Buildpack) UnmarshalJSON(data []byte) error {
	var ccBuildpack struct {
		GUID      string         `json:"guid,omitempty"`
		Links     APILinks       `json:"links,omitempty"`
		Name      string         `json:"name,omitempty"`
		Filename  string         `json:"filename,omitempty"`
		Stack     string         `json:"stack,omitempty"`
		State     string         `json:"state,omitempty"`
		Enabled   types.NullBool `json:"enabled"`
		Locked    types.NullBool `json:"locked"`
		Position  types.NullInt  `json:"position"`
		Metadata  *Metadata      `json:"metadata"`
		Lifecycle string         `json:"lifecycle"`
	}

	err := cloudcontroller.DecodeJSON(data, &ccBuildpack)
	if err != nil {
		return err
	}

	buildpack.Enabled = ccBuildpack.Enabled
	buildpack.Filename = ccBuildpack.Filename
	buildpack.GUID = ccBuildpack.GUID
	buildpack.Locked = ccBuildpack.Locked
	buildpack.Name = ccBuildpack.Name
	buildpack.Position = ccBuildpack.Position
	buildpack.Stack = ccBuildpack.Stack
	buildpack.State = ccBuildpack.State
	buildpack.Links = ccBuildpack.Links
	buildpack.Metadata = ccBuildpack.Metadata
	buildpack.Lifecycle = ccBuildpack.Lifecycle

	return nil
}
