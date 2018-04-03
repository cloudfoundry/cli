package cmd

import (
	"os"
	"path/filepath"
	"time"

	"github.com/cppforlife/go-patch/patch"

	bihttpagent "github.com/cloudfoundry/bosh-agent/agentclient/http"
	biblobstore "github.com/cloudfoundry/bosh-cli/blobstore"
	bicloud "github.com/cloudfoundry/bosh-cli/cloud"
	biconfig "github.com/cloudfoundry/bosh-cli/config"
	bicpirel "github.com/cloudfoundry/bosh-cli/cpi/release"
	bidepl "github.com/cloudfoundry/bosh-cli/deployment"
	bidisk "github.com/cloudfoundry/bosh-cli/deployment/disk"
	biinstance "github.com/cloudfoundry/bosh-cli/deployment/instance"
	biinstancestate "github.com/cloudfoundry/bosh-cli/deployment/instance/state"
	bideplmanifest "github.com/cloudfoundry/bosh-cli/deployment/manifest"
	bideplrel "github.com/cloudfoundry/bosh-cli/deployment/release"
	bisshtunnel "github.com/cloudfoundry/bosh-cli/deployment/sshtunnel"
	bidepltpl "github.com/cloudfoundry/bosh-cli/deployment/template"
	bivm "github.com/cloudfoundry/bosh-cli/deployment/vm"
	boshtpl "github.com/cloudfoundry/bosh-cli/director/template"
	biindex "github.com/cloudfoundry/bosh-cli/index"
	boshinst "github.com/cloudfoundry/bosh-cli/installation"
	boshinstmanifest "github.com/cloudfoundry/bosh-cli/installation/manifest"
	bitarball "github.com/cloudfoundry/bosh-cli/installation/tarball"
	biregistry "github.com/cloudfoundry/bosh-cli/registry"
	boshrel "github.com/cloudfoundry/bosh-cli/release"
	birelsetmanifest "github.com/cloudfoundry/bosh-cli/release/set/manifest"
	bistatepkg "github.com/cloudfoundry/bosh-cli/state/pkg"
	bistemcell "github.com/cloudfoundry/bosh-cli/stemcell"
	bitemplate "github.com/cloudfoundry/bosh-cli/templatescompiler"
	bitemplateerb "github.com/cloudfoundry/bosh-cli/templatescompiler/erbrenderer"
	"github.com/cloudfoundry/bosh-utils/httpclient"
)

type envFactory struct {
	deps         BasicDeps
	manifestPath string
	manifestVars boshtpl.Variables
	manifestOp   patch.Op

	deploymentStateService     biconfig.DeploymentStateService
	installationManifestParser ReleaseSetAndInstallationManifestParser

	releaseManager  boshinst.ReleaseManager
	releaseFetcher  boshinst.ReleaseFetcher
	stemcellFetcher bistemcell.Fetcher

	cpiInstaller   bicpirel.CpiInstaller
	targetProvider boshinst.TargetProvider
	cloudFactory   bicloud.Factory

	diskManagerFactory     bidisk.ManagerFactory
	vmManagerFactory       bivm.ManagerFactory
	stemcellManagerFactory bistemcell.ManagerFactory

	instanceManagerFactory   biinstance.ManagerFactory
	deploymentManagerFactory bidepl.ManagerFactory

	agentClientFactory bihttpagent.AgentClientFactory
	blobstoreFactory   biblobstore.Factory
	deploymentFactory  bidepl.Factory
	deploymentRecord   bidepl.Record
}

