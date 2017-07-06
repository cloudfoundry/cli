package pushaction

type CommandLineSettings struct {
	CurrentDirectory string
	ProvidedAppPath  string
	DockerImage      string
	Name             string
}

func (settings CommandLineSettings) ApplicationPath() string {
	if settings.ProvidedAppPath == "" {
		return settings.CurrentDirectory
	}
	return settings.ProvidedAppPath
}
