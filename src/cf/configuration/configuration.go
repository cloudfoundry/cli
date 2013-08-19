package configuration

type Configuration struct {
	Target string
	ApiVersion string

}

func Default() (c Configuration) {
	c.Target = "https://api.run.pivotal.io"
	c.ApiVersion = "2"
	return
}
