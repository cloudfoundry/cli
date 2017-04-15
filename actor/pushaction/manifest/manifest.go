package manifest

type Manifest struct {
	Applications []Application
}

type Application struct {
	Name string
	Path string
}
