package constant

// DropletState is the state of the droplet.
type DropletState string

const (
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
