package pushaction

import (
	"code.cloudfoundry.org/cli/actor/pushaction/manifest"
	log "github.com/sirupsen/logrus"
)

type MissingNameError struct{}

func (MissingNameError) Error() string {
	return "name not specified for app"
}

func (actor Actor) MergeAndValidateSettingsAndManifests(settings CommandLineSettings, apps []manifest.Application) ([]manifest.Application, error) {
	var mergedApps []manifest.Application

	if len(apps) == 0 {
		mergedApps = actor.generateAppSettingsFromCommandLineSettings(settings)
	} else {
		// validate premerged settings
		for _, app := range apps {
			mergedApps = append(mergedApps, actor.mergeCommandLineSettingsAndManifest(settings, app))
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

func (Actor) mergeCommandLineSettingsAndManifest(settings CommandLineSettings, app manifest.Application) manifest.Application {
	app.Path = settings.ApplicationPath()
	return app
}

func (Actor) validateMergedSettings(apps []manifest.Application) error {
	for i, app := range apps {
		if app.Name == "" {
			log.WithField("index", i).Error("does not contain an app name")
			return MissingNameError{}
		}
	}
	return nil
}
