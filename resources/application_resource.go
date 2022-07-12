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
}

// MarshalJSON converts an Application into a Cloud Controller Application.
func (a Application) MarshalJSON() ([]byte, error) {
	ccApp := ccApplication{
		Name:     a.Name,
		Metadata: a.Metadata,
	}

	if a.SpaceGUID != "" {
		ccApp.Relationships = Relationships{
			constant.RelationshipTypeSpace: Relationship{GUID: a.SpaceGUID},
		}
	}

	if a.LifecycleType == constant.AppLifecycleTypeDocker {
		ccApp.setDockerLifecycle()
	} else if a.LifecycleType == constant.AppLifecycleTypeBuildpack {
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

type ccLifecycle struct {
	Type constant.AppLifecycleType `json:"type,omitempty"`
	Data struct {
		Buildpacks []string `json:"buildpacks,omitempty"`
		Stack      string   `json:"stack,omitempty"`
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
	ccApp.Lifecycle = lifecycle
}

func (ccApp *ccApplication) setDockerLifecycle() {
	ccApp.Lifecycle = ccLifecycle{
		Type: constant.AppLifecycleTypeDocker,
	}
}
