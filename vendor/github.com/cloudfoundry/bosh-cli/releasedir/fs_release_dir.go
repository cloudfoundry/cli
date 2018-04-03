package releasedir

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"code.cloudfoundry.org/clock"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	semver "github.com/cppforlife/go-semi-semantic/version"

	boshrel "github.com/cloudfoundry/bosh-cli/release"
	boshpkg "github.com/cloudfoundry/bosh-cli/release/pkg"
	boshpkgman "github.com/cloudfoundry/bosh-cli/release/pkg/manifest"
)

var (
	DefaultFinalVersion   = semver.MustNewVersionFromString("1")
	DefaultDevVersion     = semver.MustNewVersionFromString("0+dev.0")
	DefaultDevPostRelease = semver.MustNewVersionSegmentFromString("dev.1")
)

type FSReleaseDir struct {
	dirPath string

	config    Config
	gitRepo   GitRepo
	blobsDir  BlobsDir
	generator Generator

	devReleases   ReleaseIndex
	finalReleases ReleaseIndex
	finalIndicies boshrel.ArchiveIndicies

	releaseReader        boshrel.Reader
	releaseArchiveWriter boshrel.Writer

	timeService clock.Clock
	fs          boshsys.FileSystem

	parallel int
}

func NewFSReleaseDir(
	dirPath string,
	config Config,
	gitRepo GitRepo,
	blobsDir BlobsDir,
	generator Generator,
	devReleases ReleaseIndex,
	finalReleases ReleaseIndex,
	finalIndicies boshrel.ArchiveIndicies,
	releaseReader boshrel.Reader,
	timeService clock.Clock,
	fs boshsys.FileSystem,
	parallel int,
) FSReleaseDir {
	return FSReleaseDir{
		dirPath: dirPath,

		config:    config,
		gitRepo:   gitRepo,
		blobsDir:  blobsDir,
		generator: generator,

		devReleases:   devReleases,
		finalReleases: finalReleases,
		finalIndicies: finalIndicies,

		releaseReader: releaseReader,

		timeService: timeService,
		fs:          fs,

		parallel: parallel,
	}
}

func (d FSReleaseDir) Init(git bool) error {
	for _, name := range []string{"jobs", "packages", "src"} {
		err := d.fs.MkdirAll(filepath.Join(d.dirPath, name), os.ModePerm)
		if err != nil {
			return bosherr.WrapErrorf(err, "Creating %s/", name)
		}
	}

	name := strings.TrimSuffix(filepath.Base(d.dirPath), "-release")

	err := d.config.SaveName(name)
	if err != nil {
		return err
	}

	err = d.blobsDir.Init()
	if err != nil {
		return bosherr.WrapErrorf(err, "Initing blobs")
	}

	if git {
		err = d.gitRepo.Init()
		if err != nil {
			return err
		}
	}

	return nil
}

func (d FSReleaseDir) GenerateJob(name string) error {
	return d.generator.GenerateJob(name)
}

func (d FSReleaseDir) GeneratePackage(name string) error {
	return d.generator.GeneratePackage(name)
}

func (d FSReleaseDir) Reset() error {
	for _, name := range []string{".dev_builds", "dev_releases", ".blobs", "blobs"} {
		err := d.fs.RemoveAll(filepath.Join(d.dirPath, name))
		if err != nil {
			return bosherr.WrapErrorf(err, "Removing %s/", name)
		}
	}

	return nil
}

func (d FSReleaseDir) DefaultName() (string, error) {
	return d.config.Name()
}

func (d FSReleaseDir) NextFinalVersion(name string) (semver.Version, error) {
	lastVer, err := d.finalReleases.LastVersion(name)
	if err != nil {
		return semver.Version{}, err
	} else if lastVer == nil {
		return DefaultFinalVersion, nil
	}

	incVer, err := lastVer.IncrementRelease()
	if err != nil {
		return semver.Version{}, bosherr.WrapErrorf(err, "Incrementing last final version")
	}

	return incVer, nil
}

func (d FSReleaseDir) NextDevVersion(name string, timestamp bool) (semver.Version, error) {
	lastVer, _, err := d.lastDevOrFinalVersion(name)
	if err != nil {
		return semver.Version{}, err
	} else if lastVer == nil {
		lastVer = &DefaultDevVersion
	}

	incVer, err := lastVer.IncrementPostRelease(DefaultDevPostRelease)
	if err != nil {
		return semver.Version{}, bosherr.WrapErrorf(err, "Incrementing last dev version")
	}

	if timestamp {
		ts := d.timeService.Now().Unix()

		postRelease, err := semver.NewVersionSegmentFromString(fmt.Sprintf("dev.%d", ts))
		if err != nil {
			panic(fmt.Sprintf("Failed to build post release version segment from timestamp (%d): %s", ts, err))
		}

		incVer, err = semver.NewVersion(incVer.Release.Copy(), incVer.PreRelease.Copy(), postRelease)
		if err != nil {
			panic(fmt.Sprintf("Failed to build version: %s", err))
		}
	}

	return incVer, nil
}

