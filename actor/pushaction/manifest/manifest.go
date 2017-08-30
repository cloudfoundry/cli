package manifest

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"code.cloudfoundry.org/cli/types"

	"github.com/cloudfoundry/bytefmt"

	yaml "gopkg.in/yaml.v2"
)

type Manifest struct {
	Applications []Application `yaml:"applications"`
}

type Application struct {
	Buildpack types.FilteredString
	Command   types.FilteredString
	// DiskQuota is the disk size in megabytes.
	DiskQuota      uint64
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
	Memory    uint64
	Name      string
	Path      string
	Routes    []string
	Services  []string
	StackName string
}

func (app Application) String() string {
	return fmt.Sprintf(
		"App Name: '%s', Buildpack IsSet: %t, Buildpack: '%s', Command IsSet: %t, Command: '%s', Disk Quota: '%d', Docker Image: '%s', Health Check HTTP Endpoint: '%s', Health Check Timeout: '%d', Health Check Type: '%s', Instances IsSet: %t, Instances: '%d', Memory: '%d', Path: '%s', Routes: [%s], Services: [%s], Stack Name: '%s'",
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

func (app *Application) UnmarshalYAML(unmarshaller func(interface{}) error) error {
	var manifestApp struct {
		Buildpack               string            `yaml:"buildpack"`
		Command                 string            `yaml:"command"`
		DiskQuota               string            `yaml:"disk_quota"`
		EnvironmentVariables    map[string]string `yaml:"env"`
		HealthCheckHTTPEndpoint string            `yaml:"health-check-http-endpoint"`
		HealthCheckType         string            `yaml:"health-check-type"`
		Instances               string            `yaml:"instances"`
		Memory                  string            `yaml:"memory"`
		Name                    string            `yaml:"name"`
		Path                    string            `yaml:"path"`
		Routes                  []struct {
			Route string `json:"route"`
		} `json:"routes"`
		Services  []string `yaml:"services"`
		StackName string   `yaml:"stack"`
		Timeout   int      `yaml:"timeout"`
	}

	err := unmarshaller(&manifestApp)
	if err != nil {
		return err
	}

	app.HealthCheckHTTPEndpoint = manifestApp.HealthCheckHTTPEndpoint
	app.HealthCheckType = manifestApp.HealthCheckType
	app.Name = manifestApp.Name
	app.Path = manifestApp.Path
	app.Services = manifestApp.Services
	app.StackName = manifestApp.StackName
	app.HealthCheckTimeout = manifestApp.Timeout
	app.EnvironmentVariables = manifestApp.EnvironmentVariables

	err = app.Instances.ParseFlagValue(manifestApp.Instances)
	if err != nil {
		return err
	}

	if manifestApp.DiskQuota != "" {
		disk, fmtErr := bytefmt.ToMegabytes(manifestApp.DiskQuota)
		if fmtErr != nil {
			return fmtErr
		}
		app.DiskQuota = disk
	}

	if manifestApp.Memory != "" {
		memory, fmtErr := bytefmt.ToMegabytes(manifestApp.Memory)
		if fmtErr != nil {
			return fmtErr
		}
		app.Memory = memory
	}

	for _, route := range manifestApp.Routes {
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
		app.Buildpack.ParseValue(manifestApp.Buildpack)
		app.Buildpack.IsSet = true
	}

	if _, ok := exists["command"]; ok {
		app.Command.ParseValue(manifestApp.Command)
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
