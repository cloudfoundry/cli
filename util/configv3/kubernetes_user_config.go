package configv3

type KubernetesUserConfig struct {
	// ConfigFile stores the configuration from the .cf/config
	ConfigFile *JSONConfig
}

// CurrentUser returns user information decoded from the JWT access token in
// .cf/config.json.
func (config KubernetesUserConfig) CurrentUser() (User, error) {
	return User{Name: config.ConfigFile.CFOnK8s.AuthInfo}, nil
}

// CurrentUserName returns the name of a user as returned by CurrentUser()
func (config KubernetesUserConfig) CurrentUserName() (string, error) {
	return config.ConfigFile.CFOnK8s.AuthInfo, nil
}
