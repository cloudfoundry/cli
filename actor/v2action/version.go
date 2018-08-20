package v2action

// CloudControllerAPIVersion returns the Cloud Controller API version.
func (actor Actor) CloudControllerAPIVersion() string {
	return actor.CloudControllerClient.APIVersion()
}

// UAAAPIVersion returns the UAA API version.
func (actor Actor) UAAAPIVersion() string {
	return actor.UAAClient.APIVersion()
}
