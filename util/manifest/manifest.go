package manifest

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"code.cloudfoundry.org/cli/types"

	yaml "gopkg.in/yaml.v2"
)

type ManifestCreationError struct {
	Err error
}

func (e ManifestCreationError) Error() string {
	return fmt.Sprintf("Error creating manifest file: %s", e.Err.Error())
}

type Manifest struct {
	Applications []Application `yaml:"applications"`
}

type Application struct {
	Buildpack types.FilteredString
	Command   types.FilteredString
	// DiskQuota is the disk size in megabytes.
	DiskQuota      types.NullByteSizeInMb
	DockerImage    string
	DockerUsername string
	DockerPassword string
	// EnvironmentVariables can be any valid json type (ie, strings not
	// guaranteed, although CLI only ships strings).
	EnvironmentVariables    map[string]string
	HealthCheckHTTPEndpoint string
	// HealthCheckType attribute defines the number of seconds that is allocated
	// for starting an application.
	HealthCheckTimeout int
	HealthCheckType    string
	Instances          types.NullInt
	// Memory is the amount of memory in megabytes.
	Memory    types.NullByteSizeInMb
	Name      string
	Path      string
	Routes    []string
	Services  []string
	StackName string
}

func (app Application) String() string {
	return fmt.Sprintf(
		"App Name: '%s', Buildpack IsSet: %t, Buildpack: '%s', Command IsSet: %t, Command: '%s', Disk Quota: '%s', Docker Image: '%s', Health Check HTTP Endpoint: '%s', Health Check Timeout: '%d', Health Check Type: '%s', Instances IsSet: %t, Instances: '%d', Memory: '%s', Path: '%s', Routes: [%s], Services: [%s], Stack Name: '%s'",
		app.Name,
		app.Buildpack.IsSet,
		app.Buildpack.Value,
		app.Command.IsSet,
		app.Command.Value,
		app.DiskQuota,
		app.DockerImage,
		app.HealthCheckHTTPEndpoint,
		app.HealthCheckTimeout,
		app.HealthCheckType,
		app.Instances.IsSet,
		app.Instances.Value,
		app.Memory,
		app.Path,
		strings.Join(app.Routes, ", "),
		strings.Join(app.Services, ", "),
		app.StackName,
	)
}

type manifestApp struct {
	Name                    string            `yaml:"name,omitempty"`
	Buildpack               string            `yaml:"buildpack,omitempty"`
	Command                 string            `yaml:"command,omitempty"`
	DiskQuota               string            `yaml:"disk_quota,omitempty"`
	EnvironmentVariables    map[string]string `yaml:"env,omitempty"`
	HealthCheckHTTPEndpoint string            `yaml:"health-check-http-endpoint,omitempty"`
	HealthCheckType         string            `yaml:"health-check-type,omitempty"`
	Instances               *int              `yaml:"instances,omitempty"`
	Memory                  string            `yaml:"memory,omitempty"`
	Path                    string            `yaml:"path,omitempty"`
	Routes                  []manifestRoute   `yaml:"routes,omitempty"`
	Services                []string          `yaml:"services,omitempty"`
	StackName               string            `yaml:"stack,omitempty"`
	Timeout                 int               `yaml:"timeout,omitempty"`
}

type manifestRoute struct {
	Route string `yaml:"route"`
}

func (app Application) MarshalYAML() (interface{}, error) {
	var m = manifestApp{
		Buildpack:               app.Buildpack.Value,
		Command:                 app.Command.Value,
		EnvironmentVariables:    app.EnvironmentVariables,
		HealthCheckHTTPEndpoint: app.HealthCheckHTTPEndpoint,
		HealthCheckType:         app.HealthCheckType,
		Name:                    app.Name,
		Path:                    app.Path,
		Services:                app.Services,
		StackName:               app.StackName,
		Timeout:                 app.HealthCheckTimeout,
	}
	m.DiskQuota = app.DiskQuota.String()
	m.Memory = app.Memory.String()

	if app.Instances.IsSet {
		m.Instances = &app.Instances.Value
	}
	for _, route := range app.Routes {
		m.Routes = append(m.Routes, manifestRoute{Route: route})
	}

	return m, nil
}

func (app *Application) UnmarshalYAML(unmarshaller func(interface{}) error) error {
	var m manifestApp

	err := unmarshaller(&m)
	if err != nil {
		return err
	}

	app.HealthCheckHTTPEndpoint = m.HealthCheckHTTPEndpoint
	app.HealthCheckType = m.HealthCheckType
	app.Name = m.Name
	app.Path = m.Path
	app.Services = m.Services
	app.StackName = m.StackName
	app.HealthCheckTimeout = m.Timeout
	app.EnvironmentVariables = m.EnvironmentVariables

	app.Instances.ParseIntValue(m.Instances)

	if fmtErr := app.DiskQuota.ParseStringValue(m.DiskQuota); fmtErr != nil {
		return fmtErr
	}

	if fmtErr := app.Memory.ParseStringValue(m.Memory); fmtErr != nil {
		return fmtErr
	}

	for _, route := range m.Routes {
		app.Routes = append(app.Routes, route.Route)
	}

	// "null" values are identical to non-existant values in YAML. In order to
	// detect if an explicit null is given, a manual existance check is required.
	exists := map[string]interface{}{}
	err = unmarshaller(&exists)
	if err != nil {
		return err
	}

	if _, ok := exists["buildpack"]; ok {
		app.Buildpack.ParseValue(m.Buildpack)
		app.Buildpack.IsSet = true
	}

	if _, ok := exists["command"]; ok {
		app.Command.ParseValue(m.Command)
		app.Command.IsSet = true
	}

	return nil
}

func ReadAndMergeManifests(pathToManifest string) ([]Application, error) {
	// Read all manifest files
	raw, err := ioutil.ReadFile(pathToManifest)
	if err != nil {
		return nil, err
	}

	var manifest Manifest
	err = yaml.Unmarshal(raw, &manifest)
	if err != nil {
		return nil, err
	}

	for i, app := range manifest.Applications {
		if app.Path != "" && !filepath.IsAbs(app.Path) {
			manifest.Applications[i].Path = filepath.Join(filepath.Dir(pathToManifest), app.Path)
		}
	}

	// Merge all manifest files
	return manifest.Applications, err
}

// filepath will be created if it doesn't exist
func WriteApplicationManifest(application Application, filePath string) error {
	manifest := Manifest{Applications: []Application{application}}
	manifestBytes, err := yaml.Marshal(manifest)
	if err != nil {
		return ManifestCreationError{Err: err}
	}

	err = ioutil.WriteFile(filePath, manifestBytes, 0644)
	if err != nil {
		return ManifestCreationError{Err: err}
	}

	return nil

}
