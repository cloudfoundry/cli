package constant

// BuildState represents the current state of the build.
type BuildState string

const (
	// BuildFailed is when the build has failed/erred during staging.
	BuildFailed BuildState = "FAILED"
	// BuildStaged is when the build has successfully been staged.
	BuildStaged BuildState = "STAGED"
	// BuildStaging is when the build is in the process of being staged.
	BuildStaging BuildState = "STAGING"
)
