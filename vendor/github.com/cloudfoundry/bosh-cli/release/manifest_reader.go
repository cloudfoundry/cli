package release

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	boshjob "github.com/cloudfoundry/bosh-cli/release/job"
	boshlic "github.com/cloudfoundry/bosh-cli/release/license"
	boshman "github.com/cloudfoundry/bosh-cli/release/manifest"
	boshpkg "github.com/cloudfoundry/bosh-cli/release/pkg"
	. "github.com/cloudfoundry/bosh-cli/release/resource"
)

type ManifestReader struct {
	fs boshsys.FileSystem

	logTag string
	logger boshlog.Logger
}

func NewManifestReader(fs boshsys.FileSystem, logger boshlog.Logger) ManifestReader {
	return ManifestReader{
		fs:     fs,
		logTag: "release.ManifestReader",
		logger: logger,
	}
}

func (r ManifestReader) Read(path string) (Release, error) {
	manifest, err := boshman.NewManifestFromPath(path, r.fs)
	if err != nil {
		return nil, err
	}

	release, err := r.newRelease(manifest)
	if err != nil {
		return nil, bosherr.WrapError(err, "Constructing release from manifest")
	}

	return release, nil
}

func (r ManifestReader) newRelease(manifest boshman.Manifest) (Release, error) {
	var errs []error

	jobs, err := r.newJobs(manifest.Jobs)
	if err != nil {
		errs = append(errs, bosherr.WrapError(err, "Constructing jobs from manifest"))
	}

	packages, err := r.newPackages(manifest.Packages)
	if err != nil {
		errs = append(errs, bosherr.WrapError(err, "Constructing packages from manifest"))
	}

	compiledPkgs, err := r.newCompiledPackages(manifest.CompiledPkgs)
	if err != nil {
		errs = append(errs, bosherr.WrapError(err, "Constructing compiled packages from manifest"))
	}

	license, err := r.newLicense(manifest.License)
	if err != nil {
		errs = append(errs, bosherr.WrapError(err, "Constructing license from manifest"))
	}

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

		fs: r.fs,
	}

	return release, nil
}

func (r ManifestReader) newJobs(refs []boshman.JobRef) ([]*boshjob.Job, error) {
	var jobs []*boshjob.Job
	var errs []error

	for _, ref := range refs {
		resource := NewExistingResource(ref.Name, ref.Fingerprint, ref.SHA1)

		job := boshjob.NewJob(resource)

		jobs = append(jobs, job)
	}

	if len(errs) > 0 {
		return nil, bosherr.NewMultiError(errs...)
	}

	return jobs, nil
}

func (r ManifestReader) newPackages(refs []boshman.PackageRef) ([]*boshpkg.Package, error) {
	var packages []*boshpkg.Package
	var errs []error

	for _, ref := range refs {
		resource := NewExistingResource(ref.Name, ref.Fingerprint, ref.SHA1)

		pkg := boshpkg.NewPackage(resource, ref.Dependencies)

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

func (r ManifestReader) newCompiledPackages(refs []boshman.CompiledPackageRef) ([]*boshpkg.CompiledPackage, error) {
	var compiledPkgs []*boshpkg.CompiledPackage
	var errs []error

	for _, ref := range refs {
		compiledPkg := boshpkg.NewCompiledPackageWithoutArchive(
			ref.Name, ref.Fingerprint, ref.OSVersionSlug, ref.SHA1, ref.Dependencies)

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

func (r ManifestReader) newLicense(ref *boshman.LicenseRef) (*boshlic.License, error) {
	if ref != nil {
		resource := NewExistingResource("license", ref.Fingerprint, ref.SHA1)

		license := boshlic.NewLicense(resource)

		return license, nil
	}

	return nil, nil
}
