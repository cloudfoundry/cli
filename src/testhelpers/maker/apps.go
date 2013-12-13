package maker

import (
	"cf"
)

var appGuid func () string

func init() {
	appGuid = guidGenerator("app")
}

func NewApp(overrides Overrides) (app cf.Application) {
	app.Name = "app-name"
	app.Guid = appGuid()
	app.State = "started"

	guid, ok := overrides["guid"]
	if ok {
		app.Guid = guid.(string)
	}

	name, ok := overrides["name"]
	if ok {
		app.Name = name.(string)
	}

	state, ok := overrides["state"]
	if ok {
		app.State = state.(string)
	}

	return
}

