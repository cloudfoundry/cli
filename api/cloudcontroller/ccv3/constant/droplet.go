package constant

// DropletState is the state of the droplet.
type DropletState string

const (
	// DropletAwaitingUpload is a droplet that has been created without a package.
	DropletAwaitingUpload DropletState = "AWAITING_UPLOAD"
	// DropletStaged is a droplet that has been properly processed.
	DropletStaged DropletState = "STAGED"
	// DropletFailed is a droplet that had failed the staging process.
	DropletFailed DropletState = "FAILED"
	// DropletCopying is a droplet that's being copied from another droplet.
	DropletCopying DropletState = "COPYING"
	// DropletExpired is a droplet that has expired and is no longer in the
	// system.
	DropletExpired DropletState = "EXPIRED"
)
