package pushaction

import (
	"os"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"

	log "github.com/sirupsen/logrus"
)

func (actor Actor) CreateArchive(config ApplicationConfig) (string, error) {
	log.Info("creating archive")

	var archivePath string
	var err error

	//change to look at unmatched
	if config.Archive {
		archivePath, err = actor.SharedActor.ZipArchiveResources(config.Path, actor.ConvertV2ResourcesToSharedResources(config.UnmatchedResources))
	} else {
		archivePath, err = actor.SharedActor.ZipDirectoryResources(config.Path, actor.ConvertV2ResourcesToSharedResources(config.UnmatchedResources))
	}
	if err != nil {
		log.WithField("path", config.Path).Errorln("archiving resources:", err)
		return "", err
	}
	log.WithField("archivePath", archivePath).Debug("archive created")
	return archivePath, nil
}

func (actor Actor) ConvertSharedResourcesToV2Resources(resources []sharedaction.Resource) []v2action.Resource {
	newResources := make([]v2action.Resource, 0, len(resources)) // Explicitly done to prevent nils

	for _, resource := range resources {
		newResources = append(newResources, v2action.Resource(resource))
	}
	return newResources
}

func (actor Actor) ConvertV2ResourcesToSharedResources(resources []v2action.Resource) []sharedaction.Resource {
	newResources := make([]sharedaction.Resource, 0, len(resources)) // Explicitly done to prevent nils

	for _, resource := range resources {
		newResources = append(newResources, sharedaction.Resource(resource))
	}
	return newResources
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

func (actor Actor) UploadPackage(config ApplicationConfig) (Warnings, error) {
	job, warnings, err := actor.V2Actor.UploadApplicationPackage(config.DesiredApplication.GUID, config.MatchedResources, nil, 0)
	if err != nil {
		return Warnings(warnings), err
	}

	pollWarnings, err := actor.V2Actor.PollJob(job)
	return append(Warnings(warnings), pollWarnings...), err
}

func (actor Actor) UploadPackageWithArchive(config ApplicationConfig, archivePath string, progressbar ProgressBar, eventStream chan<- Event) (Warnings, error) {
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

	eventStream <- UploadingApplicationWithArchive
	reader := progressbar.NewProgressBarWrapper(archive, archiveInfo.Size())

	var allWarnings Warnings
	// change to look at matched resoruces
	job, warnings, err := actor.V2Actor.UploadApplicationPackage(config.DesiredApplication.GUID, config.MatchedResources, reader, archiveInfo.Size())
	allWarnings = append(allWarnings, Warnings(warnings)...)

	if err != nil {
		log.WithField("archivePath", archivePath).Errorln("streaming archive:", err)
		return allWarnings, err
	}
	eventStream <- UploadWithArchiveComplete
	warnings, err = actor.V2Actor.PollJob(job)
	allWarnings = append(allWarnings, Warnings(warnings)...)

	return allWarnings, err
}
