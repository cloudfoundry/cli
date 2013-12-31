package cf

import (
	"cf/formatters"
	"errors"
	"fmt"
	"generic"
	"github.com/codegangsta/cli"
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

type Application struct {
	ApplicationFields
	Stack  Stack
	Routes []RouteSummary
}

func (model Application) ToParams() (params AppParams) {
	params = NewAppParams(map[interface{}]interface{}{
		"guid":       model.Guid,
		"name":       model.Name,
		"buildpack":  model.BuildpackUrl,
		"command":    model.Command,
		"disk_quota": model.DiskQuota,
		"instances":  model.InstanceCount,
		"memory":     model.Memory,
		"state":      strings.ToUpper(model.State),
		"stack_guid": model.Stack.Guid,
		"space_guid": model.SpaceGuid,
		"env":        generic.NewMap(model.EnvironmentVars),
	})

	return
}

type AppSummary struct {
	ApplicationFields
	RouteSummaries []RouteSummary
}

type AppParams struct {
	generic.Map
}

func (app AppParams) ToMap() generic.Map {
	return app.Map
}

func NewEmptyAppParams() AppParams {
	return NewAppParams(generic.NewEmptyMap())
}

func NewAppParams(data interface{}) (params AppParams) {
	params = AppParams{}

	switch data := data.(type) {
	case map[interface{}]interface{}:
		params.Map = generic.NewMap(data)
	case generic.Map:
		params.Map = data
	default:
		panic(fmt.Sprintf("AppParams initialized with unexpected type: %T", data))
	}

	if params.Map.NotNil("memory") {
		memory, err := formatters.ToMegabytes(params.Map.Get("memory"))
		if err != nil {
			panic(err)
		}
		params.Map.Set("memory", memory)
	}
	return
}

func NewAppParamsFromContext(c *cli.Context) (appParams AppParams, err error) {
	appParams = NewEmptyAppParams()

	if len(c.Args()) > 0 {
		appParams.Set("name", c.Args()[0])
	}

	if c.String("b") != "" {
		appParams.Set("buildpack", c.String("b"))
	}
	if c.String("m") != "" {
		var memory uint64
		memory, err = formatters.ToMegabytes(c.String("m"))
		if err != nil {
			err = errors.New(fmt.Sprintf("Invalid memory param: %s\n%s", c.String("m"), err))
			return
		}
		appParams.Set("memory", memory)
	}
	if c.String("c") != "" {
		appParams.Set("command", c.String("c"))
	}
	if c.String("i") != "" {
		var instances int
		instances, err = strconv.Atoi(c.String("i"))
		if err != nil {
			err = errors.New(fmt.Sprintf("Invalid instances param: %s\n%s", c.String("i"), err))
			return
		}
		appParams.Set("instances", instances)
	}
	if c.String("s") != "" {
		appParams.Set("stack", c.String("s"))
	}
	if c.String("t") != "" {
		var timeout int
		timeout, err = strconv.Atoi(c.String("t"))
		if err != nil {
			err = errors.New(fmt.Sprintf("Invalid timeout param: %s\n%s", c.String("t"), err))
			return
		}

		appParams.Set("health_check_timeout", timeout)
	}

	return
}

type AppSet []AppParams

func NewEmptyAppSet() AppSet {
	return NewAppSet(make([]interface{}, 0))
}

func NewAppSet(values interface{}) (set AppSet) {
	set = AppSet{}

	switch values := values.(type) {
	case []interface{}:
		for _, val := range values {
			app := generic.NewMap(val)
			set = append(set, NewAppParams(app))
		}
	default:
		app := generic.NewMap(values)
		set = append(set, NewAppParams(app))
	}

	return
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
}
