package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/types"
)

// EnvironmentVariableGroups represents all environment variables for application
type EnvironmentVariableGroups ccv3.Environment

// EnvironmentVariableGroup represents a CC environment variable group (e.g. staging or running)
type EnvironmentVariableGroup ccv3.EnvironmentVariables

// EnvironmentVariablePair represents an environment variable and its value
// on an application
type EnvironmentVariablePair struct {
	Key   string
	Value string
}

// GetEnvironmentVariableGroup returns the values of an environment variable group.
func (actor *Actor) GetEnvironmentVariableGroup(group constant.EnvironmentVariableGroupName) (EnvironmentVariableGroup, Warnings, error) {
	ccEnvGroup, warnings, err := actor.CloudControllerClient.GetEnvironmentVariableGroup(group)
	return EnvironmentVariableGroup(ccEnvGroup), Warnings(warnings), err
}

// GetEnvironmentVariablesByApplicationNameAndSpace returns the environment
// variables for an application.
func (actor *Actor) GetEnvironmentVariablesByApplicationNameAndSpace(appName string, spaceGUID string) (EnvironmentVariableGroups, Warnings, error) {
	app, warnings, appErr := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if appErr != nil {
		return EnvironmentVariableGroups{}, warnings, appErr
	}

	ccEnvGroups, v3Warnings, apiErr := actor.CloudControllerClient.GetApplicationEnvironment(app.GUID)
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

	_, v3Warnings, apiErr := actor.CloudControllerClient.UpdateApplicationEnvironmentVariables(
		app.GUID,
		ccv3.EnvironmentVariables{
			envPair.Key: {Value: envPair.Value, IsSet: true},
		})
	warnings = append(warnings, v3Warnings...)
	return warnings, apiErr
}

// SetEnvironmentVariableGroup sets a given environment variable group according to the given
// keys and values. Any existing variables that are not present in the given set of variables
// will be unset.
func (actor *Actor) SetEnvironmentVariableGroup(group constant.EnvironmentVariableGroupName, newEnvVars ccv3.EnvironmentVariables) (Warnings, error) {
	var allWarnings Warnings

	existingEnvVars, warnings, err := actor.CloudControllerClient.GetEnvironmentVariableGroup(group)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	for k := range existingEnvVars {
		if _, ok := newEnvVars[k]; !ok {
			newEnvVars[k] = types.FilteredString{IsSet: false}
		}
	}

	_, warnings, err = actor.CloudControllerClient.UpdateEnvironmentVariableGroup(group, newEnvVars)
	allWarnings = append(allWarnings, warnings...)

	return allWarnings, err
}

// UnsetEnvironmentVariableByApplicationNameAndSpace removes an environment
// variable from an application. It must be restarted for changes to take
// effect.
func (actor *Actor) UnsetEnvironmentVariableByApplicationNameAndSpace(appName string, spaceGUID string, environmentVariableName string) (Warnings, error) {
	app, warnings, appErr := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if appErr != nil {
		return warnings, appErr
	}
	envGroups, getWarnings, getErr := actor.CloudControllerClient.GetApplicationEnvironment(app.GUID)
	warnings = append(warnings, getWarnings...)
	if getErr != nil {
		return warnings, getErr
	}

	if _, ok := envGroups.EnvironmentVariables[environmentVariableName]; !ok {
		return warnings, actionerror.EnvironmentVariableNotSetError{EnvironmentVariableName: environmentVariableName}
	}

	_, patchWarnings, patchErr := actor.CloudControllerClient.UpdateApplicationEnvironmentVariables(
		app.GUID,
		ccv3.EnvironmentVariables{
			environmentVariableName: {Value: "", IsSet: false},
		})
	warnings = append(warnings, patchWarnings...)
	return warnings, patchErr
}
