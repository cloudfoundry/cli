package v7action

// GetUAAAPIVersion returns the UAA API version.
func (actor Actor) GetUAAAPIVersion() (string, error) {
	// NOTE: We are making a request here because this method is currently only
	// used in one branch of one command. However, we could probably do a refactor
	// to store the UAA version in the config file upon login, like we do with the
	// UAA URL, so we could just read from there instead of making a request here.

	return actor.UAAClient.GetAPIVersion()
}
