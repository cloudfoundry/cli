package pushaction

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func (actor Actor) CreateArchive(config ApplicationConfig) (string, error) {
	log.Info("creating archive")

	var archivePath string
	var err error

	//change to look at unmatched
	if config.Archive {
		archivePath, err = actor.V2Actor.ZipArchiveResources(config.Path, config.UnmatchedResources)
	} else {
		archivePath, err = actor.V2Actor.ZipDirectoryResources(config.Path, config.UnmatchedResources)
	}
	if err != nil {
		log.WithField("path", config.Path).Errorln("archiving resources:", err)
		return "", err
	}
	log.WithField("archivePath", archivePath).Debug("archive created")
	return archivePath, nil
}

func (actor Actor) SetMatchedResources(config ApplicationConfig) (ApplicationConfig, Warnings) {
	matched, unmatched, warnings, err := actor.V2Actor.ResourceMatch(config.AllResources)

	if err != nil {
		log.Error("uploading all resources instead of resource matching")
		config.UnmatchedResources = config.AllResources
		return config, Warnings(warnings)
	}

	config.MatchedResources = matched
	config.UnmatchedResources = unmatched

	return config, Warnings(warnings)
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

	log.WithFields(log.Fields{
		"appGUID":     config.DesiredApplication.GUID,
		"archiveSize": archiveInfo.Size(),
	}).Debug("uploading app bits")

	eventStream <- UploadingApplication
	reader := progressbar.NewProgressBarWrapper(archive, archiveInfo.Size())

	var allWarnings Warnings
	// change to look at matched resoruces
	job, warnings, err := actor.V2Actor.UploadApplicationPackage(config.DesiredApplication.GUID, config.MatchedResources, reader, archiveInfo.Size())
	allWarnings = append(allWarnings, Warnings(warnings)...)

	if err != nil {
		log.WithField("archivePath", archivePath).Errorln("streaming archive:", err)
		return allWarnings, err
	}
	eventStream <- UploadComplete
	warnings, err = actor.V2Actor.PollJob(job)
	allWarnings = append(allWarnings, Warnings(warnings)...)

	return allWarnings, err
}
