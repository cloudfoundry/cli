package manifestparser

type manifest struct {
	Applications []Application `yaml:"applications"`
}
