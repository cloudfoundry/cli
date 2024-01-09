package v7action

import (
	"errors"
	"fmt"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/batcher"
	"code.cloudfoundry.org/cli/util/unique"
)

func (actor Actor) DeleteApplicationByNameAndSpace(name, spaceGUID string, deleteRoutes bool) (Warnings, error) {
	var allWarnings Warnings
	var jobQueue []ccv3.JobURL

	app, getAppWarnings, err := actor.GetApplicationByNameAndSpace(name, spaceGUID)
	allWarnings = append(allWarnings, getAppWarnings...)
	if err != nil {
		return allWarnings, err
	}

	var routes []resources.Route
	if deleteRoutes {
		var getRoutesWarnings Warnings
		routes, getRoutesWarnings, err = actor.GetApplicationRoutes(app.GUID)
		allWarnings = append(allWarnings, getRoutesWarnings...)
		if err != nil {
			return allWarnings, err
		}

		for _, route := range routes {
			if len(route.Destinations) > 1 {
				for _, destination := range route.Destinations {
					guid := destination.App.GUID
					if guid != app.GUID {
						return allWarnings, actionerror.RouteBoundToMultipleAppsError{AppName: app.Name, RouteURL: route.URL}
					}
				}
			}
		}
	}

	appDeleteJobURL, deleteAppWarnings, err := actor.CloudControllerClient.DeleteApplication(app.GUID)
	allWarnings = append(allWarnings, deleteAppWarnings...)
	if err != nil {
		return allWarnings, err
	}

	pollWarnings, err := actor.CloudControllerClient.PollJob(appDeleteJobURL)
	allWarnings = append(allWarnings, pollWarnings...)
	if err != nil {
		return allWarnings, err
	}

	if deleteRoutes {
		for _, route := range routes {
			jobURL, deleteRouteWarnings, err := actor.CloudControllerClient.DeleteRoute(route.GUID)
			allWarnings = append(allWarnings, deleteRouteWarnings...)
			if err != nil {
				if _, ok := err.(ccerror.ResourceNotFoundError); ok {
					continue
				}
				return allWarnings, err
			}

			jobQueue = append(jobQueue, jobURL)
		}
	}

	for _, job := range jobQueue {
		pollWarnings, err := actor.CloudControllerClient.PollJob(job)
		allWarnings = append(allWarnings, pollWarnings...)
		if err != nil {
			return allWarnings, err
		}
	}

	return allWarnings, err
}

func (actor Actor) GetApplicationsByGUIDs(appGUIDs []string) ([]resources.Application, Warnings, error) {
	uniqueAppGUIDs := unique.StringSlice(appGUIDs)

	var apps []resources.Application
	warnings, err := batcher.RequestByGUID(appGUIDs, func(guids []string) (ccv3.Warnings, error) {
		batch, warnings, err := actor.CloudControllerClient.GetApplications(
			ccv3.Query{Key: ccv3.GUIDFilter, Values: guids},
		)
		apps = append(apps, batch...)
		return warnings, err
	})

	if err != nil {
		return nil, Warnings(warnings), err
	}

	if len(apps) < len(uniqueAppGUIDs) {
		return nil, Warnings(warnings), actionerror.ApplicationsNotFoundError{}
	}

	return apps, Warnings(warnings), nil
}

func (actor Actor) GetApplicationsByNamesAndSpace(appNames []string, spaceGUID string) ([]resources.Application, Warnings, error) {
	uniqueAppNames := unique.StringSlice(appNames)

	apps, warnings, err := actor.CloudControllerClient.GetApplications(
		ccv3.Query{Key: ccv3.NameFilter, Values: appNames},
		ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{spaceGUID}},
	)

	if err != nil {
		return nil, Warnings(warnings), err
	}

	if len(apps) < len(uniqueAppNames) {
		return nil, Warnings(warnings), actionerror.ApplicationsNotFoundError{}
	}

	return apps, Warnings(warnings), nil
}

