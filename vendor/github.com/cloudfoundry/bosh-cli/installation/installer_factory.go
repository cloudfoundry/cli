package installation

import (
	bideplrel "github.com/cloudfoundry/bosh-cli/deployment/release"
	biindex "github.com/cloudfoundry/bosh-cli/index"
	"github.com/cloudfoundry/bosh-cli/installation/blobextract"
	biinstallpkg "github.com/cloudfoundry/bosh-cli/installation/pkg"
	biregistry "github.com/cloudfoundry/bosh-cli/registry"
	bistatejob "github.com/cloudfoundry/bosh-cli/state/job"
	bistatepkg "github.com/cloudfoundry/bosh-cli/state/pkg"
	bitemplate "github.com/cloudfoundry/bosh-cli/templatescompiler"
	bierbrenderer "github.com/cloudfoundry/bosh-cli/templatescompiler/erbrenderer"
	biui "github.com/cloudfoundry/bosh-cli/ui"
	boshblob "github.com/cloudfoundry/bosh-utils/blobstore"
	boshcrypto "github.com/cloudfoundry/bosh-utils/crypto"
	boshcmd "github.com/cloudfoundry/bosh-utils/fileutil"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
)

type InstallerFactory interface {
	NewInstaller(Target) Installer
}

type installerFactory struct {
	ui                     biui.UI
	runner                 boshsys.CmdRunner
	extractor              boshcmd.Compressor
	releaseJobResolver     bideplrel.JobResolver
	uuidGenerator          boshuuid.Generator
	registryServerManager  biregistry.ServerManager
	logger                 boshlog.Logger
	logTag                 string
	fs                     boshsys.FileSystem
	digestCreateAlgorithms []boshcrypto.Algorithm
}

func NewInstallerFactory(
	ui biui.UI,
	runner boshsys.CmdRunner,
	extractor boshcmd.Compressor,
	releaseJobResolver bideplrel.JobResolver,
	uuidGenerator boshuuid.Generator,
	registryServerManager biregistry.ServerManager,
	logger boshlog.Logger,
	fs boshsys.FileSystem,
	digestCreateAlgorithms []boshcrypto.Algorithm,
) InstallerFactory {
	return &installerFactory{
		ui:                    ui,
		runner:                runner,
		extractor:             extractor,
		releaseJobResolver:    releaseJobResolver,
		uuidGenerator:         uuidGenerator,
		registryServerManager: registryServerManager,
		logger:                logger,
		logTag:                "installer",
		fs:                    fs,
		digestCreateAlgorithms: digestCreateAlgorithms,
	}
}

func (f *installerFactory) NewInstaller(target Target) Installer {
	context := &installerFactoryContext{
		target:             target,
		runner:             f.runner,
		logger:             f.logger,
		extractor:          f.extractor,
		uuidGenerator:      f.uuidGenerator,
		releaseJobResolver: f.releaseJobResolver,
		fs:                 f.fs,
		digestCreateAlgorithms: f.digestCreateAlgorithms,
	}

	return NewInstaller(
		target,
		context.JobRenderer(),
		context.JobResolver(),
		context.PackageCompiler(),
		context.BlobExtractor(),
		f.registryServerManager,
		f.logger,
	)
}

type installerFactoryContext struct {
	target             Target
	fs                 boshsys.FileSystem
	runner             boshsys.CmdRunner
	logger             boshlog.Logger
	extractor          boshcmd.Compressor
	uuidGenerator      boshuuid.Generator
	releaseJobResolver bideplrel.JobResolver

	jobDependencyCompiler  bistatejob.DependencyCompiler
	packageCompiler        bistatepkg.Compiler
	blobstore              boshblob.DigestBlobstore
	blobExtractor          blobextract.Extractor
	compiledPackageRepo    bistatepkg.CompiledPackageRepo
	digestCreateAlgorithms []boshcrypto.Algorithm
}

func (c *installerFactoryContext) JobRenderer() JobRenderer {

	erbRenderer := bierbrenderer.NewERBRenderer(c.fs, c.runner, c.logger)
	jobRenderer := bitemplate.NewJobRenderer(erbRenderer, c.fs, c.uuidGenerator, c.logger)
	jobListRenderer := bitemplate.NewJobListRenderer(jobRenderer, c.logger)

	return NewJobRenderer(
		jobListRenderer,
		c.extractor,
		c.Blobstore(),
	)
}

func (c *installerFactoryContext) PackageCompiler() PackageCompiler {
	return NewPackageCompiler(
		c.JobDependencyCompiler(),
		c.fs,
	)
}

func (c *installerFactoryContext) JobResolver() JobResolver {
	return NewJobResolver(c.releaseJobResolver)
}

func (c *installerFactoryContext) JobDependencyCompiler() bistatejob.DependencyCompiler {
	if c.jobDependencyCompiler != nil {
		return c.jobDependencyCompiler
	}

	c.jobDependencyCompiler = bistatejob.NewDependencyCompiler(
		c.InstallationStatePackageCompiler(),
		c.logger,
	)

	return c.jobDependencyCompiler
}

func (c *installerFactoryContext) InstallationStatePackageCompiler() bistatepkg.Compiler {
	if c.packageCompiler != nil {
		return c.packageCompiler
	}

	c.packageCompiler = biinstallpkg.NewPackageCompiler(
		c.runner,
		c.target.PackagesPath(),
		c.fs,
		c.extractor,
		c.Blobstore(),
		c.CompiledPackageRepo(),
		c.BlobExtractor(),
		c.logger,
	)

	return c.packageCompiler
}

func (c *installerFactoryContext) Blobstore() boshblob.DigestBlobstore {
	if c.blobstore != nil {
		return c.blobstore
	}

	options := map[string]interface{}{"blobstore_path": c.target.BlobstorePath()}
	localBlobstore := boshblob.NewLocalBlobstore(c.fs, c.uuidGenerator, options)
	c.blobstore = boshblob.NewDigestVerifiableBlobstore(localBlobstore, c.fs, c.digestCreateAlgorithms)

	return c.blobstore
}

func (c *installerFactoryContext) BlobExtractor() blobextract.Extractor {
	if c.blobExtractor != nil {
		return c.blobExtractor
	}

	c.blobExtractor = blobextract.NewExtractor(c.fs, c.extractor, c.Blobstore(), c.logger)

	return c.blobExtractor
}

func (c *installerFactoryContext) CompiledPackageRepo() bistatepkg.CompiledPackageRepo {
	if c.compiledPackageRepo != nil {
		return c.compiledPackageRepo
	}

	compiledPackageIndex := biindex.NewFileIndex(c.target.CompiledPackagedIndexPath(), c.fs)
	c.compiledPackageRepo = bistatepkg.NewCompiledPackageRepo(compiledPackageIndex)

	return c.compiledPackageRepo
}
