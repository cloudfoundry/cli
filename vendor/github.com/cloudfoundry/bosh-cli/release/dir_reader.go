package release

import (
	"path/filepath"
	"sort"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	boshjob "github.com/cloudfoundry/bosh-cli/release/job"
	boshlic "github.com/cloudfoundry/bosh-cli/release/license"
	boshpkg "github.com/cloudfoundry/bosh-cli/release/pkg"
)

type DirReader struct {
	jobDirReader boshjob.DirReader
	pkgDirReader boshpkg.DirReader
	licDirReader boshlic.DirReader
	fs           boshsys.FileSystem

	logTag string
	logger boshlog.Logger
}

func NewDirReader(
	jobDirReader boshjob.DirReader,
	pkgDirReader boshpkg.DirReader,
	licDirReader boshlic.DirReader,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) DirReader {
	return DirReader{
		jobDirReader: jobDirReader,
		pkgDirReader: pkgDirReader,
		licDirReader: licDirReader,
		fs:           fs,

		logTag: "release.DirReader",
		logger: logger,
	}
}

func (r DirReader) Read(path string) (Release, error) {
	var packages []*boshpkg.Package
	var jobs []*boshjob.Job

	var errs []error
	var err error

	pkgMatches, err := r.fs.Glob(filepath.Join(path, "packages", "*"))
	if err != nil {
		errs = append(errs, bosherr.WrapError(err, "Listing packages in directory"))
	} else {
		packages, err = r.newPackages(pkgMatches)
		if err != nil {
			errs = append(errs, bosherr.WrapError(err, "Constructing packages from directory"))
		}
	}

	jobMatches, err := r.fs.Glob(filepath.Join(path, "jobs", "*"))
	if err != nil {
		errs = append(errs, bosherr.WrapError(err, "Listing jobs in directory"))
	} else {
		jobs, err = r.newJobs(packages, jobMatches)
		if err != nil {
			errs = append(errs, bosherr.WrapError(err, "Constructing jobs from manifest"))
		}
	}

	license, err := r.newLicense(path)
	if err != nil {
		errs = append(errs, bosherr.WrapError(err, "Constructing license from manifest"))
	}

	if len(errs) > 0 {
		return nil, bosherr.NewMultiError(errs...)
	}

	release := &release{
		jobs:     jobs,
		packages: packages,
		license:  license,
		// no compiled packages
		// no clean up
	}

	return release, nil
}

func (r DirReader) newJobs(packages []*boshpkg.Package, jobMatches []string) ([]*boshjob.Job, error) {
	var jobs []*boshjob.Job
	var errs []error

	for _, jobMatch := range jobMatches {

		info, err := r.fs.Stat(jobMatch)
		if err != nil {
			errs = append(errs, bosherr.WrapErrorf(err, "Reading job from '%s'", jobMatch))
			continue
		}

		if info.IsDir() {
			job, err := r.jobDirReader.Read(jobMatch)
			if err != nil {
				errs = append(errs, bosherr.WrapErrorf(err, "Reading job from '%s'", jobMatch))
				continue
			}

			err = job.AttachPackages(packages)
			if err != nil {
				errs = append(errs, err)
				continue
			}

			jobs = append(jobs, job)
		}
	}

	if len(errs) > 0 {
		return nil, bosherr.NewMultiError(errs...)
	}

	sort.Sort(boshjob.ByName(jobs))

	return jobs, nil
}

func (r DirReader) newPackages(pkgMatches []string) ([]*boshpkg.Package, error) {
	var packages []*boshpkg.Package
	var errs []error

	for _, pkgMatch := range pkgMatches {
		info, err := r.fs.Stat(pkgMatch)
		if err != nil {
			errs = append(errs, bosherr.WrapErrorf(err, "Reading package from '%s'", pkgMatch))
			continue
		}

		if info.IsDir() {
			pkg, err := r.pkgDirReader.Read(pkgMatch)
			if err != nil {
				errs = append(errs, bosherr.WrapErrorf(err, "Reading package from '%s'", pkgMatch))
				continue
			}

			packages = append(packages, pkg)
		}
	}

	for _, pkg := range packages {
		err := pkg.AttachDependencies(packages)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return []*boshpkg.Package{}, bosherr.NewMultiError(errs...)
	}

	sort.Sort(boshpkg.ByName(packages))

	return packages, nil
}

func (r DirReader) newLicense(path string) (*boshlic.License, error) {
	lic, err := r.licDirReader.Read(path)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Reading license from '%s'", path)
	}

	return lic, nil
}
