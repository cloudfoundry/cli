package manifest

import (
	"cf"
	"generic"
)

type Manifest struct {
	data         generic.Map
	Applications cf.AppSet
}

func NewEmptyManifest() (m *Manifest) {
	return NewManifest(generic.NewEmptyMap())
}

func NewManifest(data generic.Map) (m *Manifest) {
	m = &Manifest{}
	m.data = data
	if data.Has("applications") {
		m.Applications = cf.NewAppSet(data.Get("applications"))
	} else {
		m.Applications = cf.NewEmptyAppSet()
	}

	if data.Has("env") {
		globalEnv := generic.NewMap(data.Get("env"))
		for _, app := range m.Applications {
			localEnv := generic.NewMap(app.Get("env"))
			app.Set("env", generic.Merge(globalEnv, localEnv))
		}
	}

	var globalServices []string
	if data.Has("services") {
		global, hasGlobal := data.Get("services").([]interface{})
		if !hasGlobal {
			globalServices = []string{}
		} else {
			globalServices = interfaceSliceToString(global)
		}
	}

	for _, app := range m.Applications {
		locals, ok := app.Get("services").([]interface{})
		if ok {
			localServices := interfaceSliceToString(locals)
			app.Set("services", mergeSets(globalServices, localServices))
		} else {
			app.Set("services", globalServices)
		}
	}

	return
}

func interfaceSliceToString(set []interface{}) (result []string) {
	for _, value := range set {
		result = append(result, value.(string))
	}
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
