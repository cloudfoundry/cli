package blobstore

import (
	boshcrypto "github.com/cloudfoundry/bosh-utils/crypto"
)

type DigestBlobstore interface {
	// Assuming that local file system is available,
	// file handle is returned to downloaded blob.
	// Caller must not assume anything about layout of such scratch space.
	// Cleanup call is needed to properly cleanup downloaded blob.
	Get(blobID string, digest boshcrypto.Digest) (fileName string, err error)

	CleanUp(fileName string) (err error)

	Create(fileName string) (blobID string, digest boshcrypto.MultipleDigest, err error)

	Validate() (err error)

	Delete(blobId string) (err error)
}
