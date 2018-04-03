package release

import (
	"os"
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshcmd "github.com/cloudfoundry/bosh-utils/fileutil"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"gopkg.in/yaml.v2"

	boshjob "github.com/cloudfoundry/bosh-cli/release/job"
	boshlic "github.com/cloudfoundry/bosh-cli/release/license"
	boshpkg "github.com/cloudfoundry/bosh-cli/release/pkg"
)

type ArchiveWriter struct {
	compressor     boshcmd.Compressor
	fs             boshsys.FileSystem
	filesToInclude []string

	logTag string
	logger boshlog.Logger
}

func NewArchiveWriter(compressor boshcmd.Compressor, fs boshsys.FileSystem, logger boshlog.Logger) ArchiveWriter {
	return ArchiveWriter{compressor: compressor, fs: fs, logTag: "release.ArchiveWriter", logger: logger}
}

func (w ArchiveWriter) Write(release Release, pkgFpsToSkip []string) (string, error) {
	stagingDir, err := w.fs.TempDir("bosh-release")
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Creating staging release dir")
	}

	defer w.cleanUp(stagingDir)

	w.logger.Info(w.logTag, "Writing release tarball into '%s'", stagingDir)

	manifestBytes, err := yaml.Marshal(release.Manifest())
	if err != nil {
		return "", bosherr.WrapError(err, "Marshalling release manifest")
	}

	manifestPath := filepath.Join(stagingDir, "release.MF")

	err = w.fs.WriteFile(manifestPath, manifestBytes)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Writing release manifest '%s'", manifestPath)
	}

	w.filesToInclude = w.appendFile("release.MF")

	jobsFiles, err := w.writeJobs(release.Jobs(), stagingDir)
	if err != nil {
		return "", bosherr.WrapError(err, "Writing jobs")
	}

	w.filesToInclude = w.appendFiles(jobsFiles)

	packagesFiles, err := w.writePackages(release.Packages(), pkgFpsToSkip, stagingDir)
	if err != nil {
		return "", bosherr.WrapError(err, "Writing packages")
	}

	w.filesToInclude = w.appendFiles(packagesFiles)

	compiledPackagesFiles, err := w.writeCompiledPackages(release.CompiledPackages(), pkgFpsToSkip, stagingDir)
	if err != nil {
		return "", bosherr.WrapError(err, "Writing compiled packages")
	}

	w.filesToInclude = w.appendFiles(compiledPackagesFiles)

	licenseFiles, err := w.writeLicense(release.License(), stagingDir)
	if err != nil {
		return "", bosherr.WrapError(err, "Writing license")
	}

	w.filesToInclude = w.appendFiles(licenseFiles)

	files := w.filesToInclude
	path, err := w.compressor.CompressSpecificFilesInDir(stagingDir, files)

	if err != nil {
		return "", bosherr.WrapError(err, "Compressing release")
	}

	return path, nil
}

func (w ArchiveWriter) cleanUp(stagingDir string) {
	removeErr := w.fs.RemoveAll(stagingDir)
	if removeErr != nil {
		w.logger.Error(w.logTag, "Failed to remove staging dir for release: %s", removeErr.Error())
	}
}

func (w ArchiveWriter) writeJobs(jobs []*boshjob.Job, stagingDir string) ([]string, error) {
	var jobsFiles []string

	if len(jobs) == 0 {
		return jobsFiles, nil
	}

	jobsPath := filepath.Join(stagingDir, "jobs")

	err := w.fs.MkdirAll(jobsPath, os.ModePerm)
	if err != nil {
		return jobsFiles, bosherr.WrapError(err, "Creating jobs/")
	}

	jobsFiles = append(jobsFiles, "jobs")

	for _, job := range jobs {
		err := w.fs.CopyFile(job.ArchivePath(), filepath.Join(jobsPath, job.Name()+".tgz"))
		if err != nil {
			return jobsFiles, bosherr.WrapErrorf(err, "Copying job '%s' archive into staging dir", job.Name())
		}
	}

	return jobsFiles, nil
}

