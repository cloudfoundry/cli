package installation

import (
	"fmt"
	"path/filepath"

	"github.com/cloudfoundry/bosh-cli/installation/blobextract"
	biinstallmanifest "github.com/cloudfoundry/bosh-cli/installation/manifest"
	biregistry "github.com/cloudfoundry/bosh-cli/registry"
	biui "github.com/cloudfoundry/bosh-cli/ui"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type InstalledJob struct {
	RenderedJobRef
	Path string
}

func NewInstalledJob(ref RenderedJobRef, path string) InstalledJob {
	return InstalledJob{RenderedJobRef: ref, Path: path}
}

type Installer interface {
	Install(biinstallmanifest.Manifest, biui.Stage) (Installation, error)
	Cleanup(Installation) error
}

type installer struct {
	target                Target
	jobRenderer           JobRenderer
	jobResolver           JobResolver
	packageCompiler       PackageCompiler
	blobExtractor         blobextract.Extractor
	registryServerManager biregistry.ServerManager
	logger                boshlog.Logger
	logTag                string
}

func NewInstaller(
	target Target,
	jobRenderer JobRenderer,
	jobResolver JobResolver,
	packageCompiler PackageCompiler,
	blobExtractor blobextract.Extractor,
	registryServerManager biregistry.ServerManager,
	logger boshlog.Logger,
) Installer {
	return &installer{
		target:                target,
		jobRenderer:           jobRenderer,
		jobResolver:           jobResolver,
		packageCompiler:       packageCompiler,
		blobExtractor:         blobExtractor,
		registryServerManager: registryServerManager,
		logger:                logger,
		logTag:                "installer",
	}
}

func (i *installer) Install(manifest biinstallmanifest.Manifest, stage biui.Stage) (Installation, error) {
	i.logger.Info(i.logTag, "Installing CPI deployment '%s'", manifest.Name)
	i.logger.Debug(i.logTag, "Installing CPI deployment '%s' with manifest: %#v", manifest.Name, manifest)

	jobs, err := i.jobResolver.From(manifest)
	if err != nil {
		return nil, bosherr.WrapError(err, "Resolving jobs from manifest")
	}

	compiledPackages, err := i.packageCompiler.For(jobs, stage)
	if err != nil {
		return nil, err
	}

	err = stage.Perform("Installing packages", func() error {
		return i.installPackages(compiledPackages)
	})
	if err != nil {
		return nil, err
	}

	renderedJobRefs, err := i.jobRenderer.RenderAndUploadFrom(manifest, jobs, stage)
	if err != nil {
		return nil, bosherr.WrapError(err, "Rendering and uploading Jobs")
	}

	renderedCPIJob := renderedJobRefs[0]
	installedJob, err := i.installJob(renderedCPIJob, stage)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Installing job '%s' for CPI release", renderedCPIJob.Name)
	}

	return NewInstallation(
		i.target,
		installedJob,
		manifest,
		i.registryServerManager,
	), nil
}

func (i *installer) Cleanup(installation Installation) error {
	job := installation.Job()
	return i.blobExtractor.Cleanup(job.BlobstoreID, job.Path)
}

func (i *installer) installPackages(compiledPackages []CompiledPackageRef) error {
	for _, pkg := range compiledPackages {
		err := i.blobExtractor.Extract(pkg.BlobstoreID, pkg.SHA1, filepath.Join(i.target.PackagesPath(), pkg.Name))
		if err != nil {
			return bosherr.WrapErrorf(err, "Installing package '%s'", pkg.Name)
		}
	}
	return nil
}

func (i *installer) installJob(renderedJobRef RenderedJobRef, stage biui.Stage) (installedJob InstalledJob, err error) {
	err = stage.Perform(fmt.Sprintf("Installing job '%s'", renderedJobRef.Name), func() error {
		var stageErr error
		jobDir := filepath.Join(i.target.JobsPath(), renderedJobRef.Name)

		stageErr = i.blobExtractor.Extract(renderedJobRef.BlobstoreID, renderedJobRef.SHA1, jobDir)
		if stageErr != nil {
			return bosherr.WrapErrorf(stageErr, "Extracting blob with ID '%s'", renderedJobRef.BlobstoreID)
		}

		stageErr = i.blobExtractor.ChmodExecutables(filepath.Join(jobDir, "bin", "*"))
		if stageErr != nil {
			return bosherr.WrapErrorf(stageErr, "Chmoding binaries for '%s'", jobDir)
		}

		installedJob = NewInstalledJob(renderedJobRef, jobDir)
		return nil
	})
	return installedJob, err
}
