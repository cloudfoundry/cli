package manifest

import (
	"errors"
	"fmt"

	"code.cloudfoundry.org/cli/cf/models"

	"gopkg.in/yaml.v2"

	"io"

	. "code.cloudfoundry.org/cli/cf/i18n"
)

//go:generate counterfeiter . App

type App interface {
	BuildpackURL(string, string)
	DiskQuota(string, int64)
	Memory(string, int64)
	Service(string, string)
	StartCommand(string, string)
	EnvironmentVars(string, string, string)
	HealthCheckTimeout(string, int)
	Instances(string, int)
	Route(string, string, string, string, int)
	GetContents() []models.Application
	Stack(string, string)
	AppPorts(string, []int)
	Save(f io.Writer) error
}

type Application struct {
	Name      string                 `yaml:"name"`
	Instances int                    `yaml:"instances,omitempty"`
	Memory    string                 `yaml:"memory,omitempty"`
	DiskQuota string                 `yaml:"disk_quota,omitempty"`
	AppPorts  []int                  `yaml:"app-ports,omitempty"`
	Routes    []map[string]string    `yaml:"routes,omitempty"`
	NoRoute   bool                   `yaml:"no-route,omitempty"`
	Buildpack string                 `yaml:"buildpack,omitempty"`
	Command   string                 `yaml:"command,omitempty"`
	Env       map[string]interface{} `yaml:"env,omitempty"`
	Services  []string               `yaml:"services,omitempty"`
	Stack     string                 `yaml:"stack,omitempty"`
	Timeout   int                    `yaml:"timeout,omitempty"`
}

type Applications struct {
	Applications []Application `yaml:"applications"`
}

type appManifest struct {
	contents []models.Application
}

func NewGenerator() App {
	return &appManifest{}
}

func (m *appManifest) Stack(appName string, stackName string) {
	i := m.findOrCreateApplication(appName)
	m.contents[i].Stack = &models.Stack{
		Name: stackName,
	}
}

func (m *appManifest) Memory(appName string, memory int64) {
	i := m.findOrCreateApplication(appName)
	m.contents[i].Memory = memory
}

func (m *appManifest) DiskQuota(appName string, diskQuota int64) {
	i := m.findOrCreateApplication(appName)
	m.contents[i].DiskQuota = diskQuota
}

func (m *appManifest) StartCommand(appName string, cmd string) {
	i := m.findOrCreateApplication(appName)
	m.contents[i].Command = cmd
}

func (m *appManifest) BuildpackURL(appName string, url string) {
	i := m.findOrCreateApplication(appName)
	m.contents[i].BuildpackURL = url
}

func (m *appManifest) HealthCheckTimeout(appName string, timeout int) {
	i := m.findOrCreateApplication(appName)
	m.contents[i].HealthCheckTimeout = timeout
}

func (m *appManifest) Instances(appName string, instances int) {
	i := m.findOrCreateApplication(appName)
	m.contents[i].InstanceCount = instances
}

func (m *appManifest) Service(appName string, name string) {
	i := m.findOrCreateApplication(appName)
	m.contents[i].Services = append(m.contents[i].Services, models.ServicePlanSummary{
		GUID: "",
		Name: name,
	})
}

func (m *appManifest) Route(appName, host, domain, path string, port int) {
	i := m.findOrCreateApplication(appName)
	m.contents[i].Routes = append(m.contents[i].Routes, models.RouteSummary{
		Host: host,
		Domain: models.DomainFields{
			Name: domain,
		},
		Path: path,
		Port: port,
	})

}

func (m *appManifest) EnvironmentVars(appName string, key, value string) {
	i := m.findOrCreateApplication(appName)
	m.contents[i].EnvironmentVars[key] = value
}

func (m *appManifest) AppPorts(appName string, appPorts []int) {
	i := m.findOrCreateApplication(appName)
	m.contents[i].AppPorts = appPorts
}

func (m *appManifest) GetContents() []models.Application {
	return m.contents
}

func generateAppMap(app models.Application) (Application, error) {
	if app.Stack == nil {
		return Application{}, errors.New(T("required attribute 'stack' missing"))
	}

	if app.Memory == 0 {
		return Application{}, errors.New(T("required attribute 'memory' missing"))
	}

	if app.DiskQuota == 0 {
		return Application{}, errors.New(T("required attribute 'disk_quota' missing"))
	}

	if app.InstanceCount == 0 {
		return Application{}, errors.New(T("required attribute 'instances' missing"))
	}

	var services []string
	for _, s := range app.Services {
		services = append(services, s.Name)
	}

	var routes []map[string]string
	for _, routeSummary := range app.Routes {
		routes = append(routes, buildRoute(routeSummary))
	}
	m := Application{
		Name:      app.Name,
		Services:  services,
		Buildpack: app.BuildpackURL,
		Memory:    fmt.Sprintf("%dM", app.Memory),
		Command:   app.Command,
		Env:       app.EnvironmentVars,
		Timeout:   app.HealthCheckTimeout,
		Instances: app.InstanceCount,
		DiskQuota: fmt.Sprintf("%dM", app.DiskQuota),
		Stack:     app.Stack.Name,
		AppPorts:  app.AppPorts,
		Routes:    routes,
	}

	if len(app.Routes) == 0 {
		m.NoRoute = true

	}

	return m, nil
}

func (m *appManifest) Save(f io.Writer) error {
	apps := Applications{}

	for _, app := range m.contents {
		appMap, mapErr := generateAppMap(app)
		if mapErr != nil {
			return fmt.Errorf(T("Error saving manifest: {{.Error}}", map[string]interface{}{
				"Error": mapErr.Error(),
			}))
		}
		apps.Applications = append(apps.Applications, appMap)
	}

	contents, err := yaml.Marshal(apps)
	if err != nil {
		return err
	}

	_, err = f.Write(contents)
	if err != nil {
		return err
	}

	return nil
}

func buildRoute(routeSummary models.RouteSummary) map[string]string {
	var route string
	if routeSummary.Host != "" {
		route = fmt.Sprintf("%s.", routeSummary.Host)
	}

	route = fmt.Sprintf("%s%s", route, routeSummary.Domain.Name)

	if routeSummary.Path != "" {
		route = fmt.Sprintf("%s%s", route, routeSummary.Path)
	}

	if routeSummary.Port != 0 {
		route = fmt.Sprintf("%s:%d", route, routeSummary.Port)
	}

	return map[string]string{
		"route": route,
	}
}

func (m *appManifest) findOrCreateApplication(name string) int {
	for i, app := range m.contents {
		if app.Name == name {
			return i
		}
	}
	m.addApplication(name)
	return len(m.contents) - 1
}

func (m *appManifest) addApplication(name string) {
	m.contents = append(m.contents, models.Application{
		ApplicationFields: models.ApplicationFields{
			Name:            name,
			EnvironmentVars: make(map[string]interface{}),
		},
	})
}
