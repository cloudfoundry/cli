package maker

import "cf/models"

var appGuid func() string

func init() {
	appGuid = guidGenerator("app")
}

func NewAppFields(overrides Overrides) (app models.ApplicationFields) {
	app.Name = "app-name"
	app.Guid = appGuid()
	app.State = "started"
	app.InstanceCount = 42
	app.DiskQuota = 1073741824
	app.Memory = 268435456

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

func NewApp(overrides Overrides) (app models.Application) {

	app.ApplicationFields = NewAppFields(overrides)

	return
}