func (d FSReleaseDir) FindRelease(name string, version semver.Version) (boshrel.Release, error) {
	if len(name) == 0 {
		defaultName, err := d.DefaultName()
		if err != nil {
			return nil, err
		}
		name = defaultName
	}

	relIndex := d.finalReleases

	if version.Empty() {
		lastVer, lastRelIndex, err := d.lastDevOrFinalVersion(name)
		if err != nil {
			return nil, err
		} else if lastVer == nil {
			return nil, bosherr.Errorf("Expected to find at least one dev or final version")
		}
		version = *lastVer
		relIndex = lastRelIndex
	}

	return d.releaseReader.Read(relIndex.ManifestPath(name, version.AsString()))
}

func (d FSReleaseDir) BuildRelease(name string, version semver.Version, force bool) (boshrel.Release, error) {
	dirty, err := d.gitRepo.MustNotBeDirty(force)
	if err != nil {
		return nil, err
	}

	commitSHA, err := d.gitRepo.LastCommitSHA()
	if err != nil {
		return nil, err
	}

	err = d.blobsDir.SyncBlobs(1)
	if err != nil {
		return nil, err
	}

	release, err := d.releaseReader.Read(d.dirPath)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Building a release from directory '%s'", d.dirPath)
	}

	release.SetName(name)
	release.SetVersion(version.AsString())
	release.SetCommitHash(commitSHA)
	release.SetUncommittedChanges(dirty)

	err = d.devReleases.Add(release.Manifest())
	if err != nil {
		return nil, err
	}

	return release, nil
}

func (d FSReleaseDir) VendorPackage(pkg *boshpkg.Package) error {
	allInterestingPkgs := map[*boshpkg.Package]struct{}{}

	d.collectDependentPackages(pkg, allInterestingPkgs)

	for pkg2, _ := range allInterestingPkgs {
		err := pkg2.Finalize(d.finalIndicies.Packages)
		if err != nil {
			return bosherr.WrapErrorf(err, "Finalizing vendored package")
		}

		err = d.writeVendoredPackage(pkg2)
		if err != nil {
			return bosherr.WrapErrorf(err, "Writing vendored package")
		}
	}

	return nil
}

func (d FSReleaseDir) collectDependentPackages(pkg *boshpkg.Package, allInterestingPkgs map[*boshpkg.Package]struct{}) {
	allInterestingPkgs[pkg] = struct{}{}
	for _, pkg2 := range pkg.Dependencies {
		d.collectDependentPackages(pkg2, allInterestingPkgs)
	}
}

func (d FSReleaseDir) writeVendoredPackage(pkg *boshpkg.Package) error {
	name := pkg.Name()
	pkgDirPath := filepath.Join(d.dirPath, "packages", name)

	err := d.fs.RemoveAll(pkgDirPath)
	if err != nil {
		return bosherr.WrapErrorf(err, "Removing package '%s' dir", name)
	}

	err = d.fs.MkdirAll(pkgDirPath, os.ModePerm)
	if err != nil {
		return bosherr.WrapErrorf(err, "Creating package '%s' dir", name)
	}

	manifestLock := boshpkgman.ManifestLock{Name: name, Fingerprint: pkg.Fingerprint()}

	for _, pkg2 := range pkg.Dependencies {
		manifestLock.Dependencies = append(manifestLock.Dependencies, pkg2.Name())
	}

	manifestLockBytes, err := manifestLock.AsBytes()
	if err != nil {
		return bosherr.WrapErrorf(err, "Marshaling vendored package '%s' spec lock", name)
	}

	err = d.fs.WriteFile(filepath.Join(pkgDirPath, "spec.lock"), manifestLockBytes)
	if err != nil {
		return bosherr.WrapErrorf(err, "Creating package '%s' spec lock file", name)
	}

	return nil
}

func (d FSReleaseDir) FinalizeRelease(release boshrel.Release, force bool) error {
	_, err := d.gitRepo.MustNotBeDirty(force)
	if err != nil {
		return err
	}

	found, err := d.finalReleases.Contains(release)
	if err != nil {
		return err
	} else if found {
		return bosherr.Errorf("Release '%s' version '%s' already exists", release.Name(), release.Version())
	}

	err = release.Finalize(d.finalIndicies, d.parallel)
	if err != nil {
		return err
	}

	return d.finalReleases.Add(release.Manifest())
}

func (d FSReleaseDir) lastDevOrFinalVersion(name string) (*semver.Version, ReleaseIndex, error) {
	lastDevVer, err := d.devReleases.LastVersion(name)
	if err != nil {
		return nil, nil, err
	}

	lastFinalVer, err := d.finalReleases.LastVersion(name)
	if err != nil {
		return nil, nil, err
	}

	switch {
	case lastDevVer != nil && lastFinalVer != nil:
		if lastFinalVer.IsGt(*lastDevVer) {
			return lastFinalVer, d.finalReleases, nil
		} else {
			return lastDevVer, d.devReleases, nil
		}
	case lastDevVer != nil && lastFinalVer == nil:
		return lastDevVer, d.devReleases, nil
	case lastDevVer == nil && lastFinalVer != nil:
		return lastFinalVer, d.finalReleases, nil
	default:
		return nil, nil, nil
	}
}
