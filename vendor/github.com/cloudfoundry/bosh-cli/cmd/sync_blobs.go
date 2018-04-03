package cmd

import (
	boshreldir "github.com/cloudfoundry/bosh-cli/releasedir"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type SyncBlobsCmd struct {
	blobsDir             boshreldir.BlobsDir
	numOfParallelWorkers int
}

func NewSyncBlobsCmd(blobsDir boshreldir.BlobsDir, numOfParallelWorkers int) SyncBlobsCmd {
	return SyncBlobsCmd{blobsDir: blobsDir, numOfParallelWorkers: numOfParallelWorkers}
}

func (c SyncBlobsCmd) Run() error {
	err := c.blobsDir.SyncBlobs(c.numOfParallelWorkers)
	if err != nil {
		return bosherr.WrapErrorf(err, "Downloading blobs")
	}

	return nil
}
