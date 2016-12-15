package v3action

func (actor Actor) CloudControllerAPIVersion() string {
	return actor.CloudControllerClient.CloudControllerAPIVersion()
}
