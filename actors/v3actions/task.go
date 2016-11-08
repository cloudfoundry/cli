package v3actions

import (
	"fmt"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

// Task represents a V3 actor Task.
type Task ccv3.Task

// TasksNotFoundError represents the scenario when no tasks were found
// associated with a particular application.
type TasksNotFoundError struct{}

func (e TasksNotFoundError) Error() string {
	return fmt.Sprintf("No tasks were found.")
}

// RunTask runs the provided command in the application environment associated
// with the provided application GUID.
func (actor Actor) RunTask(appGUID string, command string) (Task, Warnings, error) {
	task, warnings, err := actor.CloudControllerClient.RunTask(appGUID, command)
	return Task(task), Warnings(warnings), err
}

// GetApplicationTasks returns a list of tasks associated with the provided
// appplication GUID.
func (actor Actor) GetApplicationTasks(appGUID string) ([]Task, Warnings, error) {
	tasks, warnings, err := actor.CloudControllerClient.GetApplicationTasks(appGUID, nil)
	actorWarnings := Warnings(warnings)
	if err != nil {
		return nil, actorWarnings, err
	}

	if len(tasks) == 0 {
		return nil, actorWarnings, TasksNotFoundError{}
	}

	allTasks := []Task{}
	for _, task := range tasks {
		allTasks = append(allTasks, Task(task))
	}

	return allTasks, actorWarnings, nil
}
