package manifestparser

// ApplicationModel can be accessed through the top level Application struct To
// add a field for the CLI to extract from the manifest, just add it to this
// struct.
type ApplicationModel struct {
	Name   string  `yaml:"name"`
	Docker *Docker `yaml:"docker"`
	Path   string  `yaml:"path"`
}

type Application struct {
	ApplicationModel
	FullUnmarshalledApplication map[string]interface{}
}

func (application Application) MarshalYAML() (interface{}, error) {
	return application.FullUnmarshalledApplication, nil
}

func (application *Application) UnmarshalYAML(unmarshal func(v interface{}) error) error {
	err := unmarshal(&application.FullUnmarshalledApplication)
	if err != nil {
		return err
	}
	return unmarshal(&application.ApplicationModel)
}

type Docker struct {
	Username string `yaml:"username"`
	Image    string `yaml:"image"`
}
