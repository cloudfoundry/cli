package v2action

// CloudControllerAPIVersion returns the Cloud Controller API version.
func (actor Actor) CloudControllerAPIVersion() string {
	return actor.CloudControllerClient.APIVersion()
}
