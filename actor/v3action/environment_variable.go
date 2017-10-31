package v3action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/types"
)

// EnvironmentVariableGroups represents all environment variables for application
type EnvironmentVariableGroups ccv3.EnvironmentVariableGroups

// EnvironmentVariablePair represents an environment variable and its value
// on an application
type EnvironmentVariablePair struct {
	Key   string
	Value string
}

// GetEnvironmentVariablesByApplicationNameAndSpace returns the environment
// variables for an application.
func (actor *Actor) GetEnvironmentVariablesByApplicationNameAndSpace(appName string, spaceGUID string) (EnvironmentVariableGroups, Warnings, error) {
	app, warnings, appErr := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if appErr != nil {
		return EnvironmentVariableGroups{}, warnings, appErr
	}

	ccEnvGroups, v3Warnings, apiErr := actor.CloudControllerClient.GetApplicationEnvironmentVariables(app.GUID)
	warnings = append(warnings, v3Warnings...)
	return EnvironmentVariableGroups(ccEnvGroups), warnings, apiErr
}

// SetEnvironmentVariableByApplicationNameAndSpace adds an
// EnvironmentVariablePair to an application. It must be restarted for changes
// to take effect.
func (actor *Actor) SetEnvironmentVariableByApplicationNameAndSpace(appName string, spaceGUID string, envPair EnvironmentVariablePair) (Warnings, error) {
	app, warnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return warnings, err
	}

	_, v3Warnings, apiErr := actor.CloudControllerClient.PatchApplicationUserProvidedEnvironmentVariables(app.GUID, ccv3.EnvironmentVariables{Variables: map[string]types.FilteredString{envPair.Key: {Value: envPair.Value, IsSet: true}}})
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
	envGroups, getWarnings, getErr := actor.CloudControllerClient.GetApplicationEnvironmentVariables(app.GUID)
	warnings = append(warnings, getWarnings...)
	if getErr != nil {
		return warnings, getErr
	}

	if _, ok := envGroups.UserProvided[environmentVariableName]; !ok {
		return warnings, actionerror.EnvironmentVariableNotSetError{EnvironmentVariableName: environmentVariableName}
	}

	_, patchWarnings, patchErr := actor.CloudControllerClient.PatchApplicationUserProvidedEnvironmentVariables(app.GUID, ccv3.EnvironmentVariables{Variables: map[string]types.FilteredString{environmentVariableName: {Value: "", IsSet: false}}})
	warnings = append(warnings, patchWarnings...)
	return warnings, patchErr
}
