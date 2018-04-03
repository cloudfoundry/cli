package director

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"io/ioutil"
	"os"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"gopkg.in/yaml.v2"
)

type FSArchiveWithMetadata struct {
	path     string
	fileName string
	fs       boshsys.FileSystem
}

func NewFSReleaseArchive(path string, fs boshsys.FileSystem) ReleaseArchive {
	return NewFSArchiveWithMetadata(path, "release.MF", fs)
}

func NewFSStemcellArchive(path string, fs boshsys.FileSystem) ReleaseArchive {
	return NewFSArchiveWithMetadata(path, "stemcell.MF", fs)
}

func NewFSArchiveWithMetadata(path, fileName string, fs boshsys.FileSystem) StemcellArchive {
	return FSArchiveWithMetadata{path: path, fileName: fileName, fs: fs}
}

func (a FSArchiveWithMetadata) Info() (string, string, error) {
	bytes, err := a.readMFBytes()
	if err != nil {
		return "", "", err
	}

	return a.extractNameAndVersion(bytes)
}

func (a FSArchiveWithMetadata) File() (UploadFile, error) {
	file, err := a.fs.OpenFile(a.path, os.O_RDONLY, 0)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Opening archive")
	}

	return file, nil
}

func (a FSArchiveWithMetadata) readMFBytes() ([]byte, error) {
	file, err := a.fs.OpenFile(a.path, os.O_RDONLY, 0)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Opening archive")
	}

	defer file.Close()

	gr, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}

	defer gr.Close()

	tr := tar.NewReader(gr)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Reading next tar entry")
		}

		if hdr.Name == a.fileName || hdr.Name == "./"+a.fileName {
			bytes, err := ioutil.ReadAll(tr)
			if err != nil {
				return nil, bosherr.WrapErrorf(err, "Reading '%s' entry", a.fileName)
			}

			return bytes, nil
		}
	}

	return nil, bosherr.Errorf("Missing '%s'", a.fileName)
}

func (a FSArchiveWithMetadata) extractNameAndVersion(bytes []byte) (string, string, error) {
	type mfSchema struct {
		Name    string `yaml:"name"`
		Version string `yaml:"version"`
		// other fields ignored
	}

	var mf mfSchema

	err := yaml.Unmarshal(bytes, &mf)
	if err != nil {
		return "", "", bosherr.WrapErrorf(err, "Unmarshalling '%s'", a.fileName)
	}

	return mf.Name, mf.Version, nil
}
