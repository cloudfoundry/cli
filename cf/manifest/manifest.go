package manifest

import (
	"fmt"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/formatters"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/generic"
	"github.com/cloudfoundry/cli/words"
)

type Manifest struct {
	Path string
	Data generic.Map
}

func NewEmptyManifest() (m *Manifest) {
	return &Manifest{Data: generic.NewMap()}
}

func (m Manifest) Applications() (apps []models.AppParams, err error) {
	rawData, errs := expandProperties(m.Data, words.NewWordGenerator())
	if len(errs) > 0 {
		err = errors.NewWithSlice(errs)
		return
	}

	data := generic.NewMap(rawData)
	appMaps, errs := m.getAppMaps(data)
	if len(errs) > 0 {
		err = errors.NewWithSlice(errs)
		return
	}

	for _, appMap := range appMaps {
		app, errs := mapToAppParams(filepath.Dir(m.Path), appMap)
		if len(errs) > 0 {
			err = errors.NewWithSlice(errs)
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
			errs = append(errs, errors.New(T("Expected applications to be a list")))
			return
		}

		for _, appData := range appMaps {
			if !generic.IsMappable(appData) {
				errs = append(errs, errors.NewWithFmt(T("Expected application to be a list of key/value pairs\nError occurred in manifest near:\n'{{.YmlSnippet}}'",
					map[string]interface{}{"YmlSnippet": appData})))
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

func expandProperties(input interface{}, babbler words.WordGenerator) (output interface{}, errs []error) {
	switch input := input.(type) {
	case string:
		match := propertyRegex.FindStringSubmatch(input)
		if match != nil {
			if match[0] == "${random-word}" {
				output = strings.Replace(input, "${random-word}", strings.ToLower(babbler.Babble()), -1)
			} else {
				err := errors.NewWithFmt(T("Property '{{.PropertyName}}' found in manifest. This feature is no longer supported. Please remove it and try again.",
					map[string]interface{}{"PropertyName": match[0]}))
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

func mapToAppParams(basePath string, yamlMap generic.Map) (appParams models.AppParams, errs []error) {
	errs = checkForNulls(yamlMap)
	if len(errs) > 0 {
		return
	}

	appParams.BuildpackUrl = stringValOrDefault(yamlMap, "buildpack", &errs)
	appParams.DiskQuota = bytesVal(yamlMap, "disk_quota", &errs)
	appParams.Domain = stringVal(yamlMap, "domain", &errs)
	appParams.Host = stringVal(yamlMap, "host", &errs)
	appParams.Name = stringVal(yamlMap, "name", &errs)
	appParams.Path = stringVal(yamlMap, "path", &errs)
	appParams.StackName = stringVal(yamlMap, "stack", &errs)
	appParams.Command = stringValOrDefault(yamlMap, "command", &errs)
	appParams.Memory = bytesVal(yamlMap, "memory", &errs)
	appParams.InstanceCount = intVal(yamlMap, "instances", &errs)
	appParams.HealthCheckTimeout = intVal(yamlMap, "timeout", &errs)
	appParams.NoRoute = boolVal(yamlMap, "no-route", &errs)
	appParams.UseRandomHostname = boolVal(yamlMap, "random-route", &errs)
	appParams.ServicesToBind = sliceOrEmptyVal(yamlMap, "services", &errs)
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

func checkForNulls(yamlMap generic.Map) (errs []error) {
	generic.Each(yamlMap, func(key interface{}, value interface{}) {
		if key == "command" || key == "buildpack" {
			return
		}
		if value == nil {
			errs = append(errs, errors.NewWithFmt(T("{{.PropertyName}} should not be null", map[string]interface{}{"PropertyName": key})))
		}
	})

	return
}

func stringVal(yamlMap generic.Map, key string, errs *[]error) *string {
	val := yamlMap.Get(key)
	if val == nil {
		return nil
	}
	result, ok := val.(string)
	if !ok {
		*errs = append(*errs, errors.NewWithFmt(T("{{.PropertyName}} must be a string value", map[string]interface{}{"PropertyName": key})))
		return nil
	}
	return &result
}

func stringValOrDefault(yamlMap generic.Map, key string, errs *[]error) *string {
	if !yamlMap.Has(key) {
		return nil
	}
	empty := ""
	switch val := yamlMap.Get(key).(type) {
	case string:
		if val == "default" {
			return &empty
		} else {
			return &val
		}
	case nil:
		return &empty
	default:
		*errs = append(*errs, errors.NewWithFmt(T("{{.PropertyName}} must be a string or null value", map[string]interface{}{"PropertyName": key})))
		return nil
	}
}

func bytesVal(yamlMap generic.Map, key string, errs *[]error) *uint64 {
	yamlVal := yamlMap.Get(key)
	if yamlVal == nil {
		return nil
	}
	value, err := formatters.ToMegabytes(yamlVal.(string))
	if err != nil {
		*errs = append(*errs, errors.NewWithFmt(T("Unexpected value for {{.PropertyName}} :\n{{.Error}}",
			map[string]interface{}{"PropertyName": key, "Error": err.Error()})))
		return nil
	}
	return &value
}

func intVal(yamlMap generic.Map, key string, errs *[]error) *int {
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
		err = errors.NewWithFmt(T("Expected {{.PropertyName}} to be a number, but it was a {{.PropertyType}}.",
			map[string]interface{}{"PropertyName": key, "PropertyType": val}))
	}

	if err != nil {
		*errs = append(*errs, err)
		return nil
	}

	return &intVal
}

func boolVal(yamlMap generic.Map, key string, errs *[]error) bool {
	switch val := yamlMap.Get(key).(type) {
	case nil:
		return false
	case bool:
		return val
	case string:
		return val == "true"
	default:
		*errs = append(*errs, errors.NewWithFmt(T("Expected {{.PropertyName}} to be a boolean.", map[string]interface{}{"PropertyName": key})))
		return false
	}
}

func sliceOrEmptyVal(yamlMap generic.Map, key string, errs *[]error) *[]string {
	if !yamlMap.Has(key) {
		return new([]string)
	}

	var (
		stringSlice []string
		err         error
	)

	sliceErr := errors.NewWithFmt(T("Expected {{.PropertyName}} to be a list of strings.",
		map[string]interface{}{"PropertyName": key}))

	switch input := yamlMap.Get(key).(type) {
	case []interface{}:
		for _, value := range input {
			stringValue, ok := value.(string)
			if !ok {
				err = sliceErr
				break
			}
			stringSlice = append(stringSlice, stringValue)
		}
	default:
		err = sliceErr
	}

	if err != nil {
		*errs = append(*errs, err)
		return nil
	}

	return &stringSlice
}

func envVarOrEmptyMap(yamlMap generic.Map, errs *[]error) *map[string]string {
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
			*errs = append(*errs, merrs...)
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
				*errs = append(*errs, errors.NewWithFmt(T("Expected environment variable {{.PropertyName}} to have a string value, but it was a {{.PropertyType}}.",
					map[string]interface{}{"PropertyName": key, "PropertyType": value})))
			}

		})
		return &result
	default:
		*errs = append(*errs, errors.NewWithFmt(T("Expected {{.Name}} to be a set of key => value, but it was a {{.Type}}.",
			map[string]interface{}{"Name": key, "Type": envVars})))
		return nil
	}
}

func validateEnvVars(input generic.Map) (errs []error) {
	generic.Each(input, func(key, value interface{}) {
		if value == nil {
			errs = append(errs, errors.New(fmt.Sprintf(T("env var '{{.PropertyName}}' should not be null",
				map[string]interface{}{"PropertyName": key}))))
		}
	})
	return
}
