package constant

const (
	// AutodetectBuildpackValueDefault is used to unset the buildpack values on
	// an application.
	AutodetectBuildpackValueDefault = "default"
	// AutodetectBuildpackValueNull is used to unset the buildpack values on an
	// application.
	AutodetectBuildpackValueNull = "null"
)

const (
	// Represents the awaiting upload state of a buildpack
	BuildpackAwaitingUpload = "AWAITING_UPLOAD"
	// Represents the ready state of a buildpack
	BuildpackReady = "READY"
)
