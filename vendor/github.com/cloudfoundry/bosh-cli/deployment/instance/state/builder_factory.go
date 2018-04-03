package state

import (
	biagentclient "github.com/cloudfoundry/bosh-agent/agentclient"
	biblobstore "github.com/cloudfoundry/bosh-cli/blobstore"
	bideplrel "github.com/cloudfoundry/bosh-cli/deployment/release"
	bistatejob "github.com/cloudfoundry/bosh-cli/state/job"
	bistatepkg "github.com/cloudfoundry/bosh-cli/state/pkg"
	bitemplate "github.com/cloudfoundry/bosh-cli/templatescompiler"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type BuilderFactory interface {
	NewBuilder(biblobstore.Blobstore, biagentclient.AgentClient) Builder
}

type builderFactory struct {
	packageRepo               bistatepkg.CompiledPackageRepo
	releaseJobResolver        bideplrel.JobResolver
	jobRenderer               bitemplate.JobListRenderer
	renderedJobListCompressor bitemplate.RenderedJobListCompressor
	logger                    boshlog.Logger
}

func NewBuilderFactory(
	packageRepo bistatepkg.CompiledPackageRepo,
	releaseJobResolver bideplrel.JobResolver,
	jobRenderer bitemplate.JobListRenderer,
	renderedJobListCompressor bitemplate.RenderedJobListCompressor,
	logger boshlog.Logger,
) BuilderFactory {
	return &builderFactory{
		packageRepo:               packageRepo,
		releaseJobResolver:        releaseJobResolver,
		jobRenderer:               jobRenderer,
		renderedJobListCompressor: renderedJobListCompressor,
		logger: logger,
	}
}

func (f *builderFactory) NewBuilder(blobstore biblobstore.Blobstore, agentClient biagentclient.AgentClient) Builder {
	packageCompiler := NewRemotePackageCompiler(blobstore, agentClient, f.packageRepo)
	jobDependencyCompiler := bistatejob.NewDependencyCompiler(packageCompiler, f.logger)

	return NewBuilder(
		f.releaseJobResolver,
		jobDependencyCompiler,
		f.jobRenderer,
		f.renderedJobListCompressor,
		blobstore,
		f.logger,
	)
}
