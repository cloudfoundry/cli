package configv3

type DynamicUserConfig struct {
	ConfigFile           *JSONConfig
	DefaultUserConfig    UserConfig
	KubernetesUserConfig UserConfig
}

func (config DynamicUserConfig) CurrentUser() (User, error) {
	return config.pickConfig().CurrentUser()
}

func (config DynamicUserConfig) CurrentUserName() (string, error) {
	return config.pickConfig().CurrentUserName()
}

func (config DynamicUserConfig) pickConfig() UserConfig {
	if config.ConfigFile.CFOnK8s.Enabled {
		return config.KubernetesUserConfig
	}
	return config.DefaultUserConfig
}
