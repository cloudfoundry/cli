package pkg

import (
	boshman "github.com/cloudfoundry/bosh-cli/release/manifest"
)

//go:generate counterfeiter . Compilable

type Compilable interface {
	Name() string
	Fingerprint() string

	ArchivePath() string
	ArchiveDigest() string

	IsCompiled() bool

	Deps() []Compilable
}

//go:generate counterfeiter . ArchiveReader

type ArchiveReader interface {
	Read(boshman.PackageRef, string) (*Package, error)
}

//go:generate counterfeiter . DirReader

type DirReader interface {
	Read(string) (*Package, error)
}
