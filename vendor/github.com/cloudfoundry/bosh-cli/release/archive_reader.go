package release

import (
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshcmd "github.com/cloudfoundry/bosh-utils/fileutil"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	boshjob "github.com/cloudfoundry/bosh-cli/release/job"
	boshlic "github.com/cloudfoundry/bosh-cli/release/license"
	boshman "github.com/cloudfoundry/bosh-cli/release/manifest"
	boshpkg "github.com/cloudfoundry/bosh-cli/release/pkg"
	. "github.com/cloudfoundry/bosh-cli/release/resource"
)

type ArchiveReader struct {
	jobArchiveReader boshjob.ArchiveReader
	pkgArchiveReader boshpkg.ArchiveReader

	compressor boshcmd.Compressor
	fs         boshsys.FileSystem

	logTag string
	logger boshlog.Logger
}

func NewArchiveReader(
	jobArchiveReader boshjob.ArchiveReader,
	pkgArchiveReader boshpkg.ArchiveReader,
	compressor boshcmd.Compressor,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) ArchiveReader {
	return ArchiveReader{
		jobArchiveReader: jobArchiveReader,
		pkgArchiveReader: pkgArchiveReader,

		compressor: compressor,
		fs:         fs,

		logTag: "release.ArchiveReader",
		logger: logger,
	}
}

func (r ArchiveReader) Read(path string) (Release, error) {
	extractPath, err := r.fs.TempDir("bosh-release")
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Creating temp directory to extract release '%s'", path)
	}

	r.logger.Info(r.logTag, "Extracting release tarball '%s' to '%s'", path, extractPath)

	err = r.compressor.DecompressFileToDir(path, extractPath, boshcmd.CompressorOptions{})
	if err != nil {
		r.cleanUp(extractPath)
		return nil, bosherr.WrapError(err, "Extracting release")
	}

	manifestPath := filepath.Join(extractPath, "release.MF")

	manifest, err := boshman.NewManifestFromPath(manifestPath, r.fs)
	if err != nil {
		r.cleanUp(extractPath)
		return nil, err
	}

	release, err := r.newRelease(manifest, extractPath)
	if err != nil {
		r.cleanUp(extractPath)
		return nil, bosherr.WrapError(err, "Constructing release from manifest")
	}

	return release, nil
}

func (r ArchiveReader) cleanUp(extractPath string) {
	removeErr := r.fs.RemoveAll(extractPath)
	if removeErr != nil {
		r.logger.Error(r.logTag, "Failed to remove extracted release: %s", removeErr.Error())
	}
}

func (r ArchiveReader) newRelease(manifest boshman.Manifest, extractPath string) (Release, error) {
	var errs []error

	packages, err := r.newPackages(manifest.Packages, extractPath)
	if err != nil {
		errs = append(errs, bosherr.WrapError(err, "Constructing packages from manifest"))
	}

	compiledPkgs, err := r.newCompiledPackages(manifest.CompiledPkgs, extractPath)
	if err != nil {
		errs = append(errs, bosherr.WrapError(err, "Constructing compiled packages from manifest"))
	}

	jobs, err := r.newJobs(r.newCombinedPackages(packages, compiledPkgs), manifest.Jobs, extractPath)
	if err != nil {
		errs = append(errs, bosherr.WrapError(err, "Constructing jobs from manifest"))
	}

	license := r.newLicense(manifest.License, extractPath)

	if len(errs) > 0 {
		return nil, bosherr.NewMultiError(errs...)
	}

	release := &release{
		name:    manifest.Name,
		version: manifest.Version,

		commitHash:         manifest.CommitHash,
		uncommittedChanges: manifest.UncommittedChanges,

		jobs:         jobs,
		packages:     packages,
		compiledPkgs: compiledPkgs,
		license:      license,

		extractedPath: extractPath,
		fs:            r.fs,
	}

	return release, nil
}

func (r ArchiveReader) newJobs(pkgs []boshpkg.Compilable, refs []boshman.JobRef, extractPath string) ([]*boshjob.Job, error) {
	var jobs []*boshjob.Job
	var errs []error

	for _, ref := range refs {
		archivePath := filepath.Join(extractPath, "jobs", ref.Name+".tgz")

		job, err := r.jobArchiveReader.Read(ref, archivePath)
		if err != nil {
			errs = append(errs, bosherr.WrapErrorf(err, "Reading job '%s' from archive", ref.Name))
			continue
		}

		err = job.AttachCompilablePackages(pkgs)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		jobs = append(jobs, job)
	}

	if len(errs) > 0 {
		return nil, bosherr.NewMultiError(errs...)
	}

	return jobs, nil
}

func (r ArchiveReader) newPackages(refs []boshman.PackageRef, extractPath string) ([]*boshpkg.Package, error) {
	var packages []*boshpkg.Package
	var errs []error

	for _, ref := range refs {
		archivePath := filepath.Join(extractPath, "packages", ref.Name+".tgz")

		pkg, err := r.pkgArchiveReader.Read(ref, archivePath)
		if err != nil {
			errs = append(errs, bosherr.WrapErrorf(err, "Reading package '%s' from archive", ref.Name))
			continue
		}

		packages = append(packages, pkg)
	}

	for _, pkg := range packages {
		err := pkg.AttachDependencies(packages)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return nil, bosherr.NewMultiError(errs...)
	}

	return packages, nil
}

func (r ArchiveReader) newCompiledPackages(refs []boshman.CompiledPackageRef, extractPath string) ([]*boshpkg.CompiledPackage, error) {
	var compiledPkgs []*boshpkg.CompiledPackage
	var errs []error

	for _, ref := range refs {
		archivePath := filepath.Join(extractPath, "compiled_packages", ref.Name+".tgz")

		compiledPkg := boshpkg.NewCompiledPackageWithArchive(
			ref.Name, ref.Fingerprint, ref.OSVersionSlug, archivePath, ref.SHA1, ref.Dependencies)

		compiledPkgs = append(compiledPkgs, compiledPkg)
	}

	for _, compiledPkg := range compiledPkgs {
		err := compiledPkg.AttachDependencies(compiledPkgs)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return nil, bosherr.NewMultiError(errs...)
	}

	return compiledPkgs, nil
}

func (r ArchiveReader) newCombinedPackages(pkgs []*boshpkg.Package, compiledPkgs []*boshpkg.CompiledPackage) []boshpkg.Compilable {
	var coms []boshpkg.Compilable
	for _, pkg := range pkgs {
		coms = append(coms, pkg)
	}
	for _, compiledPkg := range compiledPkgs {
		coms = append(coms, compiledPkg)
	}
	return coms
}

func (r ArchiveReader) newLicense(ref *boshman.LicenseRef, extractPath string) *boshlic.License {
	if ref != nil {
		archivePath := filepath.Join(extractPath, "license.tgz")

		if r.fs.FileExists(archivePath) {
			resource := NewResourceWithBuiltArchive(
				"license", ref.Fingerprint, archivePath, ref.SHA1)

			return boshlic.NewLicense(resource)
		}
	}

	return nil
}