func NewEnvFactory(deps BasicDeps, manifestPath string, statePath string, manifestVars boshtpl.Variables, manifestOp patch.Op) *envFactory {
	f := envFactory{
		deps:         deps,
		manifestPath: manifestPath,
		manifestVars: manifestVars,
		manifestOp:   manifestOp,
	}

	f.releaseManager = boshinst.NewReleaseManager(deps.Logger)
	releaseJobResolver := bideplrel.NewJobResolver(f.releaseManager)

	// todo expand path?
	workspaceRootPath := filepath.Join(os.Getenv("HOME"), ".bosh")

	{
		tarballCacheBasePath := filepath.Join(workspaceRootPath, "downloads")
		tarballCache := bitarball.NewCache(tarballCacheBasePath, deps.FS, deps.Logger)
		httpClient := httpclient.NewHTTPClient(httpclient.CreateDefaultClient(nil), deps.Logger)
		tarballProvider := bitarball.NewProvider(
			tarballCache, deps.FS, httpClient, 3, 500*time.Millisecond, deps.Logger)

		releaseProvider := boshrel.NewProvider(
			deps.CmdRunner, deps.Compressor, deps.DigestCalculator, deps.FS, deps.Logger)

		f.releaseFetcher = boshinst.NewReleaseFetcher(
			tarballProvider,
			releaseProvider.NewExtractingArchiveReader(),
			f.releaseManager,
		)

		stemcellReader := bistemcell.NewReader(deps.Compressor, deps.FS)
		stemcellExtractor := bistemcell.NewExtractor(stemcellReader, deps.FS)

		f.stemcellFetcher = bistemcell.Fetcher{
			TarballProvider:   tarballProvider,
			StemcellExtractor: stemcellExtractor,
		}
	}

	f.deploymentStateService = biconfig.NewFileSystemDeploymentStateService(
		deps.FS, deps.UUIDGen, deps.Logger, biconfig.DeploymentStatePath(manifestPath, statePath))

	{
		registryServer := biregistry.NewServerManager(deps.Logger)
		installerFactory := boshinst.NewInstallerFactory(
			deps.UI, deps.CmdRunner, deps.Compressor, releaseJobResolver,
			deps.UUIDGen, registryServer, deps.Logger, deps.FS, deps.DigestCreationAlgorithms)

		f.cpiInstaller = bicpirel.CpiInstaller{
			ReleaseManager:   f.releaseManager,
			InstallerFactory: installerFactory,
			Validator:        bicpirel.NewValidator(),
		}
	}

	f.targetProvider = boshinst.NewTargetProvider(
		f.deploymentStateService, deps.UUIDGen, filepath.Join(workspaceRootPath, "installations"))

	{
		diskRepo := biconfig.NewDiskRepo(f.deploymentStateService, deps.UUIDGen)
		stemcellRepo := biconfig.NewStemcellRepo(f.deploymentStateService, deps.UUIDGen)
		vmRepo := biconfig.NewVMRepo(f.deploymentStateService)

		f.diskManagerFactory = bidisk.NewManagerFactory(diskRepo, deps.Logger)
		diskDeployer := bivm.NewDiskDeployer(f.diskManagerFactory, diskRepo, deps.Logger)

		f.stemcellManagerFactory = bistemcell.NewManagerFactory(stemcellRepo)
		f.vmManagerFactory = bivm.NewManagerFactory(
			vmRepo, stemcellRepo, diskDeployer, deps.UUIDGen, deps.FS, deps.Logger)

		deploymentRepo := biconfig.NewDeploymentRepo(f.deploymentStateService)
		releaseRepo := biconfig.NewReleaseRepo(f.deploymentStateService, deps.UUIDGen)
		f.deploymentRecord = bidepl.NewRecord(deploymentRepo, releaseRepo, stemcellRepo)
	}

	{
		f.blobstoreFactory = biblobstore.NewBlobstoreFactory(deps.UUIDGen, deps.FS, deps.Logger)
		f.deploymentFactory = bidepl.NewFactory(10*time.Second, 500*time.Millisecond)
		f.agentClientFactory = bihttpagent.NewAgentClientFactory(1*time.Second, deps.Logger)
		f.cloudFactory = bicloud.NewFactory(deps.FS, deps.CmdRunner, deps.Logger)
	}

	{
		erbRenderer := bitemplateerb.NewERBRenderer(deps.FS, deps.CmdRunner, deps.Logger)
		jobRenderer := bitemplate.NewJobRenderer(erbRenderer, deps.FS, deps.UUIDGen, deps.Logger)

		builderFactory := biinstancestate.NewBuilderFactory(
			bistatepkg.NewCompiledPackageRepo(biindex.NewInMemoryIndex()),
			releaseJobResolver,
			bitemplate.NewJobListRenderer(jobRenderer, deps.Logger),
			bitemplate.NewRenderedJobListCompressor(deps.FS, deps.Compressor, deps.DigestCalculator, deps.Logger),
			deps.Logger,
		)

		sshTunnelFactory := bisshtunnel.NewFactory(deps.Logger)
		instanceFactory := biinstance.NewFactory(builderFactory)

		f.instanceManagerFactory = biinstance.NewManagerFactory(
			sshTunnelFactory, instanceFactory, deps.Logger)
	}

	{
		releaseSetValidator := birelsetmanifest.NewValidator(deps.Logger)
		releaseSetParser := birelsetmanifest.NewParser(deps.FS, deps.Logger, releaseSetValidator)

		installValidator := boshinstmanifest.NewValidator(deps.Logger)
		installParser := boshinstmanifest.NewParser(deps.FS, deps.UUIDGen, deps.Logger, installValidator)

		f.installationManifestParser = ReleaseSetAndInstallationManifestParser{
			ReleaseSetParser:   releaseSetParser,
			InstallationParser: installParser,
		}
	}

	return &f
}

