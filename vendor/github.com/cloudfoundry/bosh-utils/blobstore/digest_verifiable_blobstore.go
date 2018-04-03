package blobstore

import (
	"os"

	boshcrypto "github.com/cloudfoundry/bosh-utils/crypto"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type digestVerifiableBlobstore struct {
	blobstore        Blobstore
	fs               boshsys.FileSystem
	createAlgorithms []boshcrypto.Algorithm
}

func NewDigestVerifiableBlobstore(blobstore Blobstore, fs boshsys.FileSystem, createAlgorithms []boshcrypto.Algorithm) DigestBlobstore {
	return digestVerifiableBlobstore{
		blobstore:        blobstore,
		fs:               fs,
		createAlgorithms: createAlgorithms,
	}
}

func (b digestVerifiableBlobstore) Get(blobID string, digest boshcrypto.Digest) (string, error) {
	fileName, err := b.blobstore.Get(blobID)
	if err != nil {
		return "", bosherr.WrapError(err, "Getting blob from inner blobstore")
	}

	file, err := b.fs.OpenFile(fileName, os.O_RDONLY, 0)
	if err != nil {
		return "", err
	}

	defer file.Close()

	err = digest.Verify(file)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Checking downloaded blob '%s'", blobID)
	}

	return fileName, nil
}

func (b digestVerifiableBlobstore) Delete(blobId string) error {
	return b.blobstore.Delete(blobId)
}

func (b digestVerifiableBlobstore) CleanUp(fileName string) error {
	return b.blobstore.CleanUp(fileName)
}

func (b digestVerifiableBlobstore) Create(fileName string) (string, boshcrypto.MultipleDigest, error) {
	multipleDigest, err := b.createDigest(fileName)
	if err != nil {
		return "", boshcrypto.MultipleDigest{}, err
	}

	blobID, err := b.blobstore.Create(fileName)
	return blobID, multipleDigest, err
}

func (b digestVerifiableBlobstore) Validate() error {
	return b.blobstore.Validate()
}

func (b digestVerifiableBlobstore) createDigest(fileName string) (boshcrypto.MultipleDigest, error) {
	digests := []boshcrypto.Digest{}
	for _, algo := range b.createAlgorithms {
		digest, err := b.computeDigest(algo, fileName)
		if err != nil {
			return boshcrypto.MultipleDigest{}, err
		}
		digests = append(digests, digest)
	}
	return boshcrypto.MustNewMultipleDigest(digests...), nil
}

func (b digestVerifiableBlobstore) computeDigest(algo boshcrypto.Algorithm, fileName string) (boshcrypto.Digest, error) {
	file, err := b.fs.OpenFile(fileName, os.O_RDONLY, 0)
	if err != nil {
		return boshcrypto.MultipleDigest{}, err
	}

	defer file.Close()

	return algo.CreateDigest(file)
}