func (w ArchiveWriter) writePackages(packages []*boshpkg.Package, pkgFpsToSkip []string, stagingDir string) ([]string, error) {
	var packagesFiles []string

	if len(packages) == 0 {
		return packagesFiles, nil
	}

	pkgsPath := filepath.Join(stagingDir, "packages")

	err := w.fs.MkdirAll(pkgsPath, os.ModePerm)
	if err != nil {
		return packagesFiles, bosherr.WrapError(err, "Creating packages/")
	}

	packagesFiles = append(packagesFiles, "packages")

	for _, pkg := range packages {
		if w.shouldSkip(pkg.Fingerprint(), pkgFpsToSkip) {
			w.logger.Debug(w.logTag, "Package '%s' was filtered out", pkg.Name())
		} else {
			err := w.fs.CopyFile(pkg.ArchivePath(), filepath.Join(pkgsPath, pkg.Name()+".tgz"))
			if err != nil {
				return packagesFiles, bosherr.WrapErrorf(err, "Copying package '%s' archive into staging dir", pkg.Name())
			}
		}
	}

	return packagesFiles, nil
}

func (w ArchiveWriter) writeCompiledPackages(compiledPkgs []*boshpkg.CompiledPackage, pkgFpsToSkip []string, stagingDir string) ([]string, error) {
	var compiledPackagesFiles []string

	if len(compiledPkgs) == 0 {
		return compiledPackagesFiles, nil
	}

	pkgsPath := filepath.Join(stagingDir, "compiled_packages")

	err := w.fs.MkdirAll(pkgsPath, os.ModePerm)
	if err != nil {
		return compiledPackagesFiles, bosherr.WrapError(err, "Creating compiled_packages/")
	}

	compiledPackagesFiles = append(compiledPackagesFiles, "compiled_packages")

	for _, compiledPkg := range compiledPkgs {
		if w.shouldSkip(compiledPkg.Fingerprint(), pkgFpsToSkip) {
			w.logger.Debug(w.logTag, "Compiled package '%s' was filtered out", compiledPkg.Name())
		} else {
			err := w.fs.CopyFile(compiledPkg.ArchivePath(), filepath.Join(pkgsPath, compiledPkg.Name()+".tgz"))
			if err != nil {
				return compiledPackagesFiles, bosherr.WrapErrorf(err, "Copying compiled package '%s' archive into staging dir", compiledPkg.Name())
			}
		}
	}

	return compiledPackagesFiles, nil
}

func (w ArchiveWriter) writeLicense(license *boshlic.License, stagingDir string) ([]string, error) {
	var licenseFiles []string

	if license == nil {
		return licenseFiles, nil
	}

	err := w.fs.CopyFile(license.ArchivePath(), filepath.Join(stagingDir, "license.tgz"))
	if err != nil {
		return licenseFiles, bosherr.WrapError(err, "Copying license archive into staging dir")
	}

	licenseFiles = append(licenseFiles, "license.tgz")

	err = w.compressor.DecompressFileToDir(license.ArchivePath(), stagingDir, boshcmd.CompressorOptions{})
	if err != nil {
		return licenseFiles, bosherr.WrapErrorf(err, "Decompressing license archive into staging dir")
	}

	licenseFiles, err = w.appendMatchedFiles(licenseFiles, stagingDir, "LICENSE*")
	if err != nil {
		return licenseFiles, bosherr.WrapErrorf(err, "Reading LICENSE files")
	}

	licenseFiles, err = w.appendMatchedFiles(licenseFiles, stagingDir, "NOTICE*")
	if err != nil {
		return licenseFiles, bosherr.WrapErrorf(err, "Reading NOTICE files")
	}

	return licenseFiles, nil
}

func (w ArchiveWriter) shouldSkip(fp string, pkgFpsToSkip []string) bool {
	for _, pkgFp := range pkgFpsToSkip {
		if fp == pkgFp {
			return true
		}
	}
	return false
}

func (w ArchiveWriter) appendFile(filename string) []string {
	return append(w.filesToInclude, filename)
}

func (w ArchiveWriter) appendFiles(filenames []string) []string {
	for _, filename := range filenames {
		w.filesToInclude = w.appendFile(filename)
	}

	return w.filesToInclude
}

func (w ArchiveWriter) appendMatchedFiles(files []string, stagingDir string, filePattern string) ([]string, error) {
	noticeMatches, err := w.fs.Glob(filepath.Join(stagingDir, filePattern))
	if err != nil {
		return files, err
	}
	for _, file := range noticeMatches {
		files = append(files, filepath.Base(file))
	}

	return files, nil
}
