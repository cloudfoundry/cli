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

	return
}
