package v3action

import (
	"io"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . CloudControllerClient

// CloudControllerClient is the interface to the cloud controller V3 API.
type CloudControllerClient interface {
	AppSSHEndpoint() string
	AppSSHHostKeyFingerprint() string
	CancelDeployment(deploymentGUID string) (ccv3.Warnings, error)
	CloudControllerAPIVersion() string
	CreateApplication(app resources.Application) (resources.Application, ccv3.Warnings, error)
	CreateApplicationDeployment(appGUID string, dropletGUID string) (string, ccv3.Warnings, error)
	CreateApplicationProcessScale(appGUID string, process resources.Process) (resources.Process, ccv3.Warnings, error)
	CreateApplicationTask(appGUID string, task resources.Task) (resources.Task, ccv3.Warnings, error)
	CreateBuild(build resources.Build) (resources.Build, ccv3.Warnings, error)
	CreateIsolationSegment(isolationSegment resources.IsolationSegment) (resources.IsolationSegment, ccv3.Warnings, error)
	CreatePackage(pkg resources.Package) (resources.Package, ccv3.Warnings, error)
	DeleteApplication(guid string) (ccv3.JobURL, ccv3.Warnings, error)
	DeleteApplicationProcessInstance(appGUID string, processType string, instanceIndex int) (ccv3.Warnings, error)
	DeleteIsolationSegment(guid string) (ccv3.Warnings, error)
	DeleteIsolationSegmentOrganization(isolationSegmentGUID string, organizationGUID string) (ccv3.Warnings, error)
	UnshareServiceInstanceFromSpace(serviceInstanceGUID string, sharedToSpaceGUID string) (ccv3.Warnings, error)
	EntitleIsolationSegmentToOrganizations(isoGUID string, orgGUIDs []string) (resources.RelationshipList, ccv3.Warnings, error)
	GetApplicationDropletCurrent(appGUID string) (resources.Droplet, ccv3.Warnings, error)
	GetApplicationEnvironment(appGUID string) (ccv3.Environment, ccv3.Warnings, error)
	GetApplicationProcessByType(appGUID string, processType string) (resources.Process, ccv3.Warnings, error)
	GetApplicationProcesses(appGUID string) ([]resources.Process, ccv3.Warnings, error)
	GetApplications(query ...ccv3.Query) ([]resources.Application, ccv3.Warnings, error)
	GetApplicationTasks(appGUID string, query ...ccv3.Query) ([]resources.Task, ccv3.Warnings, error)
	GetBuild(guid string) (resources.Build, ccv3.Warnings, error)
	GetDeployment(guid string) (resources.Deployment, ccv3.Warnings, error)
	GetDeployments(query ...ccv3.Query) ([]resources.Deployment, ccv3.Warnings, error)
	GetDroplet(guid string) (resources.Droplet, ccv3.Warnings, error)
	GetDroplets(query ...ccv3.Query) ([]resources.Droplet, ccv3.Warnings, error)
	GetInfo() (ccv3.Info, ccv3.Warnings, error)
	GetIsolationSegment(guid string) (resources.IsolationSegment, ccv3.Warnings, error)
	GetIsolationSegmentOrganizations(isolationSegmentGUID string) ([]resources.Organization, ccv3.Warnings, error)
	GetIsolationSegments(query ...ccv3.Query) ([]resources.IsolationSegment, ccv3.Warnings, error)
	GetOrganizationDefaultIsolationSegment(orgGUID string) (resources.Relationship, ccv3.Warnings, error)
	GetOrganizations(query ...ccv3.Query) ([]resources.Organization, ccv3.Warnings, error)
	GetPackage(guid string) (resources.Package, ccv3.Warnings, error)
	GetPackages(query ...ccv3.Query) ([]resources.Package, ccv3.Warnings, error)
	GetProcessInstances(processGUID string) ([]ccv3.ProcessInstance, ccv3.Warnings, error)
	GetServiceInstances(query ...ccv3.Query) ([]resources.ServiceInstance, ccv3.IncludedResources, ccv3.Warnings, error)
	GetSpaceIsolationSegment(spaceGUID string) (resources.Relationship, ccv3.Warnings, error)
	GetSpaces(query ...ccv3.Query) ([]resources.Space, ccv3.IncludedResources, ccv3.Warnings, error)
	PollJob(jobURL ccv3.JobURL) (ccv3.Warnings, error)
	SetApplicationDroplet(appGUID string, dropletGUID string) (resources.Relationship, ccv3.Warnings, error)
	TargetCF(settings ccv3.TargetSettings)
	UpdateApplication(app resources.Application) (resources.Application, ccv3.Warnings, error)
	UpdateApplicationApplyManifest(appGUID string, rawManifest []byte) (ccv3.JobURL, ccv3.Warnings, error)
	UpdateApplicationEnvironmentVariables(appGUID string, envVars resources.EnvironmentVariables) (resources.EnvironmentVariables, ccv3.Warnings, error)
	UpdateApplicationRestart(appGUID string) (resources.Application, ccv3.Warnings, error)
	UpdateApplicationStart(appGUID string) (resources.Application, ccv3.Warnings, error)
	UpdateApplicationStop(appGUID string) (resources.Application, ccv3.Warnings, error)
	UpdateOrganizationDefaultIsolationSegmentRelationship(orgGUID string, isolationSegmentGUID string) (resources.Relationship, ccv3.Warnings, error)
	UpdateProcess(process resources.Process) (resources.Process, ccv3.Warnings, error)
	UpdateSpaceIsolationSegmentRelationship(spaceGUID string, isolationSegmentGUID string) (resources.Relationship, ccv3.Warnings, error)
	UpdateTaskCancel(taskGUID string) (resources.Task, ccv3.Warnings, error)
	UploadBitsPackage(pkg resources.Package, matchedResources []ccv3.Resource, newResources io.Reader, newResourcesLength int64) (resources.Package, ccv3.Warnings, error)
	UploadDropletBits(dropletGUID string, dropletPath string, droplet io.Reader, dropletLength int64) (ccv3.JobURL, ccv3.Warnings, error)
	UploadPackage(pkg resources.Package, zipFilepath string) (resources.Package, ccv3.Warnings, error)
}
