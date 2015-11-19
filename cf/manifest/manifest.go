package manifest

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	. "github.com/cloudfoundry/cli/cf/i18n"

	"github.com/cloudfoundry/cli/cf/formatters"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/generic"
	"github.com/cloudfoundry/cli/words/generator"
)

type Manifest struct {
	Path string
	Data generic.Map
}

func NewEmptyManifest() (m *Manifest) {
	return &Manifest{Data: generic.NewMap()}
}

func (m Manifest) Applications() ([]models.AppParams, error) {
	rawData, err := expandProperties(m.Data, generator.NewWordGenerator())
	if err != nil {
		return []models.AppParams{}, err
	}

	data := generic.NewMap(rawData)
	appMaps, err := m.getAppMaps(data)
	if err != nil {
		return []models.AppParams{}, err
	}

	var apps []models.AppParams
	var mapToAppErrs []error
	for _, appMap := range appMaps {
		app, err := mapToAppParams(filepath.Dir(m.Path), appMap)
		if err != nil {
			mapToAppErrs = append(mapToAppErrs, err)
			continue
		}

		apps = append(apps, app)
	}

	if len(mapToAppErrs) > 0 {
		message := ""
		for i := range mapToAppErrs {
			message = message + fmt.Sprintf("%s\n", mapToAppErrs[i].Error())
		}
		return []models.AppParams{}, errors.New(message)
	}

	return apps, nil
}

func (m Manifest) getAppMaps(data generic.Map) ([]generic.Map, error) {
	globalProperties := data.Except([]interface{}{"applications"})

	var apps []generic.Map
	var errs []error
	if data.Has("applications") {
		appMaps, ok := data.Get("applications").([]interface{})
		if !ok {
			return []generic.Map{}, errors.New(T("Expected applications to be a list"))
		}

		for _, appData := range appMaps {
			if !generic.IsMappable(appData) {
				errs = append(errs, fmt.Errorf(T("Expected application to be a list of key/value pairs\nError occurred in manifest near:\n'{{.YmlSnippet}}'",
					map[string]interface{}{"YmlSnippet": appData})))
				continue
			}

			appMap := generic.DeepMerge(globalProperties, generic.NewMap(appData))
			apps = append(apps, appMap)
		}
	} else {
		apps = append(apps, globalProperties)
	}

	if len(errs) > 0 {
		message := ""
		for i := range errs {
			message = message + fmt.Sprintf("%s\n", errs[i].Error())
		}
		return []generic.Map{}, errors.New(message)
	}

	return apps, nil
}

var propertyRegex = regexp.MustCompile(`\${[\w-]+}`)

func expandProperties(input interface{}, babbler generator.WordGenerator) (interface{}, error) {
	var errs []error
	var output interface{}

	switch input := input.(type) {
	case string:
		match := propertyRegex.FindStringSubmatch(input)
		if match != nil {
			if match[0] == "${random-word}" {
				output = strings.Replace(input, "${random-word}", strings.ToLower(babbler.Babble()), -1)
			} else {
				err := fmt.Errorf(T("Property '{{.PropertyName}}' found in manifest. This feature is no longer supported. Please remove it and try again.",
					map[string]interface{}{"PropertyName": match[0]}))
				errs = append(errs, err)
			}
		} else {
			output = input
		}
	case []interface{}:
		outputSlice := make([]interface{}, len(input))
		for index, item := range input {
			itemOutput, itemErr := expandProperties(item, babbler)
			if itemErr != nil {
				errs = append(errs, itemErr)
				break
			}
			outputSlice[index] = itemOutput
		}
		output = outputSlice
	case map[interface{}]interface{}:
		outputMap := make(map[interface{}]interface{})
		for key, value := range input {
			itemOutput, itemErr := expandProperties(value, babbler)
			if itemErr != nil {
				errs = append(errs, itemErr)
				break
			}
			outputMap[key] = itemOutput
		}
		output = outputMap
	case generic.Map:
		outputMap := generic.NewMap()
		generic.Each(input, func(key, value interface{}) {
			itemOutput, itemErr := expandProperties(value, babbler)
			if itemErr != nil {
				errs = append(errs, itemErr)
				return
			}
			outputMap.Set(key, itemOutput)
		})
		output = outputMap
	default:
		output = input
	}

	if len(errs) > 0 {
		message := ""
		for _, err := range errs {
			message = message + fmt.Sprintf("%s\n", err.Error())
		}
		return nil, errors.New(message)
	}

	return output, nil
}

