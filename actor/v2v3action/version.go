package v2v3action

// CloudControllerV3APIVersion returns back the V3 Cloud Controller API version.
func (actor Actor) CloudControllerV3APIVersion() string {
	return actor.V3Actor.CloudControllerAPIVersion()
}
