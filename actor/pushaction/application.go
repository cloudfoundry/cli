package pushaction

import (
	"fmt"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/types"
	log "github.com/sirupsen/logrus"
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

func (actor Actor) CreateApplication(config ApplicationConfig) (ApplicationConfig, Event, Warnings, error) {
	log.Debug("creating application")
	v2App := config.DesiredApplication.Application
	v2App.Buildpack = actor.setBuildpack(config)

	log.WithField("application", v2App).Debug("creating application with these settings")
	var warnings Warnings
	newApp, v2warnings, err := actor.V2Actor.CreateApplication(v2App)
	warnings = append(warnings, v2warnings...)
	if err != nil {
		log.Errorln("creating application:", err)
		return ApplicationConfig{}, "", Warnings(warnings), err
	}

	if config.HasMultipleBuildpacks() {
		v3Warnings, v3Err := actor.updateBuildpacks(config, newApp)
		warnings = append(warnings, v3Warnings...)
		if v3Err != nil {
			return ApplicationConfig{}, "", warnings, v3Err
		}
	}

	config.DesiredApplication.Application = newApp
	config.CurrentApplication = config.DesiredApplication

	return config, CreatedApplication, Warnings(warnings), nil
}

func (actor Actor) UpdateApplication(config ApplicationConfig) (ApplicationConfig, Event, Warnings, error) {
	log.Debug("updating application")
	v2App := config.DesiredApplication.Application
	v2App.Buildpack = actor.setBuildpack(config)

	v2App = actor.ignoreSameState(config, v2App)
	v2App = actor.ignoreSameStackGUID(config, v2App)

	log.WithField("application", v2App).Debug("updating application with these settings")
	var warnings Warnings
	updatedApp, v2Warnings, err := actor.V2Actor.UpdateApplication(v2App)
	warnings = append(warnings, v2Warnings...)
	if err != nil {
		log.Errorln("updating application:", err)
		return ApplicationConfig{}, "", Warnings(warnings), err
	}

	if config.HasMultipleBuildpacks() {
		v3Warnings, v3Err := actor.updateBuildpacks(config, updatedApp)
		warnings = append(warnings, v3Warnings...)
		if v3Err != nil {
			return ApplicationConfig{}, "", warnings, v3Err
		}
	}

	config.DesiredApplication.Application = updatedApp
	config.CurrentApplication = config.DesiredApplication

	return config, UpdatedApplication, Warnings(warnings), err
}

func (actor Actor) FindOrReturnPartialApp(appName string, spaceGUID string) (bool, Application, v2action.Warnings, error) {
	foundApp, appWarnings, err := actor.V2Actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if _, ok := err.(actionerror.ApplicationNotFoundError); ok {
		log.Warnf("unable to find app %s in current space (GUID: %s)", appName, spaceGUID)
		return false, Application{
			Application: v2action.Application{
				Name:      appName,
				SpaceGUID: spaceGUID,
			},
		}, appWarnings, nil
	} else if err != nil {
		log.WithField("appName", appName).Error("error retrieving app")
		return false, Application{}, appWarnings, err
	}

	stack, stackWarnings, err := actor.V2Actor.GetStack(foundApp.StackGUID)
	warnings := append(appWarnings, stackWarnings...)
	if err != nil {
		log.Warnf("unable to find app's stack (GUID: %s)", foundApp.StackGUID)
		return false, Application{}, warnings, err
	}

	v3App, v3AppWarnings, err := actor.V3Actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	warnings = append(warnings, v3AppWarnings...)
	if err != nil {
		if _, ok := err.(ccerror.APINotFoundError); !ok {
			return false, Application{}, warnings, err
		}
	}

	// if err api don't exist - ignore, else return error
	app := Application{
		Application: foundApp,
		Stack:       stack,
		Buildpacks:  v3App.LifecycleBuildpacks,
	}
	return true, app, warnings, nil
}

// For some versions of CC, sending state will always result in CC
// attempting to do perform that request (i.e. started -> start/restart).
// In order to prevent repeated unintended restarts in the middle of a
// push, don't send state. This will be fixed in capi-release 1.48.0.
func (actor Actor) ignoreSameState(config ApplicationConfig, v2App v2action.Application) v2action.Application {
	if config.CurrentApplication.State == config.DesiredApplication.State {
		v2App.State = ""
	}

	return v2App
}

// Apps updates with both docker image and stack guids fail. So do not send
// StackGUID unless it is necessary.
func (actor Actor) ignoreSameStackGUID(config ApplicationConfig, v2App v2action.Application) v2action.Application {
	if config.CurrentApplication.StackGUID == config.DesiredApplication.StackGUID {
		v2App.StackGUID = ""
	}

	return v2App
}

// If 'buildpacks' is set with only one buildpack, set `buildpack` (singular)
// on the application.
func (Actor) setBuildpack(config ApplicationConfig) types.FilteredString {
	buildpacks := config.DesiredApplication.Buildpacks

	if len(buildpacks) == 1 {
		var filtered types.FilteredString
		filtered.ParseValue(buildpacks[0])
		return filtered
	}

	if buildpacks != nil && len(buildpacks) == 0 {
		return types.FilteredString{IsSet: true}
	}

	return config.DesiredApplication.Buildpack
}

func (actor Actor) updateBuildpacks(config ApplicationConfig, v2App v2action.Application) (Warnings, error) {
	log.WithField("buildpacks", config.DesiredApplication.Buildpacks).Debug("updating with multiple buildpacks")
	v3App := v3action.Application{
		Name:                v2App.Name,
		GUID:                v2App.GUID,
		StackName:           config.DesiredApplication.Stack.Name,
		LifecycleBuildpacks: config.DesiredApplication.Buildpacks,
		LifecycleType:       constant.AppLifecycleTypeBuildpack,
	}

	_, warnings, err := actor.V3Actor.UpdateApplication(v3App)
	return Warnings(warnings), err
}
