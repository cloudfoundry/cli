package state

import (
	"errors"

	agentclient "github.com/cloudfoundry/bosh-agent/agentclient"

	biblobstore "github.com/cloudfoundry/bosh-cli/blobstore"
	bideplmanifest "github.com/cloudfoundry/bosh-cli/deployment/manifest"
	bideplrel "github.com/cloudfoundry/bosh-cli/deployment/release"
	bireljob "github.com/cloudfoundry/bosh-cli/release/job"
	bistatejob "github.com/cloudfoundry/bosh-cli/state/job"
	bitemplate "github.com/cloudfoundry/bosh-cli/templatescompiler"
	biui "github.com/cloudfoundry/bosh-cli/ui"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
)

type Builder interface {
	Build(jobName string, instanceID int, deploymentManifest bideplmanifest.Manifest, stage biui.Stage, agentState agentclient.AgentState) (State, error)
	BuildInitialState(jobName string, instanceID int, deploymentManifest bideplmanifest.Manifest) (State, error)
}

type builder struct {
	releaseJobResolver        bideplrel.JobResolver
	jobDependencyCompiler     bistatejob.DependencyCompiler
	jobListRenderer           bitemplate.JobListRenderer
	renderedJobListCompressor bitemplate.RenderedJobListCompressor
	blobstore                 biblobstore.Blobstore
	logger                    boshlog.Logger
	logTag                    string
}

func NewBuilder(
	releaseJobResolver bideplrel.JobResolver,
	jobDependencyCompiler bistatejob.DependencyCompiler,
	jobListRenderer bitemplate.JobListRenderer,
	renderedJobListCompressor bitemplate.RenderedJobListCompressor,
	blobstore biblobstore.Blobstore,
	logger boshlog.Logger,
) Builder {
	return &builder{
		releaseJobResolver:        releaseJobResolver,
		jobDependencyCompiler:     jobDependencyCompiler,
		jobListRenderer:           jobListRenderer,
		renderedJobListCompressor: renderedJobListCompressor,
		blobstore:                 blobstore,
		logger:                    logger,
		logTag:                    "instanceStateBuilder",
	}
}

type renderedJobs struct {
	BlobstoreID string
	Archive     bitemplate.RenderedJobListArchive
}

func (b *builder) Build(jobName string, instanceID int, deploymentManifest bideplmanifest.Manifest, stage biui.Stage, agentState agentclient.AgentState) (State, error) {

	initialState, err := b.BuildInitialState(jobName, instanceID, deploymentManifest)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Building initial state '%s", jobName)
	}

	deploymentJob, found := deploymentManifest.FindJobByName(jobName)
	if !found {
		return nil, bosherr.Errorf("Job '%s' not found in deployment manifest", jobName)
	}

	releaseJobs, err := b.resolveJobs(deploymentJob.Templates)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Resolving jobs for instance '%s/%d'", jobName, instanceID)
	}

	releaseJobProperties := make(map[string]*biproperty.Map)
	for _, releaseJob := range deploymentJob.Templates {
		releaseJobProperties[releaseJob.Name] = releaseJob.Properties
	}

	defaultAddress, err := b.defaultAddress(initialState.NetworkInterfaces(), agentState)
	if err != nil {
		return nil, err
	}

	renderedJobTemplates, err := b.renderJobTemplates(releaseJobs, releaseJobProperties, deploymentJob.Properties, deploymentManifest.Properties, deploymentManifest.Name, defaultAddress, stage)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Rendering job templates for instance '%s/%d'", jobName, instanceID)
	}

	compiledPackageRefs, err := b.jobDependencyCompiler.Compile(releaseJobs, stage)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Compiling job package dependencies for instance '%s/%d'", jobName, instanceID)
	}

	compiledDeploymentPackageRefs := make([]PackageRef, len(compiledPackageRefs), len(compiledPackageRefs))

	for i, compiledPackageRef := range compiledPackageRefs {
		compiledDeploymentPackageRefs[i] = PackageRef{
			Name:    compiledPackageRef.Name,
			Version: compiledPackageRef.Version,
			Archive: BlobRef{
				BlobstoreID: compiledPackageRef.BlobstoreID,
				SHA1:        compiledPackageRef.SHA1,
			},
		}
	}

	// convert array to array
	renderedJobRefs := make([]JobRef, len(releaseJobs), len(releaseJobs))

	for i, releaseJob := range releaseJobs {
		renderedJobRefs[i] = JobRef{
			Name:    releaseJob.Name(),
			Version: releaseJob.Fingerprint(),
		}
	}

	renderedJobListArchiveBlobRef := BlobRef{
		BlobstoreID: renderedJobTemplates.BlobstoreID,
		SHA1:        renderedJobTemplates.Archive.SHA1(),
	}

	return &state{
		deploymentName:         deploymentManifest.Name,
		name:                   jobName,
		id:                     instanceID,
		networks:               initialState.NetworkInterfaces(),
		compiledPackages:       compiledDeploymentPackageRefs,
		renderedJobs:           renderedJobRefs,
		renderedJobListArchive: renderedJobListArchiveBlobRef,
	}, nil
}

