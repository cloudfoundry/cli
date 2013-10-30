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

func (space Space) String() string {
	return space.Name
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
	Command          string
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
	Name   string
	Guid   string
	Shared bool
	Spaces []Space
}

type Event struct {
	InstanceIndex   int
	Timestamp       time.Time
	ExitDescription string
	ExitStatus      int
}

type Route struct {
	Host     string
	Guid     string
	Domain   Domain
	Space    Space
	AppNames []string
}

func (r Route) URL() string {
	if r.Host == "" {
		return r.Domain.Name
	}
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
	Params           map[string]string
	SysLogDrainUrl   string
}

func (inst ServiceInstance) IsUserProvided() bool {
	return inst.ServicePlan.Guid == ""
}

func (inst ServiceInstance) ServiceOffering() ServiceOffering {
	return inst.ServicePlan.ServiceOffering
}

type ServiceBinding struct {
	Guid    string
	Url     string
	AppGuid string
}

type Quota struct {
	Guid        string
	Name        string
	MemoryLimit uint64 // in Megabytes
}

type ServiceAuthToken struct {
	Guid     string
	Label    string
	Provider string
	Token    string
}

type ServiceBroker struct {
	Guid     string
	Name     string
	Username string
	Password string
	Url      string
}

type User struct {
	Guid     string
	Username string
	Password string
	IsAdmin  bool
}

type Buildpack struct {
	Guid     string
	Name     string
	Position *int
}
