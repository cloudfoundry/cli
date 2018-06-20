package pushaction

type Event string

const (
	BoundRoutes                     Event = "bound routes"
	BoundServices                   Event = "bound services"
	ConfiguringServices             Event = "configuring services"
	CreatedApplication              Event = "created application"
	CreatedRoutes                   Event = "created routes"
	CreatingAndMappingRoutes        Event = "creating and mapping routes"
	CreatingArchive                 Event = "creating archive"
	CreatingPackage                 Event = "creating package"
	ReadingArchive                  Event = "reading archive"
	ResourceMatching                Event = "resource matching"
	RetryUpload                     Event = "retry upload"
	SettingUpApplication            Event = "setting up application"
	SkippingApplicationCreation     Event = "skipping creation"
	UnmappingRoutes                 Event = "unmapping routes"
	UpdatedApplication              Event = "updated application"
	UploadDropletComplete           Event = "upload droplet complete"
	UploadingApplication            Event = "uploading application"
	UploadingApplicationWithArchive Event = "uploading application with archive"
	UploadingDroplet                Event = "uploading droplet"
	UploadWithArchiveComplete       Event = "upload complete"
	Complete                        Event = "complete"
)
