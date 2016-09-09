package configactions

type Warnings []string

type Actor struct {
	Config                Config
	CloudControllerClient CloudControllerClient
}

func NewActor(config Config, client CloudControllerClient) Actor {
	return Actor{
		Config:                config,
		CloudControllerClient: client,
	}
}
