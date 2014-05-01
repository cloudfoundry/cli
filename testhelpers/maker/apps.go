package maker

import "github.com/cloudfoundry/cli/cf/models"

var appGuid func() string

func init() {
	appGuid = guidGenerator("app")
}

func NewAppFields(overrides Overrides) (app models.ApplicationFields) {
	app.Name = "app-name"
	app.Guid = appGuid()
	app.State = "started"

	if overrides.Has("Guid") {
		app.Guid = overrides.Get("Guid").(string)
	}

	if overrides.Has("Name") {
		app.Name = overrides.Get("Name").(string)
	}

	if overrides.Has("State") {
		app.State = overrides.Get("State").(string)
	}

	return
}

func NewApp(overrides Overrides) (app models.Application) {

	app.ApplicationFields = NewAppFields(overrides)

	return
}
