package maker

import (
	"cf"
)

var appGuid func() string

func init() {
	appGuid = guidGenerator("app")
}

func NewAppFields(overrides Overrides) (app cf.ApplicationFields) {
	app.Name = "app-name"
	app.Guid = appGuid()
	app.State = "started"

	if overrides.Has("guid") {
		app.Guid = overrides.Get("guid").(string)
	}

	if overrides.Has("name") {
		app.Name = overrides.Get("name").(string)
	}

	if overrides.Has("state") {
		app.State = overrides.Get("state").(string)
	}

	return
}

func NewApp(overrides Overrides) (app cf.Application) {
	app.ApplicationFields = NewAppFields(overrides)

	return
}