// GetApplicationByNameAndSpace returns the application with the given
// name in the given space.
func (actor Actor) GetApplicationByNameAndSpace(appName string, spaceGUID string) (resources.Application, Warnings, error) {
	apps, warnings, err := actor.GetApplicationsByNamesAndSpace([]string{appName}, spaceGUID)

	if err != nil {
		if _, ok := err.(actionerror.ApplicationsNotFoundError); ok {
			return resources.Application{}, warnings, actionerror.ApplicationNotFoundError{Name: appName}
		}
		return resources.Application{}, warnings, err
	}

	return apps[0], warnings, nil
}

// GetApplicationsBySpace returns all applications in a space.
func (actor Actor) GetApplicationsBySpace(spaceGUID string) ([]resources.Application, Warnings, error) {
	apps, warnings, err := actor.CloudControllerClient.GetApplications(
		ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{spaceGUID}},
	)

	if err != nil {
		return []resources.Application{}, Warnings(warnings), err
	}

	return apps, Warnings(warnings), nil
}

// CreateApplicationInSpace creates and returns the application with the given
// name in the given space.
func (actor Actor) CreateApplicationInSpace(app resources.Application, spaceGUID string) (resources.Application, Warnings, error) {
	createdApp, warnings, err := actor.CloudControllerClient.CreateApplication(
		resources.Application{
			LifecycleType:       app.LifecycleType,
			LifecycleBuildpacks: app.LifecycleBuildpacks,
			StackName:           app.StackName,
			Name:                app.Name,
			SpaceGUID:           spaceGUID,
		})

	if err != nil {
		return resources.Application{}, Warnings(warnings), err
	}

	return createdApp, Warnings(warnings), nil
}

// SetApplicationProcessHealthCheckTypeByNameAndSpace sets the health check
// information of the provided processType for an application with the given
// name and space GUID.
func (actor Actor) SetApplicationProcessHealthCheckTypeByNameAndSpace(
	appName string,
	spaceGUID string,
	healthCheckType constant.HealthCheckType,
	httpEndpoint string,
	processType string,
	invocationTimeout int64,
) (resources.Application, Warnings, error) {

	app, getWarnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return resources.Application{}, getWarnings, err
	}

	setWarnings, err := actor.UpdateProcessByTypeAndApplication(
		processType,
		app.GUID,
		resources.Process{
			HealthCheckType:              healthCheckType,
			HealthCheckEndpoint:          httpEndpoint,
			HealthCheckInvocationTimeout: invocationTimeout,
		})
	return app, append(getWarnings, setWarnings...), err
}

// StopApplication stops an application.
func (actor Actor) StopApplication(appGUID string) (Warnings, error) {
	_, warnings, err := actor.CloudControllerClient.UpdateApplicationStop(appGUID)

	return Warnings(warnings), err
}

// StartApplication starts an application.
func (actor Actor) StartApplication(appGUID string) (Warnings, error) {
	_, warnings, err := actor.CloudControllerClient.UpdateApplicationStart(appGUID)
	return Warnings(warnings), err
}

// RestartApplication restarts an application and waits for it to start.
func (actor Actor) RestartApplication(appGUID string, noWait bool) (Warnings, error) {
	// var allWarnings Warnings
	_, warnings, err := actor.CloudControllerClient.UpdateApplicationRestart(appGUID)
	return Warnings(warnings), err
}

func (actor Actor) GetUnstagedNewestPackageGUID(appGUID string) (string, Warnings, error) {
	var err error
	var allWarnings Warnings
	packages, warnings, err := actor.CloudControllerClient.GetPackages(
		ccv3.Query{Key: ccv3.AppGUIDFilter, Values: []string{appGUID}},
		ccv3.Query{Key: ccv3.OrderBy, Values: []string{ccv3.CreatedAtDescendingOrder}},
		ccv3.Query{Key: ccv3.PerPage, Values: []string{"1"}},
		ccv3.Query{Key: ccv3.Page, Values: []string{"1"}})
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return "", allWarnings, err
	}
	if len(packages) == 0 {
		return "", allWarnings, nil
	}

	newestPackage := packages[0]

	droplets, warnings, err := actor.CloudControllerClient.GetPackageDroplets(
		newestPackage.GUID,
		ccv3.Query{Key: ccv3.StatesFilter, Values: []string{"STAGED"}},
		ccv3.Query{Key: ccv3.PerPage, Values: []string{"1"}},
		ccv3.Query{Key: ccv3.Page, Values: []string{"1"}},
	)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return "", allWarnings, err
	}

	if len(droplets) == 0 {
		return newestPackage.GUID, allWarnings, nil
	}

	return "", allWarnings, nil
}

