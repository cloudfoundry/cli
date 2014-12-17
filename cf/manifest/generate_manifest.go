package manifest

import (
	"fmt"
	"os"

	"github.com/cloudfoundry/cli/cf/models"
)

type AppManifest interface {
	Memory(string, int64)
	Service(string, string)
	StartupCommand(string, string)
	EnvironmentVars(string, string, string)
	HealthCheckTimeout(string, int)
	Instances(string, int)
	Domain(string, string, string)
	GetContents() []models.Application
	FileSavePath(string)
	Save() error
}

type appManifest struct {
	savePath string
	contents []models.Application
}

func NewGenerator() AppManifest {
	return &appManifest{}
}

func (m *appManifest) FileSavePath(savePath string) {
	m.savePath = savePath
}

func (m *appManifest) Memory(appName string, memory int64) {
	i := m.findOrCreateApplication(appName)
	m.contents[i].Memory = memory
}

func (m *appManifest) StartupCommand(appName string, cmd string) {
	i := m.findOrCreateApplication(appName)
	m.contents[i].Command = cmd
}

func (m *appManifest) BuildpackUrl(appName string, url string) {
	i := m.findOrCreateApplication(appName)
	m.contents[i].BuildpackUrl = url
}

func (m *appManifest) StartCommand(appName string, cmd string) {
	i := m.findOrCreateApplication(appName)
	m.contents[i].Command = cmd
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

func (m *appManifest) GetContents() []models.Application {
	return m.contents
}

func (m *appManifest) Save() error {
	f, err := os.Create(m.savePath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = fmt.Fprintln(f, "---\napplications:")

	for _, app := range m.contents {
		if _, err := fmt.Fprintf(f, "- name: %s\n", app.Name); err != nil {
			return err
		}

		if _, err := fmt.Fprintf(f, "  memory: %dM\n", app.Memory); err != nil {
			return err
		}

		if _, err := fmt.Fprintf(f, "  instances: %d\n", app.InstanceCount); err != nil {
			return err
		}

		if app.BuildpackUrl != "" {
			if _, err := fmt.Fprintf(f, "  buildpack: %s\n", app.BuildpackUrl); err != nil {
				return err
			}
		}

		if app.HealthCheckTimeout > 0 {
			if _, err := fmt.Fprintf(f, "  timeout: %d\n", app.HealthCheckTimeout); err != nil {
				return err
			}
		}

		if app.Command != "" {
			if _, err := fmt.Fprintf(f, "  command: %s\n", app.Command); err != nil {
				return err
			}
		}

		if len(app.Routes) > 0 {
			if _, err := fmt.Fprintf(f, "  host: %s\n", app.Routes[0].Host); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(f, "  domain: %s\n", app.Routes[0].Domain.Name); err != nil {
				return err
			}
		} else {
			if _, err := fmt.Fprintf(f, "  no-route: true\n"); err != nil {
				return err
			}
		}

		if len(app.Services) > 0 {
			if err := writeServicesToFile(f, app.Services); err != nil {
				return err
			}
		}

		if len(app.EnvironmentVars) > 0 {
			if err := writeEnvironmentVarToFile(f, app.EnvironmentVars); err != nil {
				return err
			}
		}

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

func writeServicesToFile(f *os.File, entries []models.ServicePlanSummary) error {
	_, err := fmt.Fprintln(f, "  services:")
	if err != nil {
		return err
	}
	for _, service := range entries {
		_, err = fmt.Fprintf(f, "  - %s\n", service.Name)
		if err != nil {
			return err
		}
	}

	return nil
}

func writeEnvironmentVarToFile(f *os.File, envVars map[string]interface{}) error {
	_, err := fmt.Fprintln(f, "  env:")
	if err != nil {
		return err
	}
	for k, v := range envVars {
		_, err = fmt.Fprintf(f, "    %s: %s\n", k, v)
		if err != nil {
			return err
		}
	}

	return nil
}
