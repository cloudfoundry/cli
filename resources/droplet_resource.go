package resources

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"encoding/json"
)

// Droplet represents a Cloud Controller droplet's metadata. A droplet is a set of
// compiled bits for a given application.
type Droplet struct {
	// AppGUID is the unique identifier of the application associated with the droplet.
	AppGUID string `json:"app_guid"`
	//Buildpacks are the detected buildpacks from the staging process.
	Buildpacks []DropletBuildpack `json:"buildpacks,omitempty"`
	// CreatedAt is the timestamp that the Cloud Controller created the droplet.
	CreatedAt string `json:"created_at"`
	// GUID is the unique droplet identifier.
	GUID string `json:"guid"`
	// Image is the Docker image name.
	Image string `json:"image"`
	// Stack is the root filesystem to use with the buildpack.
	Stack string `json:"stack,omitempty"`
	// State is the current state of the droplet.
	State constant.DropletState `json:"state"`
	// IsCurrent does not exist on the API layer, only on the actor layer; we ignore it w.r.t. JSON
	IsCurrent bool `json:"-"`
}

// DropletBuildpack is the name and output of a buildpack used to create a
// droplet.
type DropletBuildpack struct {
	// Name is the buildpack name.
	Name string `json:"name"`
	// BuildpackName is the name reported by the buildpack.
	BuildpackName string `json:"buildpack_name"`
	// DetectOutput is the output of the buildpack detect script.
	DetectOutput string `json:"detect_output"`
	// Version is the version of the detected buildpack.
	Version string `json:"version"`
}

func (d Droplet) MarshallJSON() ([]byte, error) {
	type Data struct {
		GUID string `json:"guid,omitempty"`
	}

	type RelationshipData struct {
		Data Data `json:"data,omitempty"`
	}

	type Relationships struct {
		App RelationshipData `json:"app,omitempty"`
	}

	type ccDroplet struct {
		GUID          string                `json:"guid,omitempty"`
		Buildpacks    []DropletBuildpack    `json:"buildpacks,omitempty"`
		CreatedAt     string                `json:"created_at,omitempty"`
		Image         string                `json:"image,omitempty"`
		Stack         string                `json:"stack,omitempty"`
		State         constant.DropletState `json:"state,omitempty"`
		Relationships *Relationships        `json:"relationships,omitempty"`
	}

	ccD := ccDroplet{
		GUID:       d.GUID,
		Buildpacks: d.Buildpacks,
		CreatedAt:  d.CreatedAt,
		Image:      d.Image,
		Stack:      d.Stack,
		State:      d.State,
	}

	if d.AppGUID != "" {
		ccD.Relationships = &Relationships{
			App: RelationshipData{Data{GUID: d.AppGUID}},
		}
	}

	return json.Marshal(ccD)
}

func (d *Droplet) UnmarshalJSON(data []byte) error {
	var alias struct {
		GUID          string                `json:"guid,omitempty"`
		Buildpacks    []DropletBuildpack    `json:"buildpacks,omitempty"`
		CreatedAt     string                `json:"created_at,omitempty"`
		Image         string                `json:"image,omitempty"`
		Stack         string                `json:"stack,omitempty"`
		State         constant.DropletState `json:"state,omitempty"`
		Relationships struct {
			App struct {
				Data struct {
					GUID string `json:"guid,omitempty"`
				} `json:"data,omitempty"`
			} `json:"app,omitempty"`
		}
	}

	err := cloudcontroller.DecodeJSON(data, &alias)
	if err != nil {
		return err
	}

	d.GUID = alias.GUID
	d.Buildpacks = alias.Buildpacks
	d.CreatedAt = alias.CreatedAt
	d.Image = alias.Image
	d.Stack = alias.Stack
	d.State = alias.State
	d.AppGUID = alias.Relationships.App.Data.GUID

	return nil
}