func (b *builder) BuildInitialState(jobName string, instanceID int, deploymentManifest bideplmanifest.Manifest) (State, error) {
	deploymentJob, found := deploymentManifest.FindJobByName(jobName)
	if !found {
		return nil, bosherr.Errorf("Job '%s' not found in deployment manifest", jobName)
	}

	networkInterfaces, err := deploymentManifest.NetworkInterfaces(deploymentJob.Name)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Finding networks for job '%s", jobName)
	}

	// convert map to array
	networkRefs := make([]NetworkRef, 0, len(networkInterfaces))
	for networkName, networkInterface := range networkInterfaces {
		genericMap := make(map[string]interface{}, len(networkInterface))

		for k, v := range networkInterface {
			genericMap[k] = v
		}

		networkRefs = append(networkRefs, NetworkRef{
			Name:      networkName,
			Interface: genericMap,
		})
	}

	return &state{
		deploymentName: deploymentManifest.Name,
		name:           jobName,
		id:             instanceID,
		networks:       networkRefs,
	}, nil
}

// FIXME: why do i exist here and in installation/state/builder.go??
func (b *builder) resolveJobs(jobRefs []bideplmanifest.ReleaseJobRef) ([]bireljob.Job, error) {
	releaseJobs := make([]bireljob.Job, len(jobRefs), len(jobRefs))
	for i, jobRef := range jobRefs {
		release, err := b.releaseJobResolver.Resolve(jobRef.Name, jobRef.Release)
		if err != nil {
			return releaseJobs, bosherr.Errorf("Resolving job '%s' in release '%s'", jobRef.Name, jobRef.Release)
		}
		releaseJobs[i] = release
	}
	return releaseJobs, nil
}

// renderJobTemplates renders all the release job templates for multiple release jobs specified by a deployment job
func (b *builder) renderJobTemplates(
	releaseJobs []bireljob.Job,
	releaseJobProperties map[string]*biproperty.Map,
	jobProperties biproperty.Map,
	globalProperties biproperty.Map,
	deploymentName string,
	address string,
	stage biui.Stage,
) (renderedJobs, error) {
	var (
		renderedJobListArchive bitemplate.RenderedJobListArchive
		blobID                 string
	)
	err := stage.Perform("Rendering job templates", func() error {
		renderedJobList, err := b.jobListRenderer.Render(releaseJobs, releaseJobProperties, jobProperties, globalProperties, deploymentName, address)
		if err != nil {
			return err
		}
		defer renderedJobList.DeleteSilently()

		renderedJobListArchive, err = b.renderedJobListCompressor.Compress(renderedJobList)
		if err != nil {
			return bosherr.WrapError(err, "Compressing rendered job templates")
		}
		defer renderedJobListArchive.DeleteSilently()

		blobID, err = b.blobstore.Add(renderedJobListArchive.Path())
		if err != nil {
			return bosherr.WrapErrorf(err, "Uploading rendered job template archive '%s' to the blobstore", renderedJobListArchive.Path())
		}

		return nil
	})
	if err != nil {
		return renderedJobs{}, err
	}

	return renderedJobs{
		BlobstoreID: blobID,
		Archive:     renderedJobListArchive,
	}, nil
}

func (b *builder) defaultAddress(networkRefs []NetworkRef, agentState agentclient.AgentState) (string, error) {

	if (networkRefs == nil) || (len(networkRefs) == 0) {
		return "", errors.New("Must specify network")
	}

	if len(networkRefs) == 1 {
		return networkIp(networkRefs[0], agentState), nil
	}

	for _, ref := range networkRefs {
		if ref.Interface["default"] == nil {
			continue
		}

		for _, val := range ref.Interface["default"].([]bideplmanifest.NetworkDefault) {
			if val == "gateway" {
				return networkIp(ref, agentState), nil
			}
		}
	}

	return "", errors.New("Must specify default network")
}

func networkIp(networkRef NetworkRef, agentState agentclient.AgentState) string {
	if "dynamic" == networkRef.Interface["type"].(string) {
		return agentState.NetworkSpecs[networkRef.Name].IP
	}

	return networkRef.Interface["ip"].(string)
}
