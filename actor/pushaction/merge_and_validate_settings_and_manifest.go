package pushaction

import (
	"net/url"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/util/manifest"
	log "github.com/sirupsen/logrus"
)

func (actor Actor) MergeAndValidateSettingsAndManifests(cmdLineSettings CommandLineSettings, apps []manifest.Application) ([]manifest.Application, error) {
	var mergedApps []manifest.Application

	if len(apps) == 0 {
		log.Info("no manifest, generating one from command line settings")
		mergedApps = append(mergedApps, cmdLineSettings.OverrideManifestSettings(manifest.Application{}))
	} else {
		if cmdLineSettings.Name != "" && len(apps) > 1 {
			var err error
			apps, err = actor.selectApp(cmdLineSettings.Name, apps)
			if err != nil {
				return nil, err
			}
		}
		err := actor.validateCommandLineSettingsAndManifestCombinations(cmdLineSettings, apps)
		if err != nil {
			return nil, err
		}

		for _, app := range apps {
			mergedApps = append(mergedApps, cmdLineSettings.OverrideManifestSettings(app))
		}
	}

	mergedApps, err := actor.sanitizeAppPath(mergedApps)
	if err != nil {
		return nil, err
	}
	mergedApps = actor.setSaneEndpoint(mergedApps)

	log.Debugf("merged app settings: %#v", mergedApps)

	err = actor.validateMergedSettings(mergedApps)
	if err != nil {
		log.Errorln("validation error post merge:", err)
		return nil, err
	}
	return mergedApps, nil
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

func (Actor) setSaneEndpoint(apps []manifest.Application) []manifest.Application {
	for i, app := range apps {
		if app.HealthCheckType == "http" && app.HealthCheckHTTPEndpoint == "" {
			apps[i].HealthCheckHTTPEndpoint = "/"
		}
	}

	return apps
}

func (Actor) sanitizeAppPath(apps []manifest.Application) ([]manifest.Application, error) {
	for i, app := range apps {
		if app.Path != "" {
			var err error
			apps[i].Path, err = filepath.Abs(app.Path)
			if err != nil {
				return nil, err
			}
		}
	}

	return apps, nil
}

func (Actor) validateCommandLineSettingsAndManifestCombinations(cmdLineSettings CommandLineSettings, apps []manifest.Application) error {
	if len(apps) > 1 {
		switch {
		case
			cmdLineSettings.Buildpacks != nil,
			cmdLineSettings.Command.IsSet,
			cmdLineSettings.DefaultRouteDomain != "",
			cmdLineSettings.DefaultRouteHostname != "",
			cmdLineSettings.DiskQuota != 0,
			cmdLineSettings.DockerImage != "",
			cmdLineSettings.DockerUsername != "",
			cmdLineSettings.DropletPath != "",
			cmdLineSettings.HealthCheckTimeout != 0,
			cmdLineSettings.HealthCheckType != "",
			cmdLineSettings.Instances.IsSet,
			cmdLineSettings.Memory != 0,
			cmdLineSettings.NoHostname,
			cmdLineSettings.NoRoute,
			cmdLineSettings.ProvidedAppPath != "",
			cmdLineSettings.RandomRoute,
			cmdLineSettings.RoutePath != "",
			cmdLineSettings.StackName != "":
			log.Error("cannot use some parameters with multiple apps")
			return actionerror.CommandLineOptionsWithMultipleAppsError{}
		}
	}

	for _, app := range apps {
		switch {
		case app.NoRoute && len(app.Routes) > 0:
			return actionerror.PropertyCombinationError{AppName: app.Name, Properties: []string{"no-route", "routes"}}
		case app.DeprecatedDomain != nil ||
			app.DeprecatedDomains != nil ||
			app.DeprecatedHost != nil ||
			app.DeprecatedHosts != nil ||
			app.DeprecatedNoHostname != nil:

			deprecatedFields := []string{}
			if app.DeprecatedDomain != nil {
				deprecatedFields = append(deprecatedFields, "domain")
			}
			if app.DeprecatedDomains != nil {
				deprecatedFields = append(deprecatedFields, "domains")
			}
			if app.DeprecatedHost != nil {
				deprecatedFields = append(deprecatedFields, "host")
			}
			if app.DeprecatedHosts != nil {
				deprecatedFields = append(deprecatedFields, "hosts")
			}
			if app.DeprecatedNoHostname != nil {
				deprecatedFields = append(deprecatedFields, "no-hostname")
			}
			return actionerror.TriggerLegacyPushError{DomainHostRelated: deprecatedFields}
		case len(app.Routes) > 0:
			commandLineOptionsAndManifestConflictErr := actionerror.CommandLineOptionsAndManifestConflictError{
				ManifestAttribute:  "route",
				CommandLineOptions: []string{"-d", "--hostname", "-n", "--no-hostname", "--route-path"},
			}
			if cmdLineSettings.DefaultRouteDomain != "" ||
				cmdLineSettings.DefaultRouteHostname != "" ||
				cmdLineSettings.NoHostname != false ||
				cmdLineSettings.RoutePath != "" {
				return commandLineOptionsAndManifestConflictErr
			}
		}
	}

	return nil
}

func (actor Actor) validateMergedSettings(apps []manifest.Application) error {
	for i, app := range apps {
		log.WithField("index", i).Info("validating app")
		if app.Name == "" {
			log.WithField("index", i).Error("does not contain an app name")
			return actionerror.MissingNameError{}
		}

		for _, route := range app.Routes {
			err := actor.validateRoute(route)
			if err != nil {
				return err
			}
		}

		if app.DockerImage != "" {
			if app.DockerUsername != "" && app.DockerPassword == "" {
				log.WithField("app", app.Name).Error("no docker password found")
				return actionerror.DockerPasswordNotSetError{}
			}
			if app.Buildpack.IsSet {
				return actionerror.PropertyCombinationError{AppName: app.Name, Properties: []string{"docker", "buildpack"}}
			}
			if app.Buildpacks != nil {
				return actionerror.PropertyCombinationError{AppName: app.Name, Properties: []string{"docker", "buildpacks"}}
			}
			if app.Path != "" {
				return actionerror.PropertyCombinationError{AppName: app.Name, Properties: []string{"docker", "path"}}
			}
			if app.DropletPath != "" {
				return actionerror.PropertyCombinationError{AppName: app.Name, Properties: []string{"docker", "droplet"}}
			}
		}

		if app.DropletPath != "" {
			if app.Path != "" {
				return actionerror.PropertyCombinationError{AppName: app.Name, Properties: []string{"droplet", "path"}}
			}
			if app.Buildpack.IsSet {
				return actionerror.PropertyCombinationError{AppName: app.Name, Properties: []string{"droplet", "buildpack"}}
			}
			if app.Buildpacks != nil {
				return actionerror.PropertyCombinationError{AppName: app.Name, Properties: []string{"droplet", "buildpacks"}}
			}
		}

		if app.DockerImage == "" && app.DropletPath == "" {
			_, err := os.Stat(app.Path)
			if os.IsNotExist(err) {
				log.WithField("path", app.Path).Error("app path does not exist")
				return actionerror.NonexistentAppPathError{Path: app.Path}
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

		if app.Buildpacks != nil && app.Buildpack.IsSet {
			return actionerror.PropertyCombinationError{AppName: app.Name, Properties: []string{"buildpack", "buildpacks"}}
		}

		if len(app.Buildpacks) > 1 {
			for _, b := range app.Buildpacks {
				if b == "null" || b == "default" {
					return actionerror.InvalidBuildpacksError{}
				}
			}
		}
	}
	return nil
}

func (actor Actor) validateRoute(route string) error {
	_, err := url.Parse(route)
	if err != nil || !actor.urlValidator.MatchString(route) {
		return actionerror.InvalidRouteError{Route: route}
	}

	return nil
}
