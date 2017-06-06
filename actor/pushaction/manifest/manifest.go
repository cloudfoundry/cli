package manifest

type Manifest struct {
	Applications []Application
}

type Application struct {
	DockerImage string
	Name        string
	Path        string
}
