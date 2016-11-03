package v3actions

import "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"

// Task represents a V3 actor Task.
type Task ccv3.Task

// RunTask runs the provided command in the application environment associated
// with the provided application GUID.
func (actor Actor) RunTask(appGUID string, command string) (Task, Warnings, error) {
	task, warnings, err := actor.CloudControllerClient.RunTask(appGUID, command)
	return Task(task), Warnings(warnings), err
}
