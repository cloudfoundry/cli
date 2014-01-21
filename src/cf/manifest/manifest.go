package manifest

import (
	"cf"
	"cf/formatters"
	"errors"
	"fmt"
	"generic"
	"regexp"
	"strconv"
)

var ManifestKeys = []string{
	"buildpack",
	"command",
	"disk_quota",
	"domain",
	"env",
	"host",
	"instances",
	"memory",
	"name",
	"path",
	"stack",
	"timeout",
}

type Manifest struct {
	Applications cf.AppSet
}

func NewEmptyManifest() (m *Manifest) {
	m, _ = NewManifest(generic.NewMap())
	return m
}

func NewManifest(data generic.Map) (m *Manifest, errs ManifestErrors) {
	errs = walkManifestLookingForProperties(data)
	if !errs.Empty() {
		return
	}

	m = &Manifest{}
	m.Applications, errs = mapToAppSet(data)
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

func mapToAppSet(data generic.Map) (appSet cf.AppSet, errs ManifestErrors) {
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

			appParams, appErrs := mapToAppParams(appMap)
			if !appErrs.Empty() {
				errs = append(errs, appErrs)
				continue
			}

			appSet = append(appSet, appParams)
		}
	}

	return
}

func mapToAppParams(yamlMap generic.Map) (appParams cf.AppParams, errs ManifestErrors) {
	appParams = cf.NewEmptyAppParams()

	errs = checkForNulls(yamlMap)
	if !errs.Empty() {
		return
	}

	for _, key := range []string{"buildpack", "disk_quota", "domain", "host", "name", "path", "stack"} {
		if yamlMap.Has(key) {
			setStringVal(appParams, key, yamlMap.Get(key), &errs)
		}
	}

	if yamlMap.Has("command") {
		setStringOrNullVal(appParams, "command", yamlMap.Get("command"), &errs)
	}

	if yamlMap.Has("memory") {
		memory, err := formatters.ToMegabytes(yamlMap.Get("memory").(string))
		if err != nil {
			errs = append(errs, errors.New(fmt.Sprintf("Unexpected value for app memory:\n%s", err.Error())))
			return
		}
		appParams.Set("memory", memory)
	}

	if yamlMap.Has("timeout") {
		setIntVal(appParams, "health_check_timeout", yamlMap.Get("timeout"), &errs)
	}

	if yamlMap.Has("instances") {
		setIntVal(appParams, "instances", yamlMap.Get("instances"), &errs)
	}

	if yamlMap.Has("services") {
		setStringSlice(appParams, "services", yamlMap.Get("services"), &errs)
	} else {
		appParams.Set("services", []string{})
	}

	if yamlMap.Has("env") {
		setEnvVar(appParams, yamlMap.Get("env"), &errs)
	} else {
		appParams.Set("env", generic.NewMap())
	}

	if yamlMap.Has("no-route") {
		setBoolVal(appParams, "no-route", yamlMap.Get("no-route"), &errs)
	}

	return
}

func checkForNulls(appParams generic.Map) (errs ManifestErrors) {
	for _, key := range ManifestKeys {
		if key == "command" {
			continue
		}
		if appParams.IsNil(key) {
			errs = append(errs, errors.New(fmt.Sprintf("%s should not be null", key)))
		}
	}

	return
}

func setStringVal(appMap generic.Map, key string, val interface{}, errs *ManifestErrors) {
	stringVal, ok := val.(string)
	if !ok {
		*errs = append(*errs, errors.New(fmt.Sprintf("%s must be a string value", key)))
		return
	}
	appMap.Set(key, stringVal)
}

func setStringOrNullVal(appMap generic.Map, key string, val interface{}, errs *ManifestErrors) {
	switch val := val.(type) {
	case string:
		appMap.Set(key, val)
	case nil:
		appMap.Set(key, "")
	default:
		*errs = append(*errs, errors.New(fmt.Sprintf("%s must be a string or null value", key)))
	}
}

func setIntVal(appMap generic.Map, key string, val interface{}, errs *ManifestErrors) {
	var (
		intVal int
		err    error
	)

	switch val := val.(type) {
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

func setBoolVal(appMap generic.Map, key string, val interface{}, errs *ManifestErrors) {
	switch val := val.(type) {
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

func setStringSlice(appMap generic.Map, key string, val interface{}, errs *ManifestErrors) {
	var (
		stringSlice []string
		err         error
	)

	errMsg := fmt.Sprintf("Expected %s to be a list of strings.", key)

	switch input := val.(type) {
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

func setEnvVar(appMap generic.Map, env interface{}, errs *ManifestErrors) {
	if !generic.IsMappable(env) {
		*errs = append(*errs, errors.New("Expected env vars to be a set of key => value."))
		return
	}

	merrs := validateEnvVars(env)
	if merrs != nil {
		*errs = append(*errs, merrs)
		return
	}

	appMap.Set("env", generic.NewMap(env))
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