// PollStart polls an application's processes until some are started. If noWait is false,
// it waits for at least one instance of all processes to be running. If noWait is true,
// it only waits for an instance of the web process to be running.
func (actor Actor) PollStart(app resources.Application, noWait bool, handleInstanceDetails func(string)) (Warnings, error) {
	var allWarnings Warnings
	processes, warnings, err := actor.CloudControllerClient.GetApplicationProcesses(app.GUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	var filteredProcesses []resources.Process
	if noWait {
		for _, process := range processes {
			if process.Type == constant.ProcessTypeWeb {
				filteredProcesses = append(filteredProcesses, process)
			}
		}
	} else {
		filteredProcesses = processes
	}

	timer := actor.Clock.NewTimer(time.Millisecond)
	defer timer.Stop()
	timeout := actor.Clock.After(actor.Config.StartupTimeout())

	for {
		select {
		case <-timeout:
			return allWarnings, actionerror.StartupTimeoutError{Name: app.Name}
		case <-timer.C():
			stopPolling, warnings, err := actor.PollProcesses(filteredProcesses, handleInstanceDetails)
			allWarnings = append(allWarnings, warnings...)
			if stopPolling || err != nil {
				return allWarnings, err
			}

			timer.Reset(actor.Config.PollingInterval())
		}
	}
}

// PollStartForRolling polls a deploying application's processes until some are started. It does the same thing as PollStart, except it accounts for rolling deployments and whether
// they have failed or been canceled during polling.
func (actor Actor) PollStartForRolling(app resources.Application, deploymentGUID string, noWait bool, handleInstanceDetails func(string)) (Warnings, error) {
	var (
		deployment  resources.Deployment
		processes   []resources.Process
		allWarnings Warnings
	)

	timer := actor.Clock.NewTimer(time.Millisecond)
	defer timer.Stop()
	timeout := actor.Clock.After(actor.Config.StartupTimeout())

	for {
		select {
		case <-timeout:
			warnings, err := actor.CancelDeployment(deploymentGUID)
			allWarnings = append(allWarnings, warnings...)
			if err != nil {
				return allWarnings, err
			}
			return allWarnings, actionerror.StartupTimeoutError{Name: app.Name}
		case <-timer.C():
			if !isDeployed(deployment) {
				ccDeployment, warnings, err := actor.getDeployment(deploymentGUID)
				allWarnings = append(allWarnings, warnings...)
				if err != nil {
					return allWarnings, err
				}
				deployment = ccDeployment
				processes, warnings, err = actor.getProcesses(deployment, app.GUID, noWait)
				allWarnings = append(allWarnings, warnings...)
				if err != nil {
					return allWarnings, err
				}
			}

			if noWait || isDeployed(deployment) {
				stopPolling, warnings, err := actor.PollProcesses(processes, handleInstanceDetails)
				allWarnings = append(allWarnings, warnings...)
				if stopPolling || err != nil {
					return allWarnings, err
				}
			}

			timer.Reset(actor.Config.PollingInterval())
		}
	}
}

func isDeployed(d resources.Deployment) bool {
	return d.StatusValue == constant.DeploymentStatusValueFinalized && d.StatusReason == constant.DeploymentStatusReasonDeployed
}

// PollProcesses - return true if there's no need to keep polling
func (actor Actor) PollProcesses(processes []resources.Process, handleInstanceDetails func(string)) (bool, Warnings, error) {
	numProcesses := len(processes)
	numStableProcesses := 0
	var allWarnings Warnings
	for _, process := range processes {
		ccInstances, ccWarnings, err := actor.CloudControllerClient.GetProcessInstances(process.GUID)
		instances := ProcessInstances(ccInstances)
		allWarnings = append(allWarnings, ccWarnings...)
		if err != nil {
			return true, allWarnings, err
		}

		handleInstanceDetails(formatInstanceDetails(instances))

		if instances.Empty() || instances.AnyRunning() {
			numStableProcesses += 1
			continue
		}

		if instances.AllCrashed() {
			return true, allWarnings, actionerror.AllInstancesCrashedError{}
		}

		//precondition: !instances.Empty() && no instances are running
		// do not increment numStableProcesses
		return false, allWarnings, nil
	}
	return numStableProcesses == numProcesses, allWarnings, nil
}

// UpdateApplication updates the buildpacks on an application
func (actor Actor) UpdateApplication(app resources.Application) (resources.Application, Warnings, error) {
	ccApp := resources.Application{
		GUID:                app.GUID,
		StackName:           app.StackName,
		LifecycleType:       app.LifecycleType,
		LifecycleBuildpacks: app.LifecycleBuildpacks,
		Metadata:            app.Metadata,
		Name:                app.Name,
	}

	updatedApp, warnings, err := actor.CloudControllerClient.UpdateApplication(ccApp)
	if err != nil {
		return resources.Application{}, Warnings(warnings), err
	}

	return updatedApp, Warnings(warnings), nil
}

// UpdateApplicationName updates the name of an application
func (actor Actor) UpdateApplicationName(newAppName string, appGUID string) (resources.Application, Warnings, error) {

	updatedApp, warnings, err := actor.CloudControllerClient.UpdateApplicationName(newAppName, appGUID)
	if err != nil {
		return resources.Application{}, Warnings(warnings), err
	}

	return updatedApp, Warnings(warnings), nil
}

func (actor Actor) getDeployment(deploymentGUID string) (resources.Deployment, Warnings, error) {
	deployment, warnings, err := actor.CloudControllerClient.GetDeployment(deploymentGUID)
	if err != nil {
		return deployment, Warnings(warnings), err
	}

	if deployment.StatusValue == constant.DeploymentStatusValueFinalized {
		switch deployment.StatusReason {
		case constant.DeploymentStatusReasonCanceled:
			return deployment, Warnings(warnings), errors.New("Deployment has been canceled")
		case constant.DeploymentStatusReasonSuperseded:
			return deployment, Warnings(warnings), errors.New("Deployment has been superseded")
		}
	}

	return deployment, Warnings(warnings), err
}

func (actor Actor) getProcesses(deployment resources.Deployment, appGUID string, noWait bool) ([]resources.Process, Warnings, error) {
	if noWait {
		// these are only web processes for now so we can just use these
		return deployment.NewProcesses, nil, nil
	}

	// if the deployment is deployed we know web are all running and PollProcesses will see those as stable
	// so just getting all processes is equivalent to just getting non-web ones and polling those
	if isDeployed(deployment) {
		processes, warnings, err := actor.CloudControllerClient.GetApplicationProcesses(appGUID)
		if err != nil {
			return processes, Warnings(warnings), err
		}
		return processes, Warnings(warnings), nil
	}

	return nil, nil, nil
}

func (actor Actor) RenameApplicationByNameAndSpaceGUID(appName, newAppName, spaceGUID string) (resources.Application, Warnings, error) {
	allWarnings := Warnings{}
	application, warnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return resources.Application{}, allWarnings, err
	}
	appGUID := application.GUID
	application, warnings, err = actor.UpdateApplicationName(newAppName, appGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return resources.Application{}, allWarnings, err
	}

	return application, allWarnings, nil
}

func formatInstanceDetails(instances ProcessInstances) string {
	for _, instance := range instances {
		if instance.Details != "" {
			return fmt.Sprintf("Error starting instances: '%s'", instance.Details)
		}
	}
	return "Instances starting..."
}
