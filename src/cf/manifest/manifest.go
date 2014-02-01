package manifest

import (
	"cf"
	"cf/formatters"
	"errors"
	"fmt"
	"generic"
	"path/filepath"
	"regexp"
	"strconv"
)

var manifestKeys = map[string]func(appParams, yamlMap generic.Map, key string, errs *ManifestErrors){
	"buildpack":  setStringVal,
	"disk_quota": setStringVal,
	"domain":     setStringVal,
	"host":       setStringVal,
	"name":       setStringVal,
	"path":       setStringVal,
	"stack":      setStringVal,
	"command":    setStringOrNullVal,
	"memory":     setBytesVal,
	"instances":  setIntVal,
	"timeout":    setTimeoutVal,
	"no-route":   setBoolVal,
	"services":   setSliceOrEmptyVal,
	"env":        setEnvVarOrEmptyMap,
}

type Manifest struct {
	Applications cf.AppSet
}

func NewEmptyManifest() (m *Manifest) {
	m, _ = NewManifest("", generic.NewMap())
	return m
}

func NewManifest(basePath string, data generic.Map) (m *Manifest, errs ManifestErrors) {
	errs = walkManifestLookingForProperties(data)
	if !errs.Empty() {
		return
	}

	m = &Manifest{}
	m.Applications, errs = mapToAppSet(basePath, data)
	return
}

func walkManifestLookingForProperties(data generic.Map) (errs ManifestErrors) {
	generic.Each(data, func(key, value interface{}) {
		errs = append(errs, walkMapLookingForProperties(value)...)
	})

	return
}

func walkMapLookingForProperties(value interface{}) (errs ManifestErrors) {
	propertyRegex := regexp.MustCompile(`\${\w+}`)
	switch value := value.(type) {
	case string:
		match := propertyRegex.FindString(value)
		if match != "" {
			err := errors.New(fmt.Sprintf("Properties are not supported. Found property '%s' in manifest", match))
			errs = append(errs, err)
		}
	case []interface{}:
		for _, item := range value {
			errs = append(errs, walkMapLookingForProperties(item)...)
		}
	case map[string]interface{}:
		for _, item := range value {
			errs = append(errs, walkMapLookingForProperties(item)...)
		}
	}

	return
}

func mapToAppSet(basePath string, data generic.Map) (appSet cf.AppSet, errs ManifestErrors) {
	appSet = make([]cf.AppParams, 0)

	if data.Has("applications") {
		appMaps, ok := data.Get("applications").([]interface{})
		if !ok {
			errs = append(errs, errors.New("Expected applications to be a list"))
			return
		}

		// we delete applications so that we may merge top level app params into each app
		data.Delete("applications")

		for _, appData := range appMaps {
			if !generic.IsMappable(appData) {
				errs = append(errs, errors.New("Expected application to be a dictionary"))
				continue
			}

			appMap := generic.DeepMerge(data, generic.NewMap(appData))

			appParams, appErrs := mapToAppParams(basePath, appMap)
			if !appErrs.Empty() {
				errs = append(errs, appErrs)
				continue
			}

			appSet = append(appSet, appParams)
		}
	}

	return
}

func mapToAppParams(basePath string, yamlMap generic.Map) (appParams cf.AppParams, errs ManifestErrors) {
	appParams = cf.NewEmptyAppParams()

	errs = checkForNulls(yamlMap)
	if !errs.Empty() {
		return
	}

	for key, handler := range manifestKeys {
		if yamlMap.Has(key) {
			handler(appParams, yamlMap, key, &errs)
		}
	}

	if appParams.Has("path") {
		path := appParams.Get("path").(string)
		if filepath.IsAbs(path) {
			path = filepath.Clean(path)
		} else {
			path = filepath.Join(basePath, path)
		}
		appParams.Set("path", path)
	}

	return
}

