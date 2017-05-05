package pushaction

import (
	"os"

	"code.cloudfoundry.org/cli/actor/v2action"
	log "github.com/Sirupsen/logrus"
)

func (actor Actor) CreateArchive(config ApplicationConfig) (string, error) {
	log.Info("creating archive")

	archivePath, err := actor.V2Actor.ZipResources(config.Path, config.AllResources)
	if err != nil {
		log.WithField("path", config.Path).Errorln("archiving resources:", err)
		return "", err
	}
	log.WithField("archivePath", archivePath).Debug("archive created")
	return archivePath, nil
}

func (actor Actor) UploadPackage(config ApplicationConfig, archivePath string, progressbar ProgressBar, eventStream chan<- Event) (Warnings, error) {
	log.Info("uploading archive")
	archive, err := os.Open(archivePath)
	if err != nil {
		log.WithField("archivePath", archivePath).Errorln("opening temp archive:", err)
		return nil, err
	}
	defer archive.Close()

	archiveInfo, err := archive.Stat()
	if err != nil {
		log.WithField("archivePath", archivePath).Errorln("stat temp archive:", err)
		return nil, err
	}

	//TODO Add progressbar wrapper

	log.WithFields(log.Fields{
		"appGUID":     config.DesiredApplication.GUID,
		"archiveSize": archiveInfo.Size(),
	}).Debug("uploading app bits")

	eventStream <- UploadingApplication
	reader := progressbar.NewProgressBarWrapper(archive, archiveInfo.Size())
	warnings, err := actor.V2Actor.UploadApplicationPackage(config.DesiredApplication.GUID, []v2action.Resource{}, reader, archiveInfo.Size())
	if err != nil {
		log.WithField("archivePath", archivePath).Errorln("streaming archive:", err)
		return Warnings(warnings), err
	}
	eventStream <- UploadComplete

	return Warnings(warnings), nil
}
