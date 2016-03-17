package manifest

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/generic"

	"gopkg.in/yaml.v2"

	"io"

	. "github.com/cloudfoundry/cli/cf/i18n"
)

//go:generate counterfeiter -o fakes/fake_app_manifest.go . AppManifest
type AppManifest interface {
	BuildpackUrl(string, string)
	DiskQuota(string, int64)
	Memory(string, int64)
	Service(string, string)
	StartCommand(string, string)
	EnvironmentVars(string, string, string)
	HealthCheckTimeout(string, int)
	Instances(string, int)
	Domain(string, string, string)
	GetContents() []models.Application
	Stack(string, string)
	AppPorts(string, []int)
	Save(f io.Writer) error
}

type appManifest struct {
	contents []models.Application
}

func NewGenerator() AppManifest {
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

func (m *appManifest) BuildpackUrl(appName string, url string) {
	i := m.findOrCreateApplication(appName)
	m.contents[i].BuildpackUrl = url
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
		Guid: "",
		Name: name,
	})
}

func (m *appManifest) Domain(appName string, host string, domain string) {
	i := m.findOrCreateApplication(appName)
	m.contents[i].Routes = append(m.contents[i].Routes, models.RouteSummary{
		Host: host,
		Domain: models.DomainFields{
			Name: domain,
		},
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

func generateAppMap(app models.Application) (generic.Map, error) {
	if app.Stack == nil {
		return generic.NewMap(), errors.New(T("required attribute 'stack' missing"))
	}

	if app.Memory == 0 {
		return generic.NewMap(), errors.New(T("required attribute 'memory' missing"))
	}

	if app.DiskQuota == 0 {
		return generic.NewMap(), errors.New(T("required attribute 'disk_quota' missing"))
	}

	if app.InstanceCount == 0 {
		return generic.NewMap(), errors.New(T("required attribute 'instances' missing"))
	}

	m := generic.NewMap()

	m.Set("name", app.Name)
	m.Set("memory", fmt.Sprintf("%dM", app.Memory))
	m.Set("instances", app.InstanceCount)
	m.Set("disk_quota", fmt.Sprintf("%dM", app.DiskQuota))
	m.Set("stack", app.Stack.Name)

	if len(app.AppPorts) > 0 {
		m.Set("app-ports", app.AppPorts)
	}

	if app.BuildpackUrl != "" {
		m.Set("buildpack", app.BuildpackUrl)
	}

	if app.HealthCheckTimeout > 0 {
		m.Set("timeout", app.HealthCheckTimeout)
	}

	if app.Command != "" {
		m.Set("command", app.Command)
	}

	switch len(app.Routes) {
	case 0:
		m.Set("no-route", true)
	case 1:
		const noHostname = ""

		m.Set("domain", app.Routes[0].Domain.Name)
		host := app.Routes[0].Host

		if host == noHostname {
			m.Set("no-hostname", true)
		} else {
			m.Set("host", host)
		}
	default:
		hosts, domains := separateHostsAndDomains(app.Routes)

		switch len(hosts) {
		case 0:
			m.Set("no-hostname", true)
		case 1:
			m.Set("host", hosts[0])
		default:
			m.Set("hosts", hosts)
		}

		switch len(domains) {
		case 1:
			m.Set("domain", domains[0])
		default:
			m.Set("domains", domains)
		}
	}

	if len(app.Services) > 0 {
		var services []string

		for _, s := range app.Services {
			services = append(services, s.Name)
		}

		m.Set("services", services)
	}

	if len(app.EnvironmentVars) > 0 {
		m.Set("env", app.EnvironmentVars)
	}

	return m, nil
}

func (m *appManifest) Save(f io.Writer) error {
	y := generic.NewMap()

	apps := []generic.Map{}

	for _, app := range m.contents {
		appMap, mapErr := generateAppMap(app)
		if mapErr != nil {
			return fmt.Errorf(T("Error saving manifest: {{.Error}}", map[string]interface{}{
				"Error": mapErr.Error(),
			}))
		}
		apps = append(apps, appMap)
	}

	y.Set("applications", apps)

	contents, err := yaml.Marshal(y)
	if err != nil {
		return err
	}

	_, err = f.Write(contents)
	if err != nil {
		return err
	}

	return nil
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

func separateHostsAndDomains(routes []models.RouteSummary) ([]string, []string) {
	var (
		hostSlice    []string
		domainSlice  []string
		hostPSlice   *[]string
		domainPSlice *[]string
		hosts        []string
		domains      []string
	)

	for i := 0; i < len(routes); i++ {
		hostSlice = append(hostSlice, routes[i].Host)
		domainSlice = append(domainSlice, routes[i].Domain.Name)
	}

	hostPSlice = removeDuplicatedValue(hostSlice)
	domainPSlice = removeDuplicatedValue(domainSlice)

	if hostPSlice != nil {
		hosts = *hostPSlice
	}
	if domainPSlice != nil {
		domains = *domainPSlice
	}

	return hosts, domains
}
