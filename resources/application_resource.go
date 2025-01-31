package resources

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
)

// Application represents a Cloud Controller V3 Application.
type Application struct {
	// GUID is the unique application identifier.
	GUID string
	// StackName is the name of the stack on which the application runs.
	StackName string
	// LifecycleBuildpacks is a list of the names of buildpacks.
	LifecycleBuildpacks []string
	// LifecycleType is the type of the lifecycle.
	LifecycleType constant.AppLifecycleType
	// Metadata is used for custom tagging of API resources
	Metadata *Metadata
	// Name is the name given to the application.
	Name string
	// SpaceGUID is the unique identifier of the parent space.
	SpaceGUID string
	// State is the desired state of the application.
	State constant.ApplicationState
	// Credentials are used by Cloud Native Buildpacks lifecycle to pull buildpacks
	Credentials map[string]interface{}
	// CurrentDropletGUID is the unique identifier of the droplet currently attached to the application.
	CurrentDropletGUID string
}

// ApplicationNameOnly represents only the name field of a Cloud Controller V3 Application
type ApplicationNameOnly struct {
	// Name is the name given to the application.
	Name string `json:"name,omitempty"`
}

// MarshalJSON converts an Application into a Cloud Controller Application.
func (a Application) MarshalJSON() ([]byte, error) {
	ccApp := ccApplication{
		Name:     a.Name,
		Metadata: a.Metadata,
	}

	ccApp.Relationships = Relationships{}

	if a.SpaceGUID != "" {
		ccApp.Relationships[constant.RelationshipTypeSpace] = Relationship{GUID: a.SpaceGUID}
	}

	if a.CurrentDropletGUID != "" {
		ccApp.Relationships[constant.RelationshipTypeCurrentDroplet] = Relationship{GUID: a.CurrentDropletGUID}
	}

	if a.LifecycleType == constant.AppLifecycleTypeDocker {
		ccApp.setDockerLifecycle()
	} else if a.LifecycleType == constant.AppLifecycleTypeBuildpack || a.LifecycleType == constant.AppLifecycleTypeCNB {
		if len(a.LifecycleBuildpacks) > 0 || a.StackName != "" {
			if a.hasAutodetectedBuildpack() {
				ccApp.setAutodetectedBuildpackLifecycle(a)
			} else {
				ccApp.setBuildpackLifecycle(a)
			}
		}
	}

	return json.Marshal(ccApp)
}

// UnmarshalJSON helps unmarshal a Cloud Controller Application response.
func (a *Application) UnmarshalJSON(data []byte) error {
	lifecycle := ccLifecycle{}
	ccApp := ccApplication{
		Lifecycle: &lifecycle,
	}

	err := cloudcontroller.DecodeJSON(data, &ccApp)
	if err != nil {
		return err
	}

	a.GUID = ccApp.GUID
	a.StackName = lifecycle.Data.Stack
	a.LifecycleBuildpacks = lifecycle.Data.Buildpacks
	a.LifecycleType = lifecycle.Type
	a.Name = ccApp.Name
	a.SpaceGUID = ccApp.Relationships[constant.RelationshipTypeSpace].GUID
	if _, ok := ccApp.Relationships[constant.RelationshipTypeCurrentDroplet]; ok {
		a.CurrentDropletGUID = ccApp.Relationships[constant.RelationshipTypeCurrentDroplet].GUID
	}
	a.State = ccApp.State
	a.Metadata = ccApp.Metadata

	return nil
}

func (a Application) Started() bool {
	return a.State == constant.ApplicationStarted
}

func (a Application) Stopped() bool {
	return a.State == constant.ApplicationStopped
}

func (a Application) hasAutodetectedBuildpack() bool {
	if len(a.LifecycleBuildpacks) == 0 {
		return false
	}
	return a.LifecycleBuildpacks[0] == constant.AutodetectBuildpackValueDefault || a.LifecycleBuildpacks[0] == constant.AutodetectBuildpackValueNull
}

type ccCredentials map[string]interface{}

func (ccCredentials) UnmarshalJSON(data []byte) error {
	return nil
}

type ccLifecycle struct {
	Type constant.AppLifecycleType `json:"type,omitempty"`
	Data struct {
		Buildpacks  []string      `json:"buildpacks,omitempty"`
		Stack       string        `json:"stack,omitempty"`
		Credentials ccCredentials `json:"credentials,omitempty"`
	} `json:"data"`
}

type ccApplication struct {
	Name          string                    `json:"name,omitempty"`
	Relationships Relationships             `json:"relationships,omitempty"`
	Lifecycle     interface{}               `json:"lifecycle,omitempty"`
	GUID          string                    `json:"guid,omitempty"`
	State         constant.ApplicationState `json:"state,omitempty"`
	Metadata      *Metadata                 `json:"metadata,omitempty"`
}

func (ccApp *ccApplication) setAutodetectedBuildpackLifecycle(a Application) {
	var nullBuildpackLifecycle struct {
		Type constant.AppLifecycleType `json:"type,omitempty"`
		Data struct {
			Buildpacks []string `json:"buildpacks"`
			Stack      string   `json:"stack,omitempty"`
		} `json:"data"`
	}
	nullBuildpackLifecycle.Type = constant.AppLifecycleTypeBuildpack
	nullBuildpackLifecycle.Data.Stack = a.StackName
	ccApp.Lifecycle = nullBuildpackLifecycle
}

func (ccApp *ccApplication) setBuildpackLifecycle(a Application) {
	var lifecycle ccLifecycle
	lifecycle.Type = a.LifecycleType
	lifecycle.Data.Buildpacks = a.LifecycleBuildpacks
	lifecycle.Data.Stack = a.StackName
	lifecycle.Data.Credentials = a.Credentials
	ccApp.Lifecycle = lifecycle
}

func (ccApp *ccApplication) setDockerLifecycle() {
	ccApp.Lifecycle = ccLifecycle{
		Type: constant.AppLifecycleTypeDocker,
	}
}
