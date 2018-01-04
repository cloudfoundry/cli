package v3action

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

//go:generate counterfeiter . CloudControllerClient

// CloudControllerClient is the interface to the cloud controller V3 API.
type CloudControllerClient interface {
	AppSSHEndpoint() string
	AppSSHHostKeyFingerprint() string
	AssignSpaceToIsolationSegment(spaceGUID string, isolationSegmentGUID string) (ccv3.Relationship, ccv3.Warnings, error)
	CloudControllerAPIVersion() string
	CreateApplication(app ccv3.Application) (ccv3.Application, ccv3.Warnings, error)
	CreateApplicationProcessScale(appGUID string, process ccv3.Process) (ccv3.Process, ccv3.Warnings, error)
	CreateApplicationTask(appGUID string, task ccv3.Task) (ccv3.Task, ccv3.Warnings, error)
	CreateBuild(build ccv3.Build) (ccv3.Build, ccv3.Warnings, error)
	CreateIsolationSegment(isolationSegment ccv3.IsolationSegment) (ccv3.IsolationSegment, ccv3.Warnings, error)
	CreatePackage(pkg ccv3.Package) (ccv3.Package, ccv3.Warnings, error)
	DeleteApplication(guid string) (string, ccv3.Warnings, error)
	DeleteApplicationProcessInstance(appGUID string, processType string, instanceIndex int) (ccv3.Warnings, error)
	DeleteIsolationSegment(guid string) (ccv3.Warnings, error)
	EntitleIsolationSegmentToOrganizations(isoGUID string, orgGUIDs []string) (ccv3.RelationshipList, ccv3.Warnings, error)
	GetApplicationDropletCurrent(appGUID string) (ccv3.Droplet, ccv3.Warnings, error)
	GetApplicationEnvironmentVariables(appGUID string) (ccv3.EnvironmentVariableGroups, ccv3.Warnings, error)
	GetApplicationProcessByType(appGUID string, processType string) (ccv3.Process, ccv3.Warnings, error)
	GetApplicationProcesses(appGUID string) ([]ccv3.Process, ccv3.Warnings, error)
	GetApplications(query ...ccv3.Query) ([]ccv3.Application, ccv3.Warnings, error)
	GetApplicationTasks(appGUID string, query ...ccv3.Query) ([]ccv3.Task, ccv3.Warnings, error)
	GetBuild(guid string) (ccv3.Build, ccv3.Warnings, error)
	GetDroplet(guid string) (ccv3.Droplet, ccv3.Warnings, error)
	GetDroplets(query ...ccv3.Query) ([]ccv3.Droplet, ccv3.Warnings, error)
	GetIsolationSegment(guid string) (ccv3.IsolationSegment, ccv3.Warnings, error)
	GetIsolationSegmentOrganizationsByIsolationSegment(isolationSegmentGUID string) ([]ccv3.Organization, ccv3.Warnings, error)
	GetIsolationSegments(query ...ccv3.Query) ([]ccv3.IsolationSegment, ccv3.Warnings, error)
	GetOrganizationDefaultIsolationSegment(orgGUID string) (ccv3.Relationship, ccv3.Warnings, error)
	GetOrganizations(query ...ccv3.Query) ([]ccv3.Organization, ccv3.Warnings, error)
	GetPackage(guid string) (ccv3.Package, ccv3.Warnings, error)
	GetPackages(query ...ccv3.Query) ([]ccv3.Package, ccv3.Warnings, error)
	GetProcessInstances(processGUID string) ([]ccv3.ProcessInstance, ccv3.Warnings, error)
	GetServiceInstances(query ...ccv3.Query) ([]ccv3.ServiceInstance, ccv3.Warnings, error)
	GetSpaceIsolationSegment(spaceGUID string) (ccv3.Relationship, ccv3.Warnings, error)
	GetSpaces(query ...ccv3.Query) ([]ccv3.Space, ccv3.Warnings, error)
	PatchApplicationProcessHealthCheck(processGUID string, processHealthCheckType string, processHealthCheckEndpoint string) (ccv3.Process, ccv3.Warnings, error)
	PatchApplicationUserProvidedEnvironmentVariables(appGUID string, envVars ccv3.EnvironmentVariables) (ccv3.EnvironmentVariables, ccv3.Warnings, error)
	PatchOrganizationDefaultIsolationSegment(orgGUID string, isolationSegmentGUID string) (ccv3.Relationship, ccv3.Warnings, error)
	PollJob(jobURL string) (ccv3.Warnings, error)
	RevokeIsolationSegmentFromOrganization(isolationSegmentGUID string, organizationGUID string) (ccv3.Warnings, error)
	SetApplicationDroplet(appGUID string, dropletGUID string) (ccv3.Relationship, ccv3.Warnings, error)
	ShareServiceInstanceToSpaces(serviceInstanceGUID string, spaceGUIDs []string) (ccv3.RelationshipList, ccv3.Warnings, error)
	StartApplication(appGUID string) (ccv3.Application, ccv3.Warnings, error)
	StopApplication(appGUID string) (ccv3.Application, ccv3.Warnings, error)
	UnshareServiceInstanceFromSpace(serviceInstanceGUID string, spaceGUID string) (ccv3.Warnings, error)
	UpdateApplication(app ccv3.Application) (ccv3.Application, ccv3.Warnings, error)
	UpdateTask(taskGUID string) (ccv3.Task, ccv3.Warnings, error)
	UploadPackage(pkg ccv3.Package, zipFilepath string) (ccv3.Package, ccv3.Warnings, error)
}
