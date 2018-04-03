package ui

import (
	"github.com/dustin/go-humanize"
)

type BlobsReporter struct {
	ui UI
}

func NewBlobsReporter(ui UI) BlobsReporter {
	return BlobsReporter{ui: ui}
}

func (r BlobsReporter) BlobDownloadStarted(path string, size int64, blobID, sha1 string) {
	r.ui.BeginLinef("Blob download '%s' (%s) (id: %s sha1: %s) started\n",
		path, humanize.Bytes(uint64(size)), blobID, sha1)
}

func (r BlobsReporter) BlobDownloadFinished(path, blobID string, err error) {
	if err != nil {
		r.ui.ErrorLinef("Blob download '%s' (id: %s) failed", path, blobID)
	} else {
		r.ui.BeginLinef("Blob download '%s' (id: %s) finished\n", path, blobID)
	}
}

func (r BlobsReporter) BlobUploadStarted(path string, size int64, sha1 string) {
	r.ui.BeginLinef("Blob upload '%s' (%s) (sha1: %s) started\n",
		path, humanize.Bytes(uint64(size)), sha1)
}

func (r BlobsReporter) BlobUploadFinished(path, blobID string, err error) {
	if err != nil {
		r.ui.ErrorLinef("Blob upload '%s' failed", path)
	} else {
		r.ui.BeginLinef("Blob upload '%s' (id: %s) finished\n", path, blobID)
	}
}
