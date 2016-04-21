package maker

import "github.com/cloudfoundry/cli/cf/models"

var appGUID func() string

func init() {
	appGUID = guidGenerator("app")
}

func NewAppFields(overrides Overrides) (app models.ApplicationFields) {
	app.Name = "app-name"
	app.GUID = appGUID()
	app.State = "started"

	if overrides.Has("GUID") {
		app.GUID = overrides.Get("GUID").(string)
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
