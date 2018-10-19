package v7action

// CloudControllerAPIVersion returns back the Cloud Controller API version.
func (actor Actor) CloudControllerAPIVersion() string {
	return actor.CloudControllerClient.CloudControllerAPIVersion()
}
