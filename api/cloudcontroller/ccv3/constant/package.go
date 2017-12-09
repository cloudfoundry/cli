package constant

// PackageState represents the state of a package.
type PackageState string

const (
	// PackageProcessingUpload is a package that's being process by the CC.
	PackageProcessingUpload PackageState = "PROCESSING_UPLOAD"
	// PackageReady is a package that's ready to use.
	PackageReady PackageState = "READY"
	// PackageFailed is a package that has failed to be constructed.
	PackageFailed PackageState = "FAILED"
	// PackageAwaitingUpload is a package that does not have any bits or settings
	// yet.
	PackageAwaitingUpload PackageState = "AWAITING_UPLOAD"
	// PackageCopying is a package that's being copied from another package.
	PackageCopying PackageState = "COPYING"
	// PackageExpired is a package that has expired and is no longer in the
	// system.
	PackageExpired PackageState = "EXPIRED"
)

// PackageType represents the type of package.
type PackageType string

const (
	// PackageTypeBits is used to upload source code for an app.
	PackageTypeBits PackageType = "bits"
	// PackageTypeDocker references a docker image from a registry.
	PackageTypeDocker PackageType = "docker"
)