func checkForNulls(appParams generic.Map) (errs ManifestErrors) {
	for key, _ := range manifestKeys {
		if key == "command" {
			continue
		}
		if appParams.IsNil(key) {
			errs = append(errs, errors.New(fmt.Sprintf("%s should not be null", key)))
		}
	}

	return
}

func setStringVal(appMap generic.Map, yamlMap generic.Map, key string, errs *ManifestErrors) {
	val := yamlMap.Get(key)
	stringVal, ok := val.(string)
	if !ok {
		*errs = append(*errs, errors.New(fmt.Sprintf("%s must be a string value", key)))
		return
	}
	appMap.Set(key, stringVal)
}

func setStringOrNullVal(appMap, yamlMap generic.Map, key string, errs *ManifestErrors) {
	switch val := yamlMap.Get(key).(type) {
	case string:
		appMap.Set(key, val)
	case nil:
		appMap.Set(key, "")
	default:
		*errs = append(*errs, errors.New(fmt.Sprintf("%s must be a string or null value", key)))
	}
}

func setBytesVal(appMap, yamlMap generic.Map, key string, errs *ManifestErrors) {
	value, err := formatters.ToMegabytes(yamlMap.Get(key).(string))
	if err != nil {
		*errs = append(*errs, errors.New(fmt.Sprintf("Unexpected value for %s :\n%s", key, err.Error())))
		return
	}
	appMap.Set(key, value)
}

func setIntVal(appMap, yamlMap generic.Map, key string, errs *ManifestErrors) {
	var (
		intVal int
		err    error
	)

	switch val := yamlMap.Get(key).(type) {
	case string:
		intVal, err = strconv.Atoi(val)
	case int:
		intVal = val
	default:
		err = errors.New(fmt.Sprintf("Expected %s to be a number.", key))
	}

	if err != nil {
		*errs = append(*errs, err)
		return
	}

	appMap.Set(key, intVal)
}

func setTimeoutVal(appMap, yamlMap generic.Map, key string, errs *ManifestErrors) {
	var (
		intVal int
		err    error
	)

	switch val := yamlMap.Get(key).(type) {
	case string:
		intVal, err = strconv.Atoi(val)
	case int:
		intVal = val
	default:
		err = errors.New("Expected health_check_timeout to be a number.")
	}

	if err != nil {
		*errs = append(*errs, err)
		return
	}

	appMap.Set("health_check_timeout", intVal)
}

func setBoolVal(appMap, yamlMap generic.Map, key string, errs *ManifestErrors) {
	switch val := yamlMap.Get(key).(type) {
	case bool:
		appMap.Set(key, val)
	case string:
		boolVal := val == "true"
		appMap.Set(key, boolVal)
	default:
		*errs = append(*errs, errors.New(fmt.Sprintf("Expected %s to be a boolean.", key)))
	}

	return
}

func setEnvVarOrEmptyMap(appMap, yamlMap generic.Map, key string, errs *ManifestErrors) {
	if !yamlMap.Has(key) {
		appMap.Set(key, generic.NewMap())
		return
	}

	envVars := yamlMap.Get(key)

	if !generic.IsMappable(envVars) {
		*errs = append(*errs, errors.New(fmt.Sprintf("Expected %s to be a set of key => value.", key)))
		return
	}

	merrs := validateEnvVars(envVars)
	if merrs != nil {
		*errs = append(*errs, merrs)
		return
	}

	appMap.Set(key, generic.NewMap(envVars))
}

func setSliceOrEmptyVal(appMap, yamlMap generic.Map, key string, errs *ManifestErrors) {
	if !yamlMap.Has(key) {
		appMap.Set(key, []string{})
		return
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
		return
	}

	appMap.Set(key, stringSlice)
	return
}

func validateEnvVars(input interface{}) (errs ManifestErrors) {
	envVars := generic.NewMap(input)
	generic.Each(envVars, func(key, value interface{}) {
		if value == nil {
			errs = append(errs, errors.New(fmt.Sprintf("env var '%s' should not be null", key)))
		}
	})
	return
}
