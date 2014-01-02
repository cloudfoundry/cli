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

	if data.Has("services") {
		services := data.Get("services")
		if services != nil && len(services.([]interface{})) > 0 {
			globalServices := generic.NewMap(services.([]interface{})[0])

			for _, app := range m.Applications {
				localServices := generic.NewMap(app.Get("services"))
				app.Set("services", generic.Merge(globalServices, localServices))
			}
		}
	}

	return
}
