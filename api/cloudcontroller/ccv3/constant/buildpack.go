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
	// BuildpackAwaitingUpload represents the awaiting upload state of a buildpack.
	BuildpackAwaitingUpload = "AWAITING_UPLOAD"
	// BuildpackReady represents the ready state of a buildpack.
	BuildpackReady = "READY"
)
