package releasedir

import (
	"io"
	"os"
	"path/filepath"
	"sort"

	boshblob "github.com/cloudfoundry/bosh-utils/blobstore"
	boshcrypto "github.com/cloudfoundry/bosh-utils/crypto"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshfu "github.com/cloudfoundry/bosh-utils/fileutil"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"gopkg.in/yaml.v2"

	"fmt"

	bicrypto "github.com/cloudfoundry/bosh-cli/crypto"
	"github.com/cloudfoundry/bosh-cli/work"
)

type FSBlobsDir struct {
	indexPath string
	dirPath   string

	reporter         BlobsDirReporter
	blobstore        boshblob.DigestBlobstore
	digestCalculator bicrypto.DigestCalculator
	fs               boshsys.FileSystem

	logTag string
	logger boshlog.Logger
}

/*
---
golang/go1.5.1.linux-amd64.tar.gz:
  object_id: 36764f38-6274-4a5d-8faa-26c31a745cb2
  sha: 46eecd290d8803887dec718c691cc243f2175fe0
  size: 77875767
*/

type fsBlobsDirSchema map[string]fsBlobsDirSchema_Blob

type fsBlobsDirSchema_Blob struct {
	Size int64 `yaml:"size"`

	BlobstoreID string `yaml:"object_id,omitempty"`
	SHA1        string `yaml:"sha"`
}

func NewFSBlobsDir(
	dirPath string,
	reporter BlobsDirReporter,
	blobstore boshblob.DigestBlobstore,
	digestCalculator bicrypto.DigestCalculator,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) FSBlobsDir {
	return FSBlobsDir{
		indexPath: filepath.Join(dirPath, "config", "blobs.yml"),
		dirPath:   filepath.Join(dirPath, "blobs"),

		reporter:         reporter,
		blobstore:        blobstore,
		digestCalculator: digestCalculator,
		fs:               fs,

		logTag: "releasedir.FSBlobsDir",
		logger: logger,
	}
}

func (d FSBlobsDir) Init() error {
	err := d.fs.MkdirAll(filepath.Dir(d.indexPath), os.ModePerm)
	if err != nil {
		return bosherr.WrapErrorf(err, "Creating blobs/")
	}

	err = d.fs.WriteFileString(d.indexPath, "--- {}\n")
	if err != nil {
		return bosherr.WrapErrorf(err, "Initing blobs.yml")
	}

	return nil
}

func (d FSBlobsDir) Blobs() ([]Blob, error) {
	bytes, err := d.fs.ReadFile(d.indexPath)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Reading blobs index")
	}

	var schema fsBlobsDirSchema

	err = yaml.Unmarshal(bytes, &schema)
	if err != nil {
		return nil, bosherr.WrapError(err, "Unmarshalling blobs index")
	}

	var blobs []Blob

	for recPath, rec := range schema {
		blobs = append(blobs, Blob{
			Path:        recPath,
			Size:        rec.Size,
			BlobstoreID: rec.BlobstoreID,
			SHA1:        rec.SHA1,
		})
	}

	sort.Sort(BlobSorting(blobs))

	return blobs, nil
}

func (d FSBlobsDir) SyncBlobs(parallel int) error {
	pool := work.Pool{
		Count: parallel,
	}

	blobs, err := d.Blobs()
	if err != nil {
		return err
	}

	symlinksFound, err := d.containsSymlinks()
	if err != nil {
		return bosherr.WrapErrorf(err, "Syncing blobs")
	}

	if symlinksFound {
		return bosherr.Error("Bailing because symlinks found in blobs directory. If switching from CLI v1, please use the `reset-release` command.")
	}

	if err := d.removeUnknownBlobs(blobs); err != nil {
		return bosherr.WrapErrorf(err, "Syncing blobs")
	}

	var tasks []func() error
	for _, blob := range blobs {
		blob := blob
		tasks = append(tasks, func() error {
			if len(blob.BlobstoreID) > 0 {
				return d.downloadBlob(blob)
			}

			return nil
		})
	}

	err = pool.ParallelDo(tasks...)
	if err != nil {
		return err
	}

	return nil
}

