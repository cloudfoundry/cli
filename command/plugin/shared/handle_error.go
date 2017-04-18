package shared

import "code.cloudfoundry.org/cli/actor/pluginaction"

func HandleError(err error) error {
	switch e := err.(type) {
	case pluginaction.PluginNotFoundError:
		return PluginNotFoundError{Name: e.Name}
	}

	return err
}
