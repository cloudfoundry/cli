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
	"strings"
	"words"
)

type Manifest struct {
	Path string
	Data generic.Map
}

func NewEmptyManifest() (m *Manifest) {
	return &Manifest{Data: generic.NewMap()}
}

func (m Manifest) Applications() (apps []models.AppParams, errs ManifestErrors) {
	rawData, errs := expandProperties(m.Data, words.NewWordGenerator())
	if !errs.Empty() {
		return
	}

	data := generic.NewMap(rawData)
	appMaps, errs := m.getAppMaps(data)
	if !errs.Empty() {
		return
	}

	for _, appMap := range appMaps {
		app, appErrs := mapToAppParams(filepath.Dir(m.Path), appMap)
		if !appErrs.Empty() {
			errs = append(errs, appErrs)
			continue
		}

		apps = append(apps, app)
	}

	return
}

func (m Manifest) getAppMaps(data generic.Map) (apps []generic.Map, errs []error) {
	globalProperties := data.Except([]interface{}{"applications"})

	if data.Has("applications") {
		appMaps, ok := data.Get("applications").([]interface{})
		if !ok {
			errs = append(errs, errors.New("Expected applications to be a list"))
			return
		}

		for _, appData := range appMaps {
			if !generic.IsMappable(appData) {
				errs = append(errs, errors.New(fmt.Sprintf("Expected application to be a list of key/value pairs\nError occurred in manifest near:\n'%s'", appData)))
				continue
			}

			appMap := generic.DeepMerge(globalProperties, generic.NewMap(appData))
			apps = append(apps, appMap)
		}
	} else {
		apps = append(apps, globalProperties)
	}

	return
}

var propertyRegex = regexp.MustCompile(`\${[\w-]+}`)

func expandProperties(input interface{}, babbler words.WordGenerator) (output interface{}, errs ManifestErrors) {
	switch input := input.(type) {
	case string:
		match := propertyRegex.FindStringSubmatch(input)
		if match != nil {
			if match[0] == "${random-word}" {
				output = strings.Replace(input, "${random-word}", strings.ToLower(babbler.Babble()), -1)
			} else {
				err := errors.New(fmt.Sprintf("Property '%s' found in manifest. This feature is no longer supported. Please remove it and try again.", match[0]))
				errs = append(errs, err)
			}
		} else {
			output = input
		}
	case []interface{}:
		outputSlice := make([]interface{}, len(input))
		for index, item := range input {
			itemOutput, itemErrs := expandProperties(item, babbler)
			outputSlice[index] = itemOutput
			errs = append(errs, itemErrs...)
		}
		output = outputSlice
	case map[interface{}]interface{}:
		outputMap := make(map[interface{}]interface{})
		for key, value := range input {
			itemOutput, itemErrs := expandProperties(value, babbler)
			outputMap[key] = itemOutput
			errs = append(errs, itemErrs...)
		}
		output = outputMap
	case generic.Map:
		outputMap := generic.NewMap()
		generic.Each(input, func(key, value interface{}) {
			itemOutput, itemErrs := expandProperties(value, babbler)
			outputMap.Set(key, itemOutput)
			errs = append(errs, itemErrs...)
		})
		output = outputMap
	default:
		output = input
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
	appParams.UseRandomHostname = boolVal(yamlMap, "random-route", &errs)
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

func boolVal(yamlMap generic.Map, key string, errs *ManifestErrors) bool {
	switch val := yamlMap.Get(key).(type) {
	case nil:
		return false
	case bool:
		return val
	case string:
		return val == "true"
	default:
		*errs = append(*errs, errors.New(fmt.Sprintf("Expected %s to be a boolean.", key)))
		return false
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

			switch value.(type) {
			case string:
				result[key.(string)] = value.(string)
			case int64, int, int32:
				result[key.(string)] = fmt.Sprintf("%d", value)
			case float32, float64:
				result[key.(string)] = fmt.Sprintf("%f", value)
			default:
				*errs = append(*errs, errors.New(fmt.Sprintf("Expected environment variable %s to have a string value, but it was a %T.", key, value)))
			}

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
