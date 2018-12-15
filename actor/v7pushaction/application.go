package v7pushaction

import (
	"fmt"

	"code.cloudfoundry.org/cli/actor/v2action"
)

type Application struct {
	v2action.Application
	Buildpacks []string
	Stack      v2action.Stack
}

// CalculatedBuildpacks will return back the buildpacks for the application.
func (app Application) CalculatedBuildpacks() []string {
	buildpack := app.CalculatedBuildpack()
	switch {
	case app.Buildpacks != nil:
		return app.Buildpacks
	case len(buildpack) > 0:
		return []string{buildpack}
	default:
		return nil
	}
}

func (app Application) String() string {
	return fmt.Sprintf("%s, Stack Name: '%s', Buildpacks: %s", app.Application, app.Stack.Name, app.Buildpacks)
}

func (app *Application) SetStack(stack v2action.Stack) {
	app.Stack = stack
	app.StackGUID = stack.GUID
}
