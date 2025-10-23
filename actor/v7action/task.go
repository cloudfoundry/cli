package v7action

import (
	"sort"
	"strconv"
	"time"

	"code.cloudfoundry.org/cli/v8/actor/actionerror"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/v8/resources"
	log "github.com/sirupsen/logrus"
)

// Run resources.Task runs the provided command in the application environment associated
// with the provided application GUID.
func (actor Actor) RunTask(appGUID string, task resources.Task) (resources.Task, Warnings, error) {
	createdTask, warnings, err := actor.CloudControllerClient.CreateApplicationTask(appGUID, resources.Task(task))
	if err != nil {
		if e, ok := err.(ccerror.TaskWorkersUnavailableError); ok {
			return resources.Task{}, Warnings(warnings), actionerror.TaskWorkersUnavailableError{Message: e.Error()}
		}
	}

	return resources.Task(createdTask), Warnings(warnings), err
}

// GetApplicationTasks returns a list of tasks associated with the provided
// application GUID.
func (actor Actor) GetApplicationTasks(appGUID string, sortOrder SortOrder) ([]resources.Task, Warnings, error) {
	tasks, warnings, err := actor.CloudControllerClient.GetApplicationTasks(appGUID)
	actorWarnings := Warnings(warnings)
	if err != nil {
		return nil, actorWarnings, err
	}

	allTasks := []resources.Task{}
	for _, task := range tasks {
		allTasks = append(allTasks, resources.Task(task))
	}

	if sortOrder == Descending {
		sort.Slice(allTasks, func(i int, j int) bool { return allTasks[i].SequenceID > allTasks[j].SequenceID })
	} else {
		sort.Slice(allTasks, func(i int, j int) bool { return allTasks[i].SequenceID < allTasks[j].SequenceID })
	}

	return allTasks, actorWarnings, nil
}

func (actor Actor) GetTaskBySequenceIDAndApplication(sequenceID int, appGUID string) (resources.Task, Warnings, error) {
	tasks, warnings, err := actor.CloudControllerClient.GetApplicationTasks(
		appGUID,
		ccv3.Query{Key: ccv3.SequenceIDFilter, Values: []string{strconv.Itoa(sequenceID)}},
		ccv3.Query{Key: ccv3.PerPage, Values: []string{"1"}},
		ccv3.Query{Key: ccv3.Page, Values: []string{"1"}},
	)
	if err != nil {
		return resources.Task{}, Warnings(warnings), err
	}

	if len(tasks) == 0 {
		return resources.Task{}, Warnings(warnings), actionerror.TaskNotFoundError{SequenceID: sequenceID}
	}

	return resources.Task(tasks[0]), Warnings(warnings), nil
}

func (actor Actor) TerminateTask(taskGUID string) (resources.Task, Warnings, error) {
	task, warnings, err := actor.CloudControllerClient.UpdateTaskCancel(taskGUID)
	return resources.Task(task), Warnings(warnings), err
}

func (actor Actor) PollTask(task resources.Task) (resources.Task, Warnings, error) {
	var allWarnings Warnings

	for task.State != constant.TaskSucceeded && task.State != constant.TaskFailed {

		time.Sleep(actor.Config.PollingInterval())

		ccTask, warnings, err := actor.CloudControllerClient.GetTask(task.GUID)
		log.WithFields(log.Fields{
			"task_guid": task.GUID,
			"state":     task.State,
		}).Debug("polling task state")

		allWarnings = append(allWarnings, warnings...)

		if err != nil {
			return resources.Task{}, allWarnings, err
		}

		task = resources.Task(ccTask)
	}

	if task.State == constant.TaskFailed {
		return task, allWarnings, actionerror.TaskFailedError{}
	}

	return task, allWarnings, nil
}
