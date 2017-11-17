package pushaction

import (
	"os"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/util/manifest"
	log "github.com/sirupsen/logrus"
)

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
		return nil, actionerror.AppNotFoundInManifestError{Name: appName}
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
			settings.RoutePath != "",
			settings.StackName != "":
			log.Error("cannot use some parameters with multiple apps")
			return actionerror.CommandLineOptionsWithMultipleAppsError{}
		}
	}

	for _, app := range apps {
		switch {
		case app.NoRoute && len(app.Routes) > 0:
			return actionerror.PropertyCombinationError{AppName: app.Name, Properties: []string{"no-route", "routes"}}
		case app.DeprecatedDomain != nil || app.DeprecatedDomains != nil:
			return actionerror.TriggerLegacyPushError{DomainRelated: true}
		case app.DeprecatedHost != nil || app.DeprecatedHosts != nil || app.DeprecatedNoHostname != nil:
			return actionerror.TriggerLegacyPushError{HostnameRelated: true}
		case len(app.Routes) > 0:
			commandLineOptionsAndManifestConflictErr := actionerror.CommandLineOptionsAndManifestConflictError{
				ManifestAttribute:  "route",
				CommandLineOptions: []string{"-d", "--hostname", "-n", "--no-hostname", "--route-path"},
			}
			if settings.DefaultRouteDomain != "" ||
				settings.DefaultRouteHostname != "" ||
				settings.NoHostname != false ||
				settings.RoutePath != "" {
				return commandLineOptionsAndManifestConflictErr
			}
		}
	}

	return nil
}

func (Actor) validateMergedSettings(apps []manifest.Application) error {
	for i, app := range apps {
		if app.Name == "" {
			log.WithField("index", i).Error("does not contain an app name")
			return actionerror.MissingNameError{}
		}

		if app.DockerImage == "" {
			_, err := os.Stat(app.Path)
			if os.IsNotExist(err) {
				log.WithField("path", app.Path).Error("app path does not exist")
				return actionerror.NonexistentAppPathError{Path: app.Path}
			}
		} else {
			if app.DockerUsername != "" && app.DockerPassword == "" {
				log.WithField("app", app.Name).Error("no docker password found")
				return actionerror.DockerPasswordNotSetError{}
			}
			if app.Buildpack.IsSet {
				return actionerror.PropertyCombinationError{AppName: app.Name, Properties: []string{"docker", "buildpack"}}
			}
			if app.Path != "" {
				return actionerror.PropertyCombinationError{AppName: app.Name, Properties: []string{"docker", "path"}}
			}
		}

		if app.NoRoute {
			if app.Hostname != "" {
				return actionerror.PropertyCombinationError{AppName: app.Name, Properties: []string{"hostname", "no-route"}}
			}
			if app.NoHostname {
				return actionerror.PropertyCombinationError{AppName: app.Name, Properties: []string{"no-hostname", "no-route"}}
			}
			if app.RoutePath != "" {
				return actionerror.PropertyCombinationError{AppName: app.Name, Properties: []string{"route-path", "no-route"}}
			}
		}

		if app.HealthCheckHTTPEndpoint != "" && app.HealthCheckType != "http" {
			return actionerror.HTTPHealthCheckInvalidError{}
		}
	}
	return nil
}
