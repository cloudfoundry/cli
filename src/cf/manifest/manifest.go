package manifest

import (
	"cf"
	"errors"
	"fmt"
	"generic"
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
	m = &Manifest{}
	m.Applications, errs = mapToAppSet(data)
	return
}

func mapToAppSet(data generic.Map) (appSet cf.AppSet, errs ManifestErrors) {
	appSet = cf.NewEmptyAppSet()

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
	appMap := generic.NewMap()

	errs = checkForNulls(yamlMap)
	if !errs.Empty() {
		return
	}

	for _, key := range []string{"buildpack", "command", "disk_quota", "domain", "host", "memory", "name", "path", "stack"} {
		if yamlMap.Has(key) {
			setStringVal(appMap, key, yamlMap.Get(key), &errs)
		}
	}

	if yamlMap.Has("timeout") {
		setIntVal(appMap, "health_check_timeout", yamlMap.Get("timeout"), &errs)
	}

	if yamlMap.Has("instances") {
		setIntVal(appMap, "instances", yamlMap.Get("instances"), &errs)
	}

	if yamlMap.Has("services") {
		setStringSlice(appMap, "services", yamlMap.Get("services"), &errs)
	} else {
		appMap.Set("services", []string{})
	}

	if yamlMap.Has("env") {
		setEnvVar(appMap, yamlMap.Get("env"), &errs)
	} else {
		appMap.Set("env", generic.NewMap())
	}

	appParams = cf.NewAppParams(appMap)
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

func mergeSets(set1, set2 []string) (result []string) {
	for _, aString := range set1 {
		result = append(result, aString)
	}

	for _, aString := range set2 {
		if !stringInSlice(aString, result) {
			result = append(result, aString)
		}
	}
	return
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
