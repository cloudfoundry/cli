package manifest

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/cloudfoundry/bytefmt"

	yaml "gopkg.in/yaml.v2"
)

type Manifest struct {
	Applications []Application `yaml:"applications"`
}

type Application struct {
	BuildpackName string
	Command       string
	// DiskQuota is the disk size in megabytes.
	DiskQuota               uint64
	DockerImage             string
	HealthCheckHTTPEndpoint string
	// HealthCheckType attribute defines the number of seconds that is allocated
	// for starting an application.
	HealthCheckTimeout int
	HealthCheckType    string
	Instances          int
	// Memory is the amount of memory in megabytes.
	Memory    uint64
	Name      string
	Path      string
	StackName string
}

func (app Application) String() string {
	return fmt.Sprintf(
		"App Name: '%s', Buildpack: '%s', Command: '%s', Disk Quota: '%d', Docker Image: '%s', Health Check HTTP Endpoint: '%s', Health Check Timeout: '%d', Health Check Type: '%s', Instances: '%d', Memory: '%d', Path: '%s', Stack Name: '%s'",
		app.Name,
		app.BuildpackName,
		app.Command,
		app.DiskQuota,
		app.DockerImage,
		app.HealthCheckHTTPEndpoint,
		app.HealthCheckTimeout,
		app.HealthCheckType,
		app.Instances,
		app.Memory,
		app.Path,
		app.StackName,
	)
}

func (a *Application) UnmarshalYAML(unmarshaller func(interface{}) error) error {
	var manifestApp struct {
		Buildpack               string `yaml:"buildpack"`
		Command                 string `yaml:"command"`
		DiskQuota               string `yaml:"disk_quota"`
		HealthCheckHTTPEndpoint string `yaml:"health-check-http-endpoint"`
		HealthCheckType         string `yaml:"health-check-type"`
		Instances               int    `yaml:"instances"`
		Memory                  string `yaml:"memory"`
		Name                    string `yaml:"name"`
		Path                    string `yaml:"path"`
		StackName               string `yaml:"stack"`
		Timeout                 int    `yaml:"timeout"`
	}

	err := unmarshaller(&manifestApp)
	if err != nil {
		return err
	}

	a.BuildpackName = manifestApp.Buildpack
	a.Command = manifestApp.Command
	a.HealthCheckHTTPEndpoint = manifestApp.HealthCheckHTTPEndpoint
	a.HealthCheckType = manifestApp.HealthCheckType
	a.Instances = manifestApp.Instances
	a.Name = manifestApp.Name
	a.Path = manifestApp.Path
	a.StackName = manifestApp.StackName
	a.HealthCheckTimeout = manifestApp.Timeout

	if manifestApp.DiskQuota != "" {
		disk, err := bytefmt.ToMegabytes(manifestApp.DiskQuota)
		if err != nil {
			return err
		}
		a.DiskQuota = disk
	}

	if manifestApp.Memory != "" {
		memory, err := bytefmt.ToMegabytes(manifestApp.Memory)
		if err != nil {
			return err
		}
		a.Memory = memory
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
