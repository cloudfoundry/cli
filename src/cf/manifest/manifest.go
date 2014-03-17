package manifest

import (
	"cf/formatters"
	"cf/models"
	"errors"
	"fmt"
	"generic"
	"path/filepath"
	"regexp"
	"strconv"
)

type Manifest struct {
	Path string
	Data generic.Map
}

func NewEmptyManifest() (m *Manifest) {
	return &Manifest{Data: generic.NewMap()}
}

func (m Manifest) Applications() (apps []models.AppParams, errs ManifestErrors) {
	errs = walkManifestLookingForProperties(m.Data)
	if !errs.Empty() {
		return
	}

	if m.Data.Has("applications") {
		appMaps, ok := m.Data.Get("applications").([]interface{})
		if !ok {
			errs = append(errs, errors.New("Expected applications to be a list"))
			return
		}

		globalProperties := m.Data.Except([]interface{}{"applications"})

		for _, appData := range appMaps {
			if !generic.IsMappable(appData) {
				errs = append(errs, errors.New("Expected application to be a dictionary"))
				continue
			}

			appMap := generic.DeepMerge(globalProperties, generic.NewMap(appData))

			basePath := filepath.Dir(m.Path)
			appParams, appErrs := mapToAppParams(basePath, appMap)
			if !appErrs.Empty() {
				errs = append(errs, appErrs)
				continue
			}

			apps = append(apps, appParams)
		}
	}

	return
}

func walkManifestLookingForProperties(data generic.Map) (errs ManifestErrors) {
	generic.Each(data, func(key, value interface{}) {
		errs = append(errs, walkMapLookingForProperties(value)...)
	})

	return
}

var propertyRegex = regexp.MustCompile(`\${[\w-]+}`)

func walkMapLookingForProperties(value interface{}) (errs ManifestErrors) {
	switch value := value.(type) {
	case string:
		match := propertyRegex.FindString(value)
		if match != "" {
			err := errors.New(fmt.Sprintf("Property '%s' found in manifest. This feature is no longer supported. Please remove it and try again.", match))
			errs = append(errs, err)
		}
	case []interface{}:
		for _, item := range value {
			errs = append(errs, walkMapLookingForProperties(item)...)
		}
	case map[interface{}]interface{}:
		for _, item := range value {
			errs = append(errs, walkMapLookingForProperties(item)...)
		}
	}
	return
}

func mapToAppParams(basePath string, yamlMap generic.Map) (appParams models.AppParams, errs ManifestErrors) {
	errs = checkForNulls(yamlMap)
	if !errs.Empty() {
		return
	}

	appParams.BuildpackUrl = stringVal(yamlMap, "buildpack", &errs)
	appParams.DiskQuota = bytesVal(yamlMap, "disk_quota", &errs)
	appParams.Domain = stringVal(yamlMap, "domain", &errs)
	appParams.Host = stringVal(yamlMap, "host", &errs)
	appParams.Name = stringVal(yamlMap, "name", &errs)
	appParams.Path = stringVal(yamlMap, "path", &errs)
	appParams.StackName = stringVal(yamlMap, "stack", &errs)
	appParams.Command = stringOrNullVal(yamlMap, "command", &errs)
	appParams.Memory = bytesVal(yamlMap, "memory", &errs)
	appParams.InstanceCount = intVal(yamlMap, "instances", &errs)
	appParams.HealthCheckTimeout = intVal(yamlMap, "timeout", &errs)
	appParams.NoRoute = boolVal(yamlMap, "no-route", &errs)
	appParams.Services = sliceOrEmptyVal(yamlMap, "services", &errs)
	appParams.EnvironmentVars = envVarOrEmptyMap(yamlMap, &errs)

	if appParams.Path != nil {
		path := *appParams.Path
		if filepath.IsAbs(path) {
			path = filepath.Clean(path)
		} else {
			path = filepath.Join(basePath, path)
		}
		appParams.Path = &path
	}

	return
}

func checkForNulls(yamlMap generic.Map) (errs ManifestErrors) {
	generic.Each(yamlMap, func(key interface{}, value interface{}) {
		if key == "command" {
			return
		}
		if value == nil {
			errs = append(errs, errors.New(fmt.Sprintf("%s should not be null", key)))
		}
	})

	return
}

