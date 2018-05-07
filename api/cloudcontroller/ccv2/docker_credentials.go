package ccv2

// DockerCredentials are the authentication credentials to pull a docker image
// from it's repository.
type DockerCredentials struct {
	// Username is the username for a user that has access to a given docker
	// image.
	Username string `json:"username,omitempty"`

	// Password is the password for the user.
	Password string `json:"password,omitempty"`
}
