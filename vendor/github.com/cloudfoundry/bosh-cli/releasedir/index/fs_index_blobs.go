package index

import (
	"fmt"
	"os"
	"path/filepath"

	boshblob "github.com/cloudfoundry/bosh-utils/blobstore"
	boshcrypto "github.com/cloudfoundry/bosh-utils/crypto"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshfu "github.com/cloudfoundry/bosh-utils/fileutil"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type FSIndexBlobs struct {
	dirPath  string
	reporter Reporter

	blobstore boshblob.DigestBlobstore
	fs        boshsys.FileSystem
}

func NewFSIndexBlobs(
	dirPath string,
	reporter Reporter,
	blobstore boshblob.DigestBlobstore,
	fs boshsys.FileSystem,
) FSIndexBlobs {
	return FSIndexBlobs{
		dirPath:  dirPath,
		reporter: reporter,

		blobstore: blobstore,
		fs:        fs,
	}
}

// Get gurantees that returned file matches requested digest string.
func (c FSIndexBlobs) Get(name string, blobID string, digestString string) (string, error) {
	dstPath, err := c.blobPath(digestString)
	if err != nil {
		return "", err
	}

	if c.fs.FileExists(dstPath) {
		digest, err := boshcrypto.ParseMultipleDigest(digestString)
		if err != nil {
			return "", err
		}

		err = digest.VerifyFilePath(dstPath, c.fs)
		if err != nil {
			errMsg := "Local copy ('%s') of blob '%s' digest verification error"
			return "", bosherr.WrapErrorf(err, errMsg, dstPath, blobID)
		}

		return dstPath, nil
	}

	if c.blobstore != nil && len(blobID) > 0 {
		desc := fmt.Sprintf("sha1=%s", digestString)

		digest, err := boshcrypto.ParseMultipleDigest(digestString)
		if err != nil {
			return "", bosherr.WrapErrorf(err, "Downloading blob '%s' with digest '%s'", blobID, digestString)
		}

		c.reporter.IndexEntryDownloadStarted(name, desc)

		path, err := c.blobstore.Get(blobID, digest)
		if err != nil {
			c.reporter.IndexEntryDownloadFinished(name, desc, err)
			return "", bosherr.WrapErrorf(err, "Downloading blob '%s' with digest string '%s'", blobID, digestString)
		}

		err = boshfu.NewFileMover(c.fs).Move(path, dstPath)
		if err != nil {
			c.reporter.IndexEntryDownloadFinished(name, desc, err)
			return "", bosherr.WrapErrorf(err, "Moving blob '%s' into cache", blobID)
		}

		c.reporter.IndexEntryDownloadFinished(name, desc, nil)

		return dstPath, nil
	}

	if len(blobID) == 0 {
		return "", bosherr.Errorf("Cannot find blob named '%s' with SHA1 '%s'", name, digestString)
	}

	return "", bosherr.Errorf("Cannot find blob '%s' with SHA1 '%s'", blobID, digestString)
}

// Add adds file to cache and blobstore but does not guarantee
// that file have expected SHA1 when retrieved later.
func (c FSIndexBlobs) Add(name, path, sha1 string) (string, string, error) {
	dstPath, err := c.blobPath(sha1)
	if err != nil {
		return "", "", err
	}

	if !c.fs.FileExists(dstPath) {
		err := c.fs.CopyFile(path, dstPath)
		if err != nil {
			return "", "", bosherr.WrapErrorf(err, "Copying file '%s' with SHA1 '%s' into cache", path, sha1)
		}
	}

	if c.blobstore != nil {
		desc := fmt.Sprintf("sha1=%s", sha1)

		c.reporter.IndexEntryUploadStarted(name, desc)

		blobID, _, err := c.blobstore.Create(path)
		if err != nil {
			c.reporter.IndexEntryUploadFinished(name, desc, err)
			return "", "", bosherr.WrapErrorf(err, "Creating blob for path '%s'", path)
		}

		c.reporter.IndexEntryUploadFinished(name, desc, nil)

		return blobID, dstPath, nil
	}

	return "", dstPath, nil
}

func (c FSIndexBlobs) blobPath(sha1 string) (string, error) {
	absDirPath, err := c.fs.ExpandPath(c.dirPath)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Expanding cache directory")
	}

	err = c.fs.MkdirAll(absDirPath, os.ModePerm)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Creating cache directory")
	}

	return filepath.Join(absDirPath, sha1), nil
}
