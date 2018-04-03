package release

import (
	boshjob "github.com/cloudfoundry/bosh-cli/release/job"
	boshlic "github.com/cloudfoundry/bosh-cli/release/license"
	boshman "github.com/cloudfoundry/bosh-cli/release/manifest"
	boshpkg "github.com/cloudfoundry/bosh-cli/release/pkg"
	boshres "github.com/cloudfoundry/bosh-cli/release/resource"
)

type Extractor interface {
	Extract(string) (Release, error)
}

//go:generate counterfeiter . Reader

type Reader interface {
	// Read reads an archive for example and returns a Release.
	Read(string) (Release, error)
}

//go:generate counterfeiter . Writer

type Writer interface {
	// Write writes an archive for example and returns its path.
	// Archive does not include packages that have fingerprints
	// included in the second argument.
	Write(Release, []string) (string, error)
}

//go:generate counterfeiter . Release

type Release interface {
	Name() string
	SetName(string)

	Version() string
	SetVersion(string)

	CommitHashWithMark(string) string
	SetCommitHash(string)
	SetUncommittedChanges(bool)

	Jobs() []*boshjob.Job
	Packages() []*boshpkg.Package
	CompiledPackages() []*boshpkg.CompiledPackage
	License() *boshlic.License

	IsCompiled() bool

	FindJobByName(string) (boshjob.Job, bool)
	Manifest() boshman.Manifest

	Build(dev, final ArchiveIndicies, parallel int) error
	Finalize(final ArchiveIndicies, parallel int) error

	CopyWith(jobs []*boshjob.Job,
		packages []*boshpkg.Package,
		lic *boshlic.License,
		compiledPackages []*boshpkg.CompiledPackage) Release

	CleanUp() error
}

type ArchiveIndicies struct {
	Jobs     boshres.ArchiveIndex
	Packages boshres.ArchiveIndex
	Licenses boshres.ArchiveIndex
}

type Manager interface {
	Add(Release)
	List() []Release
	Find(string) (Release, bool)
	DeleteAll() error
}
