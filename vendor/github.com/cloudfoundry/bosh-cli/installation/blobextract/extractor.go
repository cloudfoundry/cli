package blobextract

import (
	"os"

	boshblob "github.com/cloudfoundry/bosh-utils/blobstore"
	boshcrypto "github.com/cloudfoundry/bosh-utils/crypto"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshcmd "github.com/cloudfoundry/bosh-utils/fileutil"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

//go:generate counterfeiter -o fakeblobextract/fake_extractor.go extractor.go Extractor
type Extractor interface {
	Extract(blobID, blobSHA1, targetDir string) error
	Cleanup(blobID string, extractedBlobPath string) error
	ChmodExecutables(binPath string) error
}

type extractor struct {
	fs         boshsys.FileSystem
	compressor boshcmd.Compressor
	blobstore  boshblob.DigestBlobstore
	logger     boshlog.Logger
	logTag     string
}

func NewExtractor(
	fs boshsys.FileSystem,
	compressor boshcmd.Compressor,
	blobstore boshblob.DigestBlobstore,
	logger boshlog.Logger,
) Extractor {
	return &extractor{
		fs:         fs,
		compressor: compressor,
		blobstore:  blobstore,
		logger:     logger,
		logTag:     "blobExtractor",
	}
}

func (e *extractor) Extract(blobID string, digestString string, targetDir string) error {
	// Retrieve a temp copy of blob

	digest, err := boshcrypto.ParseMultipleDigest(digestString)
	if err != nil {
		return bosherr.WrapErrorf(err, "Parsing multiple digest string: %s", digestString)
	}

	filePath, err := e.blobstore.Get(blobID, digest)
	if err != nil {
		return bosherr.WrapErrorf(err, "Getting object from blobstore: %s", blobID)
	}
	// Clean up temp copy of blob
	defer e.cleanUpBlob(filePath)

	existed := e.fs.FileExists(targetDir)
	if !existed {
		err = e.fs.MkdirAll(targetDir, os.ModePerm)
		if err != nil {
			return bosherr.WrapErrorf(err, "Creating target dir: %s", targetDir)
		}
	}

	err = e.compressor.DecompressFileToDir(filePath, targetDir, boshcmd.CompressorOptions{})
	if err != nil {
		if !existed {
			// Clean up extracted contents of blob
			e.cleanUpFile(targetDir)
		}
		return bosherr.WrapErrorf(err, "Decompressing compiled package: BlobID: '%s', BlobSHA1: '%s'", blobID, digestString)
	}
	return nil
}

func (e *extractor) ChmodExecutables(binGlob string) error {
	files, err := e.fs.Glob(binGlob)
	if err != nil {
		return bosherr.WrapErrorf(err, "Globbing %s", binGlob)
	}

	for _, file := range files {
		err = e.fs.Chmod(file, os.FileMode(0755))
		if err != nil {
			return bosherr.WrapErrorf(err, "Making '%s' executable in '%s'", file, binGlob)
		}
	}
	return nil
}

func (e *extractor) Cleanup(blobID string, decompressedBlobPath string) error {
	e.cleanUpFile(decompressedBlobPath)

	err := e.blobstore.Delete(blobID)
	if err != nil {
		return bosherr.WrapErrorf(err, "Deleting object from blobstore: %s", blobID)
	}
	return nil
}

func (e *extractor) cleanUpBlob(filePath string) {
	err := e.blobstore.CleanUp(filePath)
	if err != nil {
		e.logger.Error(
			e.logTag,
			bosherr.WrapErrorf(err, "Removing compiled package tarball: %s", filePath).Error(),
		)
	}
}

func (e *extractor) cleanUpFile(filePath string) {
	err := e.fs.RemoveAll(filePath)
	if err != nil {
		e.logger.Error(
			e.logTag,
			bosherr.WrapErrorf(err, "Removing: %s", filePath).Error(),
		)
	}
}
