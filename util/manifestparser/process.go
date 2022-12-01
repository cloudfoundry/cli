package manifestparser

import (
	"errors"
	"reflect"
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
)

type Process struct {
	DiskQuota               string                   `yaml:"disk_quota,omitempty"`
	HealthCheckEndpoint     string                   `yaml:"health-check-http-endpoint,omitempty"`
	HealthCheckType         constant.HealthCheckType `yaml:"health-check-type,omitempty"`
	HealthCheckTimeout      int64                    `yaml:"timeout,omitempty"`
	Instances               *int                     `yaml:"instances,omitempty"`
	Memory                  string                   `yaml:"memory,omitempty"`
	Type                    string                   `yaml:"type"`
	LogRateLimit            string                   `yaml:"log-rate-limit-per-second,omitempty"`
	RemainingManifestFields map[string]interface{}   `yaml:"-,inline"`
}

func (process *Process) SetStartCommand(command string) {
	if process.RemainingManifestFields == nil {
		process.RemainingManifestFields = map[string]interface{}{}
	}

	if command == "" {
		process.RemainingManifestFields["command"] = nil
	} else {
		process.RemainingManifestFields["command"] = command
	}
}

func (process *Process) UnmarshalYAML(unmarshal func(v interface{}) error) error {
	// This prevents infinite recursion. The Alias type does not implement the unmarshaller interface
	// so by casting application to a alias pointer, it will unmarshal into the same memory without calling
	// UnmarshalYAML on itself infinite times
	type Alias Process
	aliasPntr := (*Alias)(process)

	err := unmarshal(aliasPntr)
	if err != nil {
		return err
	}

	err = unmarshal(&process.RemainingManifestFields)
	if err != nil {
		return err
	}

	value := reflect.ValueOf(*process)
	removeDuplicateMapKeys(value, process.RemainingManifestFields)
	// old style was `disk_quota` (underscore not hyphen)
	// we maintain backwards-compatibility by supporting both flavors
	if process.RemainingManifestFields["disk-quota"] != nil {
		if process.DiskQuota != "" {
			return errors.New("cannot define both `disk_quota` and `disk-quota`")
		}
		diskQuota, ok := process.RemainingManifestFields["disk-quota"].(string)
		if !ok {
			return errors.New("`disk-quota` must be a string")
		}
		process.DiskQuota = diskQuota
		delete(process.RemainingManifestFields, "disk-quota")
	}

	return nil
}

func removeDuplicateMapKeys(model reflect.Value, fieldMap map[string]interface{}) {
	for i := 0; i < model.NumField(); i++ {
		structField := model.Type().Field(i)

		yamlTag := strings.Split(structField.Tag.Get("yaml"), ",")
		yamlKey := yamlTag[0]

		if yamlKey == "" {
			yamlKey = structField.Name
		}

		delete(fieldMap, yamlKey)
	}
}