func (d FSBlobsDir) removeUnknownBlobs(blobs []Blob) error {
	files, err := d.fs.RecursiveGlob(filepath.Join(d.dirPath, "**/*"))
	if err != nil {
		return bosherr.WrapErrorf(err, "Checking for unknown blobs")
	}

	for _, file := range files {
		fileInfo, err := d.fs.Stat(file)
		if err != nil {
			return bosherr.WrapErrorf(err, "Determining existing blobs")
		}
		if fileInfo.IsDir() {
			continue
		}

		found := false

		for _, blob := range blobs {
			if file == filepath.Join(d.dirPath, blob.Path) {
				found = true
				break
			}
		}

		if !found {
			d.logger.Info(d.logTag, fmt.Sprintf("Deleting blob at '%s' that is not in the blob index.", file))
			if err := d.fs.RemoveAll(file); err != nil {
				return bosherr.WrapErrorf(err, "Removing unknown blob")
			}
		}
	}

	return nil
}

func (d FSBlobsDir) TrackBlob(path string, src io.ReadCloser) (Blob, error) {
	tempFile, err := d.fs.TempFile("track-blob")
	if err != nil {
		return Blob{}, bosherr.WrapErrorf(err, "Creating temp blob")
	}

	defer tempFile.Close()

	_, err = io.Copy(tempFile, src)
	if err != nil {
		return Blob{}, bosherr.WrapErrorf(err, "Populating temp blob")
	}

	//generation of digest string
	sha1, err := d.digestCalculator.Calculate(tempFile.Name())
	if err != nil {
		return Blob{}, bosherr.WrapErrorf(err, "Calculating temp blob sha1")
	}

	fileInfo, err := tempFile.Stat()
	if err != nil {
		return Blob{}, bosherr.WrapErrorf(err, "Stating temp blob")
	}

	blobs, err := d.Blobs()
	if err != nil {
		return Blob{}, err
	}

	idx := -1

	for i, blob := range blobs {
		if blob.Path == path {
			idx = i
			break
		}
	}

	if idx == -1 {
		blobs = append(blobs, Blob{})
		idx = len(blobs) - 1
	}

	blobs[idx] = Blob{Path: path, Size: fileInfo.Size(), SHA1: sha1}

	tempFile.Close()

	err = d.moveBlobLocally(tempFile.Name(), filepath.Join(d.dirPath, path))
	if err != nil {
		return Blob{}, err
	}

	return blobs[idx], d.save(blobs)
}

func (d FSBlobsDir) UntrackBlob(path string) error {
	blobs, err := d.Blobs()
	if err != nil {
		return err
	}

	err = d.fs.RemoveAll(filepath.Join(d.dirPath, path))
	if err != nil {
		return bosherr.WrapErrorf(err, "Removing blob from blobs/")
	}

	for i, blob := range blobs {
		if blob.Path == path {
			return d.save(append(blobs[:i], blobs[i+1:]...))
		}
	}

	return nil
}

func (d FSBlobsDir) UploadBlobs() error {
	blobs, err := d.Blobs()
	if err != nil {
		return err
	}

	for i, blob := range blobs {
		if len(blob.BlobstoreID) == 0 {
			blobID, err := d.uploadBlob(blob)
			if err != nil {
				return err
			}

			blob.BlobstoreID = blobID

			blobs[i] = blob

			err = d.save(blobs)
			if err != nil {
				return bosherr.WrapErrorf(
					err, "Saving newly created blob '%s' for path '%s'", blobID, blob.Path)
			}
		}
	}

	return nil
}

