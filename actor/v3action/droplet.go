package v3action

// Droplet represents a Cloud Controller droplet.
type Droplet struct {
	GUID       string
	Stack      string
	Buildpacks []Buildpack
}
