package manifestparser

import "errors"

type Application struct {
	Name string `yaml:"name"`
	Data map[string]interface{}
}

func (application Application) MarshalYAML() (interface{}, error) {
	return application.Data, nil
}

func (app *Application) UnmarshalYAML(unmarshal func(v interface{}) error) error {
	err := unmarshal(&app.Data)
	if err != nil {
		return err
	}

	if name, ok := app.Data["name"].(string); ok {
		app.Name = name
		return nil
	}

	return errors.New("Found an application with no name specified")
}
