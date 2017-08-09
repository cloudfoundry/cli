package pushaction

type Event string

const (
	SettingUpApplication Event = "setting up application"
	CreatedApplication   Event = "created application"
	UpdatedApplication   Event = "updated application"
	ConfiguringRoutes    Event = "configuring routes"
	CreatedRoutes        Event = "created routes"
	BoundRoutes          Event = "bound routes"
	ConfiguringServices  Event = "configuring services"
	BoundServices        Event = "bound services"
	CreatingArchive      Event = "creating archive"
	ResourceMatching     Event = "resource matching"
	UploadingApplication Event = "uploading application"
	UploadComplete       Event = "upload complete"
	RetryUpload          Event = "retry upload"
	Complete             Event = "complete"
)
