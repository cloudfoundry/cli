package pushaction

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/pushaction/manifest"
	log "github.com/Sirupsen/logrus"
)

func (_ Actor) MergeAndValidateSettingsAndManifests(cmdConfig CommandLineSettings, apps []manifest.Application) ([]manifest.Application, error) {
	if len(apps) != 0 {
		return nil, errors.New("functionality still pending")
	}
	manifests := []manifest.Application{{
		Name: cmdConfig.Name,
		Path: cmdConfig.CurrentDirectory,
	}}

	//TODO Add validations

	log.Debugf("merged and validated manifests: %#v", manifests)
	return manifests, nil
}
