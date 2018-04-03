package resource

import (
	"github.com/cloudfoundry/bosh-cli/crypto"
	crypto2 "github.com/cloudfoundry/bosh-utils/crypto"
)

//go:generate counterfeiter . Archive

type Archive interface {
	Fingerprint() (string, error)
	Build(expectedFp string) (string, string, error)
}

type ArchiveFunc func(args ArchiveFactoryArgs) Archive

type ArchiveFactoryArgs struct {
	Files          []File
	PrepFiles      []File
	Chunks         []string
	FollowSymlinks bool
}

//go:generate counterfeiter . ArchiveIndex

type ArchiveIndex interface {
	Find(name, fingerprint string) (string, string, error)
	Add(name, fingerprint, path, sha1 string) (string, string, error)
}

//go:generate counterfeiter . Resource

type Resource interface {
	Name() string
	Fingerprint() string

	ArchivePath() string
	ArchiveDigest() string

	Build(dev, final ArchiveIndex) error
	Finalize(final ArchiveIndex) error

	RehashWithCalculator(calculator crypto.DigestCalculator, archiveFilePathReader crypto2.ArchiveDigestFilePathReader) (Resource, error)
}

//go:generate counterfeiter . Fingerprinter

type Fingerprinter interface {
	Calculate([]File, []string) (string, error)
}
