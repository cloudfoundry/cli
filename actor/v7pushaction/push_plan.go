package v7pushaction

import (
	"fmt"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/types"
	"github.com/cloudfoundry/bosh-cli/director/template"
)

type PushPlan struct {
	SpaceGUID string
	OrgGUID   string

	Application v7action.Application

	NoStart  bool
	NoWait   bool
	Strategy constant.DeploymentStrategy

	DockerImageCredentials v7action.DockerImageCredentials

	Archive      bool
	BitsPath     string
	DropletPath  string
	AllResources []sharedaction.V3Resource

	PackageGUID string
	DropletGUID string
}

type FlagOverrides struct {
	AppName             string
	Buildpacks          []string
	Stack               string
	Disk                string
	DropletPath         string
	DockerImage         string
	DockerPassword      string
	DockerUsername      string
	HealthCheckEndpoint string
	HealthCheckTimeout  int64
	HealthCheckType     constant.HealthCheckType
	Instances           types.NullInt
	Memory              string
	NoStart             bool
	NoWait              bool
	ProvidedAppPath     string
	NoRoute             bool
	RandomRoute         bool
	StartCommand        types.FilteredString
	Strategy            constant.DeploymentStrategy
	ManifestPath        string
	PathsToVarsFiles    []string
	Vars                []template.VarKV
	NoManifest          bool
}

func (state PushPlan) String() string {
	return fmt.Sprintf(
		"Application: %#v - Space GUID: %s, Org GUID: %s, Archive: %t, Bits Path: %s",
		state.Application,
		state.SpaceGUID,
		state.OrgGUID,
		state.Archive,
		state.BitsPath,
	)
}
