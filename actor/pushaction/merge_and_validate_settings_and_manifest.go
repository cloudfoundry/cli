package pushaction

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/pushaction/manifest"
	log "github.com/sirupsen/logrus"
)

func (_ Actor) MergeAndValidateSettingsAndManifests(cmdConfig CommandLineSettings, apps []manifest.Application) ([]manifest.Application, error) {
	if len(apps) != 0 {
		return nil, errors.New("functionality still pending")
	}

	path := cmdConfig.DirectoryPath
	if path == "" {
		path = cmdConfig.CurrentDirectory
	}

	manifests := []manifest.Application{{
		Name:        cmdConfig.Name,
		Path:        path,
		DockerImage: cmdConfig.DockerImage,
	}}

	//TODO Add validations

	log.Debugf("merged and validated manifests: %#v", manifests)
	return manifests, nil
}
