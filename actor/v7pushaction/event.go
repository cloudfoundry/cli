package v7pushaction

type PushEvent struct {
	Event    Event
	Plan     PushPlan
	Err      error
	Warnings Warnings
}

type Event string

const (
	ApplyManifest                   Event = "Applying manifest"
	ApplyManifestComplete           Event = "Applying manifest Complete"
	CreatingArchive                 Event = "creating archive"
	CreatingDroplet                 Event = "creating droplet"
	CreatingPackage                 Event = "creating package"
	PollingBuild                    Event = "polling build"
	ReadingArchive                  Event = "reading archive"
	ResourceMatching                Event = "resource matching"
	RestartingApplication           Event = "restarting application"
	RestartingApplicationComplete   Event = "restarting application complete"
	RetryUpload                     Event = "retry upload"
	SetDockerImage                  Event = "setting docker properties"
	SetDockerImageComplete          Event = "completed setting docker properties"
	SetDropletComplete              Event = "set droplet complete"
	SettingDroplet                  Event = "setting droplet"
	StagingComplete                 Event = "staging complete"
	StartingDeployment              Event = "starting deployment"
	StartingStaging                 Event = "starting staging"
	StoppingApplication             Event = "stopping application"
	StoppingApplicationComplete     Event = "stopping application complete"
	UploadDropletComplete           Event = "upload droplet complete"
	UploadingApplication            Event = "uploading application"
	UploadingApplicationWithArchive Event = "uploading application with archive"
	UploadingDroplet                Event = "uploading droplet"
	UploadWithArchiveComplete       Event = "upload complete"
	WaitingForDeployment            Event = "waiting for deployment"
)
