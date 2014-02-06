package cf

import (
	"cf/formatters"
	"errors"
	"fmt"
	"github.com/codegangsta/cli"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
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
	QuotaDefinition QuotaFields
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
	BuildpackUrl     string
	Command          string
	DiskQuota        uint64 // in Megabytes
	EnvironmentVars  map[string]string
	InstanceCount    int
	Memory           uint64 // in Megabytes
	RunningInstances int
	State            string
	SpaceGuid        string
}

type ApplicationSet []Application
type Application struct {
	ApplicationFields
	Stack  Stack
	Routes []RouteSummary
}

func (model Application) ToParams() (params AppParams) {
	state := strings.ToUpper(model.State)
	params = AppParams{
		Guid:            &model.Guid,
		Name:            &model.Name,
		BuildpackUrl:    &model.BuildpackUrl,
		Command:         &model.Command,
		DiskQuota:       &model.DiskQuota,
		InstanceCount:   &model.InstanceCount,
		Memory:          &model.Memory,
		State:           &state,
		StackGuid:       &model.Stack.Guid,
		SpaceGuid:       &model.SpaceGuid,
		EnvironmentVars: &model.EnvironmentVars,
	}

	return
}

type AppSummary struct {
	ApplicationFields
	RouteSummaries []RouteSummary
}

type AppParams struct {
	BuildpackUrl       *string
	Command            *string
	DiskQuota          *uint64
	Domain             *string
	EnvironmentVars    *map[string]string
	Guid               *string
	HealthCheckTimeout *int
	Host               *string
	InstanceCount      *int
	Memory             *uint64
	Name               *string
	NoRoute            *bool
	Path               *string
	RunningInstances   *int
	Services           *[]string
	SpaceGuid          *string
	StackGuid          *string
	StackName          *string
	State              *string
}

func (app *AppParams) Merge(other *AppParams) {
	if other.BuildpackUrl != nil {
		app.BuildpackUrl = other.BuildpackUrl
	}
	if other.Command != nil {
		app.Command = other.Command
	}
	if other.DiskQuota != nil {
		app.DiskQuota = other.DiskQuota
	}
	if other.Domain != nil {
		app.Domain = other.Domain
	}
	if other.EnvironmentVars != nil {
		app.EnvironmentVars = other.EnvironmentVars
	}
	if other.Guid != nil {
		app.Guid = other.Guid
	}
	if other.HealthCheckTimeout != nil {
		app.HealthCheckTimeout = other.HealthCheckTimeout
	}
	if other.Host != nil {
		app.Host = other.Host
	}
	if other.InstanceCount != nil {
		app.InstanceCount = other.InstanceCount
	}
	if other.Memory != nil {
		app.Memory = other.Memory
	}
	if other.Name != nil {
		app.Name = other.Name
	}
	if other.NoRoute != nil {
		app.NoRoute = other.NoRoute
	}
	if other.Path != nil {
		app.Path = other.Path
	}
	if other.RunningInstances != nil {
		app.RunningInstances = other.RunningInstances
	}
	if other.Services != nil {
		app.Services = other.Services
	}
	if other.SpaceGuid != nil {
		app.SpaceGuid = other.SpaceGuid
	}
	if other.StackGuid != nil {
		app.StackGuid = other.StackGuid
	}
	if other.StackName != nil {
		app.StackName = other.StackName
	}
	if other.State != nil {
		app.State = other.State
	}
}

func NewEmptyAppParams() AppParams {
	return AppParams{}
}

func NewAppParamsFromContext(c *cli.Context) (appParams AppParams, err error) {
	appParams = AppParams{}

	if len(c.Args()) > 0 {
		appParams.Name = &c.Args()[0]
	}

	if c.String("b") != "" {
		buildpack := c.String("b")
		appParams.BuildpackUrl = &buildpack
	}

	if c.String("m") != "" {
		var memory uint64
		memory, err = formatters.ToMegabytes(c.String("m"))
		if err != nil {
			err = errors.New(fmt.Sprintf("Invalid memory param: %s\n%s", c.String("m"), err))
			return
		}
		appParams.Memory = &memory
	}

	if c.String("c") != "" {
		command := c.String("c")
		appParams.Command = &command
	}

	if c.String("c") == "null" {
		emptyStr := ""
		appParams.Command = &emptyStr
	}

	if c.String("i") != "" {
		var instances int
		instances, err = strconv.Atoi(c.String("i"))
		if err != nil {
			err = errors.New(fmt.Sprintf("Invalid instances param: %s\n%s", c.String("i"), err))
			return
		}
		appParams.InstanceCount = &instances
	}

	if c.String("s") != "" {
		stackName := c.String("s")
		appParams.StackName = &stackName
	}

	if c.String("t") != "" {
		var timeout int
		timeout, err = strconv.Atoi(c.String("t"))
		if err != nil {
			err = errors.New(fmt.Sprintf("Invalid timeout param: %s\n%s", c.String("t"), err))
			return
		}

		appParams.HealthCheckTimeout = &timeout
	}

	if c.String("p") != "" {
		var path string
		path, err = filepath.Abs(c.String("p"))
		if err != nil {
			err = errors.New(fmt.Sprintf("Error finding app path: %s", err))
			return
		}
		appParams.Path = &path
	}
	return
}

func (app *AppParams) Equals(otherParams *AppParams) bool {
	return reflect.DeepEqual(*app, *otherParams)
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
}

type EventFields struct {
	BasicFields
	Timestamp   time.Time
	Description string
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

type ServiceOfferings []ServiceOffering
type ServiceOffering struct {
	ServiceOfferingFields
	Plans []ServicePlanFields
}

func (s ServiceOfferings) Len() int {
	return len(s)
}

func (s ServiceOfferings) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ServiceOfferings) Less(i, j int) bool {
	return s[i].Label < s[j].Label
}

type ServiceInstanceFields struct {
	BasicFields
	SysLogDrainUrl   string
	ApplicationNames []string
	Params           map[string]string
}

type ServiceInstanceSet []ServiceInstance
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

func NewQuotaFields(name string, memory uint64) (q QuotaFields) {
	q.Name = name
	q.MemoryLimit = memory
	return
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
	Enabled  *bool
	Key      string
	Filename string
	Locked   *bool
}
