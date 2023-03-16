package sharedaction

type DefaultAuthActor struct {
	config Config
}

func NewDefaultAuthActor(config Config) DefaultAuthActor {
	return DefaultAuthActor{
		config: config,
	}
}

func (a DefaultAuthActor) IsLoggedIn() bool {
	return a.config.AccessToken() != "" || a.config.RefreshToken() != ""
}

type K8sAuthActor struct {
	config Config
}

func NewK8sAuthActor(config Config) K8sAuthActor {
	return K8sAuthActor{
		config: config,
	}
}

func (a K8sAuthActor) IsLoggedIn() bool {
	name, err := a.config.CurrentUserName()

	return err == nil && name != ""
}