func (f *envFactory) Preparer() DeploymentPreparer {
	return NewDeploymentPreparer(
		f.deps.UI,
		f.deps.Logger,
		"DeploymentPreparer",
		f.deploymentStateService,
		biconfig.NewLegacyDeploymentStateMigrator(
			f.deploymentStateService,
			f.deps.FS,
			f.deps.UUIDGen,
			f.deps.Logger,
		),
		f.releaseManager,
		f.deploymentRecord,
		f.cloudFactory,
		f.stemcellManagerFactory,
		f.agentClientFactory,
		f.vmManagerFactory,
		f.blobstoreFactory,
		bidepl.NewDeployer(
			f.vmManagerFactory,
			f.instanceManagerFactory,
			f.deploymentFactory,
			f.deps.Logger,
		),
		f.manifestPath,
		f.manifestVars,
		f.manifestOp,
		f.cpiInstaller,
		f.releaseFetcher,
		f.stemcellFetcher,
		f.installationManifestParser,
		NewDeploymentManifestParser(
			bideplmanifest.NewParser(f.deps.FS, f.deps.Logger),
			bideplmanifest.NewValidator(f.deps.Logger),
			f.releaseManager,
			bidepltpl.NewDeploymentTemplateFactory(f.deps.FS),
		),
		NewTempRootConfigurator(f.deps.FS),
		f.targetProvider,
	)
}

func (f *envFactory) Deleter() DeploymentDeleter {
	return NewDeploymentDeleter(
		f.deps.UI,
		"DeploymentDeleter",
		f.deps.Logger,
		f.deploymentStateService,
		f.releaseManager,
		f.cloudFactory,
		f.agentClientFactory,
		f.blobstoreFactory,
		bidepl.NewManagerFactory(
			f.vmManagerFactory,
			f.instanceManagerFactory,
			f.diskManagerFactory,
			f.stemcellManagerFactory,
			f.deploymentFactory,
		),
		f.manifestPath,
		f.manifestVars,
		f.manifestOp,
		f.cpiInstaller,
		boshinst.NewUninstaller(f.deps.FS, f.deps.Logger),
		f.releaseFetcher,
		f.installationManifestParser,
		NewTempRootConfigurator(f.deps.FS),
		f.targetProvider,
	)
}
