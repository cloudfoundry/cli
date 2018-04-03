package manifest

import (
	biproperty "github.com/cloudfoundry/bosh-utils/property"
)

type Job struct {
	Name               string
	Instances          int
	Lifecycle          JobLifecycle
	Templates          []ReleaseJobRef
	Networks           []JobNetwork
	PersistentDisk     int
	PersistentDiskPool string
	ResourcePool       string
	Properties         biproperty.Map
}

type JobLifecycle string

const (
	JobLifecycleService JobLifecycle = "service"
	JobLifecycleErrand  JobLifecycle = "errand"
)

type ReleaseJobRef struct {
	Name       string
	Release    string
	Properties *biproperty.Map
}

type JobNetwork struct {
	Name      string
	Defaults  []NetworkDefault
	StaticIPs []string
}

type NetworkDefault string

const (
	NetworkDefaultDNS     NetworkDefault = "dns"
	NetworkDefaultGateway NetworkDefault = "gateway"
)
