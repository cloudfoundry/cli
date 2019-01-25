package v2v3action

import (
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/util/manifest"
)

type ManifestV2Actor interface {
	CreateApplicationManifestByNameAndSpace(string, string) (manifest.Application, v2action.Warnings, error)
}

type ManifestV3Actor interface {
	GetApplicationByNameAndSpace(string, string) (v3action.Application, v3action.Warnings, error)
}

func (actor *Actor) CreateApplicationManifestByNameAndSpace(appName string, appSpace string) (manifest.Application, Warnings, error) {
	var allWarnings Warnings

	manifestApp, v2warnings, err := actor.V2Actor.CreateApplicationManifestByNameAndSpace(appName, appSpace)
	allWarnings = append(allWarnings, v2warnings...)
	if err != nil {
		return manifest.Application{}, allWarnings, err
	}

	v3App, v3warnings, v3Err := actor.V3Actor.GetApplicationByNameAndSpace(appName, appSpace)
	allWarnings = append(allWarnings, v3warnings...)
	if v3Err != nil {
		return manifest.Application{}, allWarnings, v3Err
	}

	manifestApp.Buildpacks = v3App.LifecycleBuildpacks

	return manifestApp, allWarnings, err
}

func (Actor) WriteApplicationManifest(manifestApp manifest.Application, manifestPath string) error {
	return manifest.WriteApplicationManifest(manifestApp, manifestPath)
}
