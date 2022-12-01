package resources

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
)

// Droplet represents a Cloud Controller droplet's metadata. A droplet is a set of
// compiled bits for a given application.
type Droplet struct {
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
