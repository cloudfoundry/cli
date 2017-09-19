package v3action

import "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"

// EnvironmentVariablePair represents an environment variable and its value
// on an application
type EnvironmentVariablePair struct {
	Key   string
	Value string
}

// SetEnvironmentVariableByApplicationNameAndSpace adds an
// EnvironmentVariablePair to an application. It must be restarted for changes
// to take effect.
func (actor *Actor) SetEnvironmentVariableByApplicationNameAndSpace(appName string, spaceGUID string, envPair EnvironmentVariablePair) (Warnings, error) {
	app, warnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return warnings, err
	}

	_, v3Warnings, apiErr := actor.CloudControllerClient.UpdateApplicationEnvironmentVariables(app.GUID, ccv3.EnvironmentVariables{Variables: map[string]string{envPair.Key: envPair.Value}})
	warnings = append(warnings, v3Warnings...)
	return warnings, apiErr
}
