package manifestparser

import (
	"reflect"
	"strings"
)

// ApplicationModel can be accessed through the top level Application struct To
// add a field for the CLI to extract from the manifest, just add it to this
// struct.
type ApplicationModel struct {
	Name    string  `yaml:"name"`
	Docker  *Docker `yaml:"docker,omitempty"`
	Path    string  `yaml:"path,omitempty"`
	NoRoute bool    `yaml:"no-route,omitempty"`
}

type Application struct {
	ApplicationModel            `yaml:",inline"`
	FullUnmarshalledApplication map[string]interface{} `yaml:",inline"`
}

func (application *Application) UnmarshalYAML(unmarshal func(v interface{}) error) error {
	err := unmarshal(&application.FullUnmarshalledApplication)
	if err != nil {
		return err
	}
	err = unmarshal(&application.ApplicationModel)
	if err != nil {
		return err
	}

	// Remove any ApplicationModel keys from the generic unmarshal to avoid repeats
	appModelStruct := reflect.TypeOf(application.ApplicationModel)

	for i := 0; i < appModelStruct.NumField(); i++ {
		structField := appModelStruct.Field(i)

		// specification of yaml tag found here https://godoc.org/gopkg.in/yaml.v2#Marshal
		yamlKey := strings.Split(structField.Tag.Get("yaml"), ",")[0]
		if yamlKey == "" {
			yamlKey = strings.ToLower(structField.Name)
		}

		delete(application.FullUnmarshalledApplication, yamlKey)
	}
	return nil
}

type Docker struct {
	Image    string `yaml:"image,omitempty"`
	Username string `yaml:"username,omitempty"`
}
