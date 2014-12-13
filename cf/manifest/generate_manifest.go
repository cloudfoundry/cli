package manifest

import (
	"fmt"
	"os"

	"github.com/cloudfoundry/cli/cf/models"
)

type AppManifest interface {
	Memory(string, int64)
	Service(string, string)
	EnvironmentVars(string, string, string)
	HealthCheckTimeout(string, int)
	Instances(string, int)
	Domain(string, string, string)
	GetContents() []models.Application
	FileSavePath(string)
	GetFileSavePath() string
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

func (m *appManifest) GetFileSavePath() string {
	return m.savePath
}

func (m *appManifest) Memory(appName string, memory int64) {
	i := m.findOrCreateApplication(appName)
	m.contents[i].Memory = memory
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

	_, err = f.Write([]byte("---\napplications:\n"))

	for _, app := range m.contents {
		if _, err := f.Write([]byte("- name: " + app.Name + "\n")); err != nil {
			return err
		}

		if _, err := f.Write([]byte(fmt.Sprintf("  memory: %dM\n", app.Memory))); err != nil {
			return err
		}

		if _, err := f.Write([]byte(fmt.Sprintf("  instances: %d\n", app.InstanceCount))); err != nil {
			return err
		}

		if app.BuildpackUrl != "" {
			if _, err := f.Write([]byte(fmt.Sprintf("  buildpack: %s\n", app.BuildpackUrl))); err != nil {
				return err
			}
		}

		if app.HealthCheckTimeout > 0 {
			if _, err := f.Write([]byte(fmt.Sprintf("  timeout: %d\n", app.HealthCheckTimeout))); err != nil {
				return err
			}
		}

		if app.Command != "" {
			if _, err := f.Write([]byte(fmt.Sprintf("  command: %s\n", app.Command))); err != nil {
				return err
			}
		}

		if len(app.Routes) > 0 {
			if _, err := f.Write([]byte(fmt.Sprintf("  host: %s\n", app.Routes[0].Host))); err != nil {
				return err
			}
			if _, err := f.Write([]byte(fmt.Sprintf("  domain: %s\n", app.Routes[0].Domain.Name))); err != nil {
				return err
			}
		} else {
			if _, err := f.Write([]byte(fmt.Sprintf("  no-route: true\n"))); err != nil {
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
	_, err := f.Write([]byte(fmt.Sprintf("  services:\n")))
	if err != nil {
		return err
	}
	for _, service := range entries {
		_, err = f.Write([]byte(fmt.Sprintf("  - %s\n", service.Name)))
		if err != nil {
			return err
		}
	}

	return nil
}

func writeEnvironmentVarToFile(f *os.File, envVars map[string]interface{}) error {
	_, err := f.Write([]byte(fmt.Sprintf("  env:\n")))
	if err != nil {
		return err
	}
	for k, v := range envVars {
		_, err = f.Write([]byte(fmt.Sprintf("    %s: %s\n", k, v)))
		if err != nil {
			return err
		}
	}

	return nil
}