func stringVal(yamlMap generic.Map, key string, errs *ManifestErrors) *string {
	val := yamlMap.Get(key)
	if val == nil {
		return nil
	}
	result, ok := val.(string)
	if !ok {
		*errs = append(*errs, errors.New(fmt.Sprintf("%s must be a string value", key)))
		return nil
	}
	return &result
}

func stringOrNullVal(yamlMap generic.Map, key string, errs *ManifestErrors) *string {
	if !yamlMap.Has(key) {
		return nil
	}
	switch val := yamlMap.Get(key).(type) {
	case string:
		return &val
	case nil:
		empty := ""
		return &empty
	default:
		*errs = append(*errs, errors.New(fmt.Sprintf("%s must be a string or null value", key)))
		return nil
	}
}

func bytesVal(yamlMap generic.Map, key string, errs *ManifestErrors) *uint64 {
	yamlVal := yamlMap.Get(key)
	if yamlVal == nil {
		return nil
	}
	value, err := formatters.ToMegabytes(yamlVal.(string))
	if err != nil {
		*errs = append(*errs, errors.New(fmt.Sprintf("Unexpected value for %s :\n%s", key, err.Error())))
		return nil
	}
	return &value
}

func intVal(yamlMap generic.Map, key string, errs *ManifestErrors) *int {
	var (
		intVal int
		err    error
	)

	switch val := yamlMap.Get(key).(type) {
	case string:
		intVal, err = strconv.Atoi(val)
	case int:
		intVal = val
	case int64:
		intVal = int(val)
	case nil:
		return nil
	default:
		err = errors.New(fmt.Sprintf("Expected %s to be a number, but it was a %T.", key, val))
	}

	if err != nil {
		*errs = append(*errs, err)
		return nil
	}

	return &intVal
}

func boolVal(yamlMap generic.Map, key string, errs *ManifestErrors) *bool {
	switch val := yamlMap.Get(key).(type) {
	case nil:
		return nil
	case bool:
		return &val
	case string:
		boolVal := val == "true"
		return &boolVal
	default:
		*errs = append(*errs, errors.New(fmt.Sprintf("Expected %s to be a boolean.", key)))
		return nil
	}
}

func sliceOrEmptyVal(yamlMap generic.Map, key string, errs *ManifestErrors) *[]string {
	if !yamlMap.Has(key) {
		return new([]string)
	}

	var (
		stringSlice []string
		err         error
	)

	errMsg := fmt.Sprintf("Expected %s to be a list of strings.", key)

	switch input := yamlMap.Get(key).(type) {
	case []interface{}:
		for _, value := range input {
			stringValue, ok := value.(string)
			if !ok {
				err = errors.New(errMsg)
				break
			}
			stringSlice = append(stringSlice, stringValue)
		}
	default:
		err = errors.New(errMsg)
	}

	if err != nil {
		*errs = append(*errs, err)
		return nil
	}

	return &stringSlice
}

func envVarOrEmptyMap(yamlMap generic.Map, errs *ManifestErrors) *map[string]string {
	key := "env"
	switch envVars := yamlMap.Get(key).(type) {
	case nil:
		aMap := make(map[string]string, 0)
		return &aMap
	case map[string]interface{}:
		yamlMap.Set(key, generic.NewMap(yamlMap.Get(key)))
		return envVarOrEmptyMap(yamlMap, errs)
	case map[interface{}]interface{}:
		yamlMap.Set(key, generic.NewMap(yamlMap.Get(key)))
		return envVarOrEmptyMap(yamlMap, errs)
	case generic.Map:
		merrs := validateEnvVars(envVars)
		if merrs != nil {
			*errs = append(*errs, merrs)
			return nil
		}

		result := make(map[string]string, envVars.Count())
		generic.Each(envVars, func(key, value interface{}) {
			result[key.(string)] = value.(string)
		})
		return &result
	default:
		*errs = append(*errs, errors.New(fmt.Sprintf("Expected %s to be a set of key => value, but it was a %T.", key, envVars)))
		return nil
	}
}

func validateEnvVars(input generic.Map) (errs ManifestErrors) {
	generic.Each(input, func(key, value interface{}) {
		if value == nil {
			errs = append(errs, errors.New(fmt.Sprintf("env var '%s' should not be null", key)))
		}
	})
	return
}
