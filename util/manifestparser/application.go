package manifestparser

import (
	"errors"
	"reflect"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
)

type Docker struct {
	Image    string `yaml:"image,omitempty"`
	Username string `yaml:"username,omitempty"`
}

// ApplicationModel can be accessed through the top level Application struct To
// add a field for the CLI to extract from the manifest, just add it to this
// struct.
type Application struct {
	Name                    string                   `yaml:"name"`
	DiskQuota               string                   `yaml:"disk-quota,omitempty"`
	Docker                  *Docker                  `yaml:"docker,omitempty"`
	HealthCheckType         constant.HealthCheckType `yaml:"health-check-type,omitempty"`
	HealthCheckEndpoint     string                   `yaml:"health-check-http-endpoint,omitempty"`
	HealthCheckTimeout      int64                    `yaml:"timeout,omitempty"`
	Instances               *int                     `yaml:"instances,omitempty"`
	Path                    string                   `yaml:"path,omitempty"`
	Processes               []Process                `yaml:"processes,omitempty"`
	Memory                  string                   `yaml:"memory,omitempty"`
	NoRoute                 bool                     `yaml:"no-route,omitempty"`
	RandomRoute             bool                     `yaml:"random-route,omitempty"`
	DefaultRoute            bool                     `yaml:"default-route,omitempty"`
	Stack                   string                   `yaml:"stack,omitempty"`
	LogRateLimit            string                   `yaml:"log-rate-limit-per-second,omitempty"`
	RemainingManifestFields map[string]interface{}   `yaml:"-,inline"`
}

func (application Application) HasBuildpacks() bool {
	_, ok := application.RemainingManifestFields["buildpacks"]
	return ok
}

func (application *Application) SetBuildpacks(buildpacks []string) {
	if application.RemainingManifestFields == nil {
		application.RemainingManifestFields = map[string]interface{}{}
	}

	application.RemainingManifestFields["buildpacks"] = buildpacks
}

func (application *Application) SetStartCommand(command string) {
	if application.RemainingManifestFields == nil {
		application.RemainingManifestFields = map[string]interface{}{}
	}

	if command == "" {
		application.RemainingManifestFields["command"] = nil
	} else {
		application.RemainingManifestFields["command"] = command
	}
}

func (application *Application) UnmarshalYAML(unmarshal func(v interface{}) error) error {
	// This prevents infinite recursion. The alias type does not implement the unmarshaller interface
	// so by casting application to a alias pointer, it will unmarshal into the same memory without calling
	// UnmarshalYAML on itself infinite times
	type Alias Application
	aliasPntr := (*Alias)(application)

	err := unmarshal(aliasPntr)
	if err != nil {
		return err
	}

	err = unmarshal(&application.RemainingManifestFields)
	if err != nil {
		return err
	}

	value := reflect.ValueOf(*application)
	removeDuplicateMapKeys(value, application.RemainingManifestFields)
	// old style was `disk_quota` (underscore not hyphen)
	// we maintain backwards-compatibility by supporting both flavors
	if application.RemainingManifestFields["disk_quota"] != nil {
		if application.DiskQuota != "" {
			return errors.New("cannot define both `disk_quota` and `disk-quota`")
		}
		diskQuota, ok := application.RemainingManifestFields["disk_quota"].(string)
		if !ok {
			return errors.New("`disk_quota` must be a string")
		}
		application.DiskQuota = diskQuota
		delete(application.RemainingManifestFields, "disk_quota")
	}

	return nil
}
