package blobstore

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

// LocalBlob represents a local copy of a blob retrieved from the blobstore
type LocalBlob interface {
	// Path returns the path to the local copy of the blob
	Path() string
	// Delete removes the local copy of the blob (does not effect the blobstore)
	Delete() error
	// DeleteSilently removes the local copy of the blob (does not effect the blobstore), logging instead of returning an error.
	DeleteSilently()
}

type localBlob struct {
	path   string
	fs     boshsys.FileSystem
	logger boshlog.Logger
	logTag string
}

func NewLocalBlob(path string, fs boshsys.FileSystem, logger boshlog.Logger) LocalBlob {
	return &localBlob{
		path:   path,
		fs:     fs,
		logger: logger,
		logTag: "localBlob",
	}
}

func (b *localBlob) Path() string {
	return b.path
}

func (b *localBlob) Delete() error {
	err := b.fs.RemoveAll(b.path)
	if err != nil {
		return bosherr.WrapErrorf(err, "Deleting local blob '%s'", b.path)
	}
	return nil
}

func (b *localBlob) DeleteSilently() {
	err := b.Delete()
	if err != nil {
		b.logger.Error(b.logTag, "Failed to delete local blob: %s", err.Error())
	}
}

func (b *localBlob) String() string {
	return fmt.Sprintf("localBlob{path: '%s'}", b.path)
}