func mapToAppParams(basePath string, yamlMap generic.Map) (models.AppParams, error) {
	err := checkForNulls(yamlMap)
	if err != nil {
		return models.AppParams{}, err
	}

	var appParams models.AppParams
	var errs []error
	appParams.BuildpackUrl = stringValOrDefault(yamlMap, "buildpack", &errs)
	appParams.DiskQuota = bytesVal(yamlMap, "disk_quota", &errs)

	domainAry := *sliceOrEmptyVal(yamlMap, "domains", &errs)
	if domain := stringVal(yamlMap, "domain", &errs); domain != nil {
		domainAry = append(domainAry, *domain)
	}
	appParams.Domains = removeDuplicatedValue(domainAry)

	hostsArr := *sliceOrEmptyVal(yamlMap, "hosts", &errs)
	if host := stringVal(yamlMap, "host", &errs); host != nil {
		hostsArr = append(hostsArr, *host)
	}
	appParams.Hosts = removeDuplicatedValue(hostsArr)

	appParams.Name = stringVal(yamlMap, "name", &errs)
	appParams.Path = stringVal(yamlMap, "path", &errs)
	appParams.StackName = stringVal(yamlMap, "stack", &errs)
	appParams.Command = stringValOrDefault(yamlMap, "command", &errs)
	appParams.Memory = bytesVal(yamlMap, "memory", &errs)
	appParams.InstanceCount = intVal(yamlMap, "instances", &errs)
	appParams.HealthCheckTimeout = intVal(yamlMap, "timeout", &errs)
	appParams.NoRoute = boolVal(yamlMap, "no-route", &errs)
	appParams.NoHostname = boolVal(yamlMap, "no-hostname", &errs)
	appParams.UseRandomHostname = boolVal(yamlMap, "random-route", &errs)
	appParams.ServicesToBind = sliceOrEmptyVal(yamlMap, "services", &errs)
	appParams.EnvironmentVars = envVarOrEmptyMap(yamlMap, &errs)
	appParams.HealthCheckType = stringVal(yamlMap, "health-check-type", &errs)

	if appParams.Path != nil {
		path := *appParams.Path
		if filepath.IsAbs(path) {
			path = filepath.Clean(path)
		} else {
			path = filepath.Join(basePath, path)
		}
		appParams.Path = &path
	}

	if len(errs) > 0 {
		message := ""
		for _, err := range errs {
			message = message + fmt.Sprintf("%s\n", err.Error())
		}
		return models.AppParams{}, errors.New(message)
	}

	return appParams, nil
}

func removeDuplicatedValue(ary []string) *[]string {
	if ary == nil {
		return nil
	}

	m := make(map[string]bool)
	for _, v := range ary {
		m[v] = true
	}

	newAry := []string{}
	for k := range m {
		newAry = append(newAry, k)
	}
	return &newAry
}

func checkForNulls(yamlMap generic.Map) error {
	var errs []error
	generic.Each(yamlMap, func(key interface{}, value interface{}) {
		if key == "command" || key == "buildpack" {
			return
		}
		if value == nil {
			errs = append(errs, fmt.Errorf(T("{{.PropertyName}} should not be null", map[string]interface{}{"PropertyName": key})))
		}
	})

	if len(errs) > 0 {
		message := ""
		for i := range errs {
			message = message + fmt.Sprintf("%s\n", errs[i].Error())
		}
		return errors.New(message)
	}

	return nil
}

func stringVal(yamlMap generic.Map, key string, errs *[]error) *string {
	val := yamlMap.Get(key)
	if val == nil {
		return nil
	}
	result, ok := val.(string)
	if !ok {
		*errs = append(*errs, fmt.Errorf(T("{{.PropertyName}} must be a string value", map[string]interface{}{"PropertyName": key})))
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
		}
		return &val
	case nil:
		return &empty
	default:
		*errs = append(*errs, fmt.Errorf(T("{{.PropertyName}} must be a string or null value", map[string]interface{}{"PropertyName": key})))
		return nil
	}
}

func bytesVal(yamlMap generic.Map, key string, errs *[]error) *int64 {
	yamlVal := yamlMap.Get(key)
	if yamlVal == nil {
		return nil
	}

	stringVal := coerceToString(yamlVal)
	value, err := formatters.ToMegabytes(stringVal)
	if err != nil {
		*errs = append(*errs, fmt.Errorf(T("Invalid value for '{{.PropertyName}}': {{.StringVal}}\n{{.Error}}",
			map[string]interface{}{
				"PropertyName": key,
				"Error":        err.Error(),
				"StringVal":    stringVal,
			})))
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
		err = fmt.Errorf(T("Expected {{.PropertyName}} to be a number, but it was a {{.PropertyType}}.",
			map[string]interface{}{"PropertyName": key, "PropertyType": val}))
	}

	if err != nil {
		*errs = append(*errs, err)
		return nil
	}

	return &intVal
}

func coerceToString(value interface{}) string {
	return fmt.Sprintf("%v", value)
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
		*errs = append(*errs, fmt.Errorf(T("Expected {{.PropertyName}} to be a boolean.", map[string]interface{}{"PropertyName": key})))
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

	sliceErr := fmt.Errorf(T("Expected {{.PropertyName}} to be a list of strings.", map[string]interface{}{"PropertyName": key}))

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
		return &[]string{}
	}

	return &stringSlice
}

func envVarOrEmptyMap(yamlMap generic.Map, errs *[]error) *map[string]interface{} {
	key := "env"
	switch envVars := yamlMap.Get(key).(type) {
	case nil:
		aMap := make(map[string]interface{}, 0)
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

		result := make(map[string]interface{}, envVars.Count())
		generic.Each(envVars, func(key, value interface{}) {
			result[key.(string)] = value
		})

		return &result
	default:
		*errs = append(*errs, fmt.Errorf(T("Expected {{.Name}} to be a set of key => value, but it was a {{.Type}}.",
			map[string]interface{}{"Name": key, "Type": envVars})))
		return nil
	}
}

func validateEnvVars(input generic.Map) (errs []error) {
	generic.Each(input, func(key, value interface{}) {
		if value == nil {
			errs = append(errs, fmt.Errorf(T("env var '{{.PropertyName}}' should not be null",
				map[string]interface{}{"PropertyName": key})))
		}
	})
	return
}
