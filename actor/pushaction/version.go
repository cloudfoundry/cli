package pushaction

// CloudControllerAPIVersion returns the Cloud Controller API version.
func (actor Actor) CloudControllerAPIVersion() string {
	return actor.V2Actor.CloudControllerAPIVersion()
}
