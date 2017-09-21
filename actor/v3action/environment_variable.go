package v3action

import (
	"fmt"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/types"
)

// EnvironmentVariables represents a map of env variables on an application
type EnvironmentVariables ccv3.EnvironmentVariables

// EnvironmentVariableNotSetError is returned when trying to unset env variable
// that was not previously set.
type EnvironmentVariableNotSetError struct {
	EnvironmentVariableName string
}

func (e EnvironmentVariableNotSetError) Error() string {
	return fmt.Sprintf("Env variable %s was not set.", e.EnvironmentVariableName)
}

// EnvironmentVariablePair represents an environment variable and its value
// on an application
type EnvironmentVariablePair struct {
	Key   string
	Value string
}

// GetEnvironmentVariableByApplicationNameAndSpace returns the environment
// variables for an application.
func (actor *Actor) GetEnvironmentVariableByApplicationNameAndSpace(appName string, spaceGUID string) (EnvironmentVariables, Warnings, error) {
	app, warnings, appErr := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if appErr != nil {
		return EnvironmentVariables{}, warnings, appErr
	}

	envVars, v3Warnings, apiErr := actor.CloudControllerClient.GetApplicationEnvironmentVariables(app.GUID)
	warnings = append(warnings, v3Warnings...)
	return EnvironmentVariables(envVars), warnings, apiErr
}

// SetEnvironmentVariableByApplicationNameAndSpace adds an
// EnvironmentVariablePair to an application. It must be restarted for changes
// to take effect.
func (actor *Actor) SetEnvironmentVariableByApplicationNameAndSpace(appName string, spaceGUID string, envPair EnvironmentVariablePair) (Warnings, error) {
	app, warnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return warnings, err
	}

	_, v3Warnings, apiErr := actor.CloudControllerClient.PatchApplicationEnvironmentVariables(app.GUID, ccv3.EnvironmentVariables{Variables: map[string]types.FilteredString{envPair.Key: {Value: envPair.Value, IsSet: true}}})
	warnings = append(warnings, v3Warnings...)
	return warnings, apiErr
}

// UnsetEnvironmentVariableByApplicationNameAndSpace removes an enviornment
// variable from an application. It must be restarted for changes to take
// effect.
func (actor *Actor) UnsetEnvironmentVariableByApplicationNameAndSpace(appName string, spaceGUID string, environmentVariableName string) (Warnings, error) {
	app, warnings, appErr := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if appErr != nil {
		return warnings, appErr
	}
	envVars, getWarnings, getErr := actor.CloudControllerClient.GetApplicationEnvironmentVariables(app.GUID)
	warnings = append(warnings, getWarnings...)
	if getErr != nil {
		return warnings, getErr
	}

	if _, ok := envVars.Variables[environmentVariableName]; !ok {
		return warnings, EnvironmentVariableNotSetError{EnvironmentVariableName: environmentVariableName}
	}

	_, patchWarnings, patchErr := actor.CloudControllerClient.PatchApplicationEnvironmentVariables(app.GUID, ccv3.EnvironmentVariables{Variables: map[string]types.FilteredString{environmentVariableName: {Value: "", IsSet: false}}})
	warnings = append(warnings, patchWarnings...)
	return warnings, patchErr
}
