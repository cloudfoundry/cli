package sharedaction

// IsLoggedIn checks whether a user has authenticated with CF
func (actor Actor) IsLoggedIn() bool {
	return actor.Config.AccessToken() != "" && actor.Config.RefreshToken() != ""
}
