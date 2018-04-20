package pushaction

// CloudControllerV2APIVersion returns the Cloud Controller V2 API version.
func (actor Actor) CloudControllerV2APIVersion() string {
	return actor.V2Actor.CloudControllerAPIVersion()
}

// CloudControllerV3APIVersion returns the Cloud Controller V3 API version.
func (actor Actor) CloudControllerV3APIVersion() string {
	return actor.V3Actor.CloudControllerAPIVersion()
}
