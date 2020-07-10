package v2action

// CloudControllerAPIVersion returns the Cloud Controller API version.
func (actor Actor) CloudControllerAPIVersion() string {
	return actor.CloudControllerClient.APIVersion()
}

// GetUAAAPIVersion returns the UAA API version.
func (actor Actor) GetUAAAPIVersion() (string, error) {
	return actor.UAAClient.GetAPIVersion()
}
