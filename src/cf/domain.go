package cf

import (
	"fmt"
	"time"
)

type InstanceState string

const (
	InstanceStarting InstanceState = "starting"
	InstanceRunning                = "running"
	InstanceFlapping               = "flapping"
	InstanceDown                   = "down"
)

type Organization struct {
	Name    string
	Guid    string
	Spaces  []Space
	Domains []Domain
}

type Space struct {
	Name             string
	Guid             string
	Applications     []Application
	ServiceInstances []ServiceInstance
	Organization     Organization
	Domains          []Domain
}

type Application struct {
	Name             string
	Guid             string
	State            string
	Instances        int
	RunningInstances int
	Memory           uint64 // in Megabytes
	DiskQuota        uint64 // in Megabytes
	Urls             []string
	BuildpackUrl     string
	Stack            Stack
	EnvironmentVars  map[string]string
}

func (app Application) Health() string {
	if app.State != "started" {
		return app.State
	}

	if app.Instances > 0 {
		ratio := float32(app.RunningInstances) / float32(app.Instances)
		if ratio == 1 {
			return "running"
		}
		return fmt.Sprintf("%.0f%%", ratio*100)
	}

	return "N/A"
}

type AppSummary struct {
	App       Application
	Instances []ApplicationInstance
}

type AppFile struct {
	Path string
	Sha1 string
	Size int64
}

type Domain struct {
	Name string
	Guid string
}

type Route struct {
	Host   string
	Guid   string
	Domain Domain
}

func (r Route) URL() string {
	return fmt.Sprintf("%s.%s", r.Host, r.Domain.Name)
}

type Stack struct {
	Name        string
	Guid        string
	Description string
}

type ApplicationInstance struct {
	State     InstanceState
	Since     time.Time
	CpuUsage  float64 // percentage
	DiskQuota uint64  // in bytes
	DiskUsage uint64
	MemQuota  uint64
	MemUsage  uint64
}

type ServicePlan struct {
	Name            string
	Guid            string
	ServiceOffering ServiceOffering
}

type ServiceOffering struct {
	Guid             string
	Label            string
	Provider         string
	Version          string
	Description      string
	DocumentationUrl string
	Plans            []ServicePlan
}

type ServiceInstance struct {
	Name             string
	Guid             string
	ServiceBindings  []ServiceBinding
	ServicePlan      ServicePlan
	ApplicationNames []string
	ServiceOffering  ServiceOffering
}

type ServiceBinding struct {
	Url     string
	Guid    string
	AppGuid string
}
