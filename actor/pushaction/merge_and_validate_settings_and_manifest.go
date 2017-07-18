package pushaction

import (
	"fmt"
	"os"

	"code.cloudfoundry.org/cli/actor/pushaction/manifest"
	log "github.com/sirupsen/logrus"
)

type MissingNameError struct{}

func (MissingNameError) Error() string {
	return "name not specified for app"
}

type NonexistentAppPathError struct {
	Path string
}

func (e NonexistentAppPathError) Error() string {
	return fmt.Sprint("app path not found:", e.Path)
}

type CommandLineOptionsWithMultipleAppsError struct{}

func (CommandLineOptionsWithMultipleAppsError) Error() string {
	return "cannot use command line flag with multiple apps"
}

type AppNotFoundInManifestError struct {
	Name string
}

func (e AppNotFoundInManifestError) Error() string {
	return fmt.Sprintf("specfied app: %s not found in manifest", e.Name)
}

func (actor Actor) MergeAndValidateSettingsAndManifests(settings CommandLineSettings, apps []manifest.Application) ([]manifest.Application, error) {
	var mergedApps []manifest.Application

	if len(apps) == 0 {
		mergedApps = actor.generateAppSettingsFromCommandLineSettings(settings)
	} else {
		err := actor.validatePremergedSettings(settings, apps)
		if err != nil {
			return nil, err
		}

		for _, app := range apps {
			mergedApps = append(mergedApps, settings.OverrideManifestSettings(app))
		}

		if settings.Name != "" {
			mergedApps, err = actor.selectApp(settings.Name, mergedApps)
			if err != nil {
				return nil, err
			}
		}
	}

	log.Debugf("merged app settings: %#v", mergedApps)
	return mergedApps, actor.validateMergedSettings(mergedApps)
}

func (Actor) generateAppSettingsFromCommandLineSettings(settings CommandLineSettings) []manifest.Application {
	log.Info("no manifest, generating one from command line settings")

	appSettings := []manifest.Application{{
		DockerImage: settings.DockerImage,
		Name:        settings.Name,
		Path:        settings.ApplicationPath(),
	}}

	return appSettings
}

func (Actor) selectApp(appName string, apps []manifest.Application) ([]manifest.Application, error) {
	var returnedApps []manifest.Application
	for _, app := range apps {
		if app.Name == appName {
			returnedApps = append(returnedApps, app)
		}
	}
	if len(returnedApps) == 0 {
		return nil, AppNotFoundInManifestError{Name: appName}
	}

	return returnedApps, nil
}

func (Actor) validatePremergedSettings(settings CommandLineSettings, apps []manifest.Application) error {
	if settings.ProvidedAppPath != "" && len(apps) > 1 {
		log.Error("cannot use -p with multiple apps")
		return CommandLineOptionsWithMultipleAppsError{}
	}
	return nil
}

func (Actor) validateMergedSettings(apps []manifest.Application) error {
	for i, app := range apps {
		if app.Name == "" {
			log.WithField("index", i).Error("does not contain an app name")
			return MissingNameError{}
		}
		_, err := os.Stat(app.Path)
		if os.IsNotExist(err) {
			log.WithField("path", app.Path).Error("app path does not exist")
			return NonexistentAppPathError{Path: app.Path}
		}
	}
	return nil
}
