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
	m, _ = NewManifest(generic.NewMap())
	return m
}

func NewManifest(data generic.Map) (m *Manifest, errs ManifestErrors) {
	m = &Manifest{}
	m.data = data

	components, errs := newManifestComponents(data)
	if len(errs) > 0 {
		return
	}

	m.Applications = components.Applications

	for _, app := range m.Applications {
		localEnv := generic.NewMap(app.Get("env"))
		localServices := app.Get("services").([]string)

		app.Set("env", generic.Merge(components.GlobalEnvVars, localEnv))
		app.Set("services", mergeSets(components.GlobalServices, localServices))
	}

	return
}