func (d FSBlobsDir) containsSymlinks() (bool, error) {
	files, err := d.fs.RecursiveGlob(filepath.Join(d.dirPath, "**/*"))
	if err != nil {
		return false, nil
	}

	for _, file := range files {
		fileInfo, err := d.fs.Lstat(file)
		if err != nil {
			return false, err
		}

		if fileInfo.Mode()&os.ModeSymlink != 0 {
			return true, nil
		}
	}

	return false, nil
}

func (d FSBlobsDir) checkBlobExistence(dstPath string, digest boshcrypto.MultipleDigest) bool {
	if d.fs.FileExists(dstPath) {
		if err := digest.VerifyFilePath(dstPath, d.fs); err != nil {
			d.logger.Error(d.logTag, fmt.Sprintf("Incorrect SHA sum for blob at '%s'. Re-downloading from blobstore.", dstPath))
			return false
		}

		return true
	}

	return false
}

func (d FSBlobsDir) downloadBlob(blob Blob) error {
	dstPath := filepath.Join(d.dirPath, blob.Path)

	digest, err := boshcrypto.ParseMultipleDigest(blob.SHA1)
	if err != nil {
		return bosherr.WrapErrorf(
			err, "Generating multi digest for blob '%s' for path '%s' with digest string '%s'", blob.BlobstoreID, blob.Path, blob.SHA1)
	}

	if d.checkBlobExistence(dstPath, digest) {
		return nil
	}

	d.reporter.BlobDownloadStarted(blob.Path, blob.Size, blob.BlobstoreID, blob.SHA1)

	path, err := d.blobstore.Get(blob.BlobstoreID, digest)
	if err != nil {
		d.reporter.BlobDownloadFinished(blob.Path, blob.BlobstoreID, err)
		return bosherr.WrapErrorf(
			err, "Getting blob '%s' for path '%s'", blob.BlobstoreID, blob.Path)
	}

	d.reporter.BlobDownloadFinished(blob.Path, blob.BlobstoreID, nil)

	return d.moveBlobLocally(path, dstPath)
}

func (d FSBlobsDir) uploadBlob(blob Blob) (string, error) {
	var blobID string

	d.reporter.BlobUploadStarted(blob.Path, blob.Size, blob.SHA1)

	srcPath := filepath.Join(d.dirPath, blob.Path)

	blobID, _, err := d.blobstore.Create(srcPath)
	if err != nil {
		d.reporter.BlobUploadFinished(blob.Path, "", err)
		return "", bosherr.WrapErrorf(err, "Creating blob for path '%s'", blob.Path)
	}

	d.reporter.BlobUploadFinished(blob.Path, blobID, nil)

	return blobID, nil
}

func (d FSBlobsDir) moveBlobLocally(srcPath, dstPath string) error {
	err := d.fs.MkdirAll(filepath.Dir(dstPath), os.ModePerm)
	if err != nil {
		return bosherr.WrapErrorf(err, "Creating subdirs in blobs/")
	}

	err = boshfu.NewFileMover(d.fs).Move(srcPath, dstPath)
	if err != nil {
		return bosherr.WrapErrorf(err, "Moving temp blob to blobs/")
	}

	return nil
}

func (d FSBlobsDir) save(blobs []Blob) error {
	schema := fsBlobsDirSchema{}

	for _, blob := range blobs {
		schema[blob.Path] = fsBlobsDirSchema_Blob{
			Size:        blob.Size,
			BlobstoreID: blob.BlobstoreID,
			SHA1:        blob.SHA1,
		}
	}

	bytes, err := yaml.Marshal(schema)
	if err != nil {
		return bosherr.WrapError(err, "Marshalling blobs index")
	}

	err = d.fs.WriteFile(d.indexPath, bytes)
	if err != nil {
		return bosherr.WrapErrorf(err, "Writing blobs index")
	}

	return nil
}

type BlobSorting []Blob

func (s BlobSorting) Len() int {
	return len(s)
}
func (s BlobSorting) Less(i, j int) bool {
	return s[i].Path < s[j].Path
}
func (s BlobSorting) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
