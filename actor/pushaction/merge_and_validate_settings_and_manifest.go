package pushaction

import (
	"fmt"
	"os"

	"code.cloudfoundry.org/cli/util/manifest"
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
		log.Info("no manifest, generating one from command line settings")
		mergedApps = append(mergedApps, settings.OverrideManifestSettings(manifest.Application{}))
	} else {
		if settings.Name != "" && len(apps) > 1 {
			var err error
			apps, err = actor.selectApp(settings.Name, apps)
			if err != nil {
				return nil, err
			}
		}

		err := actor.validatePremergedSettings(settings, apps)
		if err != nil {
			return nil, err
		}

		for _, app := range apps {
			mergedApps = append(mergedApps, settings.OverrideManifestSettings(app))
		}
	}

	mergedApps = actor.setSaneDefaults(mergedApps)

	log.Debugf("merged app settings: %#v", mergedApps)
	return mergedApps, actor.validateMergedSettings(mergedApps)
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

func (Actor) setSaneDefaults(apps []manifest.Application) []manifest.Application {
	for i, app := range apps {
		if app.HealthCheckType == "http" && app.HealthCheckHTTPEndpoint == "" {
			apps[i].HealthCheckHTTPEndpoint = "/"
		}
	}

	return apps
}

func (Actor) validatePremergedSettings(settings CommandLineSettings, apps []manifest.Application) error {
	if len(apps) > 1 {
		switch {
		case
			settings.Buildpack.IsSet,
			settings.Command.IsSet,
			settings.DiskQuota != 0,
			settings.DockerImage != "",
			settings.HealthCheckTimeout != 0,
			settings.HealthCheckType != "",
			settings.Instances.IsSet,
			settings.Memory != 0,
			settings.ProvidedAppPath != "",
			settings.StackName != "":
			log.Error("cannot use some parameters with multiple apps")
			return CommandLineOptionsWithMultipleAppsError{}
		}
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
