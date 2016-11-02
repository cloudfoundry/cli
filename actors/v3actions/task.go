package v3actions

import "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"

// Task represents a CLI Task.
type Task ccv3.Task

func (actor Actor) RunTask(appGUID string, command string) (Task, Warnings, error) {
	task, warnings, err := actor.CloudControllerClient.RunTask(appGUID, command)
	return Task(task), Warnings(warnings), err
}
