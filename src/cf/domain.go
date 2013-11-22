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

type BasicFields struct {
	Guid string
	Name string
}

func (model BasicFields) String() string {
	return model.Name
}

type OrganizationFields struct {
	BasicFields
}

type Organization struct {
	OrganizationFields
	Spaces  []SpaceFields
	Domains []DomainFields
}

type SpaceFields struct {
	BasicFields
}

type Space struct {
	SpaceFields
	Organization     OrganizationFields
	Applications     []ApplicationFields
	ServiceInstances []ServiceInstanceFields
	Domains          []DomainFields
}

type ApplicationFields struct {
	BasicFields
	State            string
	Command          string
	BuildpackUrl     string
	InstanceCount    int
	RunningInstances int
	Memory           uint64 // in Megabytes
	DiskQuota        uint64 // in Megabytes
	EnvironmentVars  map[string]string
}

type Application struct {
	ApplicationFields
	Stack  Stack
	Routes []RouteSummary
}

type AppSummary struct {
	ApplicationFields
	RouteSummaries []RouteSummary
}

type AppFileFields struct {
	Path string
	Sha1 string
	Size int64
}

type DomainFields struct {
	BasicFields
	OwningOrganizationGuid string
	Shared                 bool
}

func (model DomainFields) UrlForHost(host string) string {
	if host == "" {
		return model.Name
	}
	return fmt.Sprintf("%s.%s", host, model.Name)
}

type Domain struct {
	DomainFields
	Spaces []SpaceFields
}

type EventFields struct {
	InstanceIndex   int
	Timestamp       time.Time
	ExitDescription string
	ExitStatus      int
}

type RouteFields struct {
	Guid string
	Host string
}

type Route struct {
	RouteSummary
	Space SpaceFields
	Apps  []ApplicationFields
}

type RouteSummary struct {
	RouteFields
	Domain DomainFields
}

func (model RouteSummary) URL() string {
	if model.Host == "" {
		return model.Domain.Name
	}
	return fmt.Sprintf("%s.%s", model.Host, model.Domain.Name)
}

type Stack struct {
	BasicFields
	Description string
}

type AppInstanceFields struct {
	State     InstanceState
	Since     time.Time
	CpuUsage  float64 // percentage
	DiskQuota uint64  // in bytes
	DiskUsage uint64
	MemQuota  uint64
	MemUsage  uint64
}

type ServicePlanFields struct {
	BasicFields
}

type ServicePlan struct {
	ServicePlanFields
	ServiceOffering ServiceOfferingFields
}

type ServiceOfferingFields struct {
	Guid             string
	Label            string
	Provider         string
	Version          string
	Description      string
	DocumentationUrl string
}

type ServiceOffering struct {
	ServiceOfferingFields
	Plans []ServicePlanFields
}

type ServiceInstanceFields struct {
	BasicFields
	SysLogDrainUrl   string
	ApplicationNames []string
	Params           map[string]string
}

type ServiceInstance struct {
	ServiceInstanceFields
	ServiceBindings []ServiceBindingFields
	ServicePlan     ServicePlanFields
	ServiceOffering ServiceOfferingFields
}

func (inst ServiceInstance) IsUserProvided() bool {
	return inst.ServicePlan.Guid == ""
}

type ServiceBindingFields struct {
	Guid    string
	Url     string
	AppGuid string
}

type QuotaFields struct {
	BasicFields
	MemoryLimit uint64 // in Megabytes
}

type ServiceAuthTokenFields struct {
	Guid     string
	Label    string
	Provider string
	Token    string
}

type ServiceBroker struct {
	BasicFields
	Username string
	Password string
	Url      string
}

type UserFields struct {
	Guid     string
	Username string
	Password string
	IsAdmin  bool
}

type Buildpack struct {
	BasicFields
	Position *int
}
