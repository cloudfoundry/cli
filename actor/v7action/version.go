package v7action

// CloudControllerAPIVersion returns the Cloud Controller API version.
func (actor Actor) CloudControllerAPIVersion() string {
	return actor.CloudControllerClient.CloudControllerAPIVersion()
}

// UAAAPIVersion returns the UAA API version.
func (actor Actor) UAAAPIVersion() string {
	return actor.UAAClient.APIVersion()
}
