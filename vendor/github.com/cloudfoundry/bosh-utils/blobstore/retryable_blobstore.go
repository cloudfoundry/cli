package blobstore

import (
	boshcrypto "github.com/cloudfoundry/bosh-utils/crypto"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type retryableBlobstore struct {
	blobstore DigestBlobstore
	maxTries  int

	logTag string
	logger boshlog.Logger
}

func NewRetryableBlobstore(blobstore DigestBlobstore, maxTries int, logger boshlog.Logger) DigestBlobstore {
	return retryableBlobstore{
		blobstore: blobstore,
		maxTries:  maxTries,
		logTag:    "retryableBlobstore",
		logger:    logger,
	}
}

func (b retryableBlobstore) Get(blobID string, fingerprint boshcrypto.Digest) (string, error) {
	var fileName string
	var lastErr error

	for i := 1; i <= b.maxTries; i++ {
		fileName, lastErr = b.blobstore.Get(blobID, fingerprint)
		if lastErr == nil {
			return fileName, nil
		}

		b.logger.Info(b.logTag,
			"Failed to get blob with error '%s', attempt %d out of %d", lastErr.Error(), i, b.maxTries)
	}

	return "", bosherr.WrapError(lastErr, "Getting blob from inner blobstore")
}

func (b retryableBlobstore) CleanUp(fileName string) error {
	return b.blobstore.CleanUp(fileName)
}

func (b retryableBlobstore) Delete(blobID string) error {
	return b.blobstore.Delete(blobID)
}

func (b retryableBlobstore) Create(fileName string) (string, boshcrypto.MultipleDigest, error) {
	var lastErr error

	for i := 1; i <= b.maxTries; i++ {
		blobID, digest, thisErr := b.blobstore.Create(fileName)
		if thisErr == nil {
			return blobID, digest, nil
		}

		lastErr = thisErr
		b.logger.Info(b.logTag,
			"Failed to create blob with error %s, attempt %d out of %d", lastErr.Error(), i, b.maxTries)
	}

	return "", boshcrypto.MultipleDigest{}, bosherr.WrapError(lastErr, "Creating blob in inner blobstore")
}

func (b retryableBlobstore) Validate() error {
	if b.maxTries < 1 {
		return bosherr.Error("Max tries must be > 0")
	}

	return b.blobstore.Validate()
}
