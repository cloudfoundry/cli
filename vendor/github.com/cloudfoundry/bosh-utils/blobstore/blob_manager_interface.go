package blobstore

import (
	boshcrypto "github.com/cloudfoundry/bosh-utils/crypto"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"io"
)

type BlobManagerInterface interface {
	Fetch(blobID string) (boshsys.File, error, int)

	Write(blobID string, reader io.Reader) error

	GetPath(blobID string, digest boshcrypto.Digest) (string, error)

	Delete(blobID string) error

	BlobExists(blobID string) bool
}
