// +build V7

package models

import (
	"time"

	"code.cloudfoundry.org/cli/types"
)

type Metadata struct {
	Labels map[string]types.NullString
}

type AppLifecycleType string
type ApplicationState string

type Application struct {
	Name                string
	GUID                string
	StackName           string
	State               ApplicationState
	LifecycleType       AppLifecycleType
	LifecycleBuildpacks []string
	Metadata            *Metadata
}

type HealthCheckType string

type Process struct {
	GUID                         string
	Type                         string
	Command                      types.FilteredString
	HealthCheckType              HealthCheckType
	HealthCheckEndpoint          string
	HealthCheckInvocationTimeout int64
	HealthCheckTimeout           int64
	Instances                    types.NullInt
	MemoryInMB                   types.NullUint64
	DiskInMB                     types.NullUint64
}

type Sidecar struct {
	GUID    string               `json:"guid"`
	Name    string               `json:"name"`
	Command types.FilteredString `json:"command"`
}

type ProcessInstanceState string

type ProcessInstance struct {
	CPU              float64
	Details          string
	DiskQuota        uint64
	DiskUsage        uint64
	Index            int64
	IsolationSegment string
	MemoryQuota      uint64
	MemoryUsage      uint64
	State            ProcessInstanceState
	Type             string
	Uptime           time.Duration
}

type ProcessSummary struct {
	Process

	Sidecars []Sidecar

	InstanceDetails []ProcessInstance
}

type ProcessSummaries []ProcessSummary

type Route struct {
	GUID       string
	SpaceGUID  string
	DomainGUID string
	Host       string
	Path       string
	DomainName string
	SpaceName  string
	URL        string
}

type ApplicationSummary struct {
	Application
	ProcessSummaries ProcessSummaries
	Routes           []Route
}

type DropletState string

type DropletBuildpack struct {
	Name         string
	DetectOutput string `json:"detect_output"`
}

type Droplet struct {
	GUID       string
	State      DropletState
	CreatedAt  string
	Stack      string
	Image      string
	Buildpacks []DropletBuildpack
}

type DetailedApplicationSummary struct {
	ApplicationSummary
	CurrentDroplet Droplet
}
