package v2action

import (
	"archive/zip"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	log "github.com/Sirupsen/logrus"
)

type Resource ccv2.Resource

// ZipResources zips a directory and a sorted (based on full path/filename)
// list of resources and returns the location. On Windows, the filemode for
// user is forced to be readable and executable.
func (actor Actor) ZipResources(sourceDir string, filesToInclude []Resource) (string, error) {
	log.WithField("sourceDir", sourceDir).Info("zipping source files")
	zipFile, err := ioutil.TempFile("", "cf-cli-")
	if err != nil {
		return "", err
	}
	defer zipFile.Close()

	writer := zip.NewWriter(zipFile)
	defer writer.Close()

	for _, resource := range filesToInclude {
		fullPath := filepath.Join(sourceDir, resource.Filename)
		log.WithField("fullPath", fullPath).Debug("zipping file")
		err := actor.addFileToZip(fullPath, resource.Filename, writer)
		if err != nil {
			log.WithField("fullPath", fullPath).Errorln("zipping file:", err)
			return "", err
		}
	}

	log.WithFields(log.Fields{
		"zip_file_location": zipFile.Name(),
		"zipped_file_count": len(filesToInclude),
	}).Info("zip file created")
	return zipFile.Name(), nil
}

func (_ Actor) addFileToZip(srcPath string, destPath string, zipFile *zip.Writer) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		log.WithField("srcPath", srcPath).Errorln("opening path in dir:", err)
		return err
	}
	defer srcFile.Close()

	fileInfo, err := srcFile.Stat()
	if err != nil {
		log.WithField("srcPath", srcPath).Errorln("stat error in dir:", err)
		return err
	}
	// check error here
	header, err := zip.FileInfoHeader(fileInfo)
	if err != nil {
		log.WithField("srcPath", srcPath).Errorln("getting file info in dir:", err)
		return err
	}
	header.Name = destPath
	header.Method = zip.Deflate

	mode := fixMode(fileInfo.Mode())
	header.SetMode(mode)
	log.WithFields(log.Fields{
		"srcPath": srcPath,
		"mode":    mode,
	}).Debug("setting mode for file")

	destFileWriter, err := zipFile.CreateHeader(header)
	if err != nil {
		log.Errorln("creating header:", err)
		return err
	}

	if !fileInfo.IsDir() {
		if _, err := io.Copy(destFileWriter, srcFile); err != nil {
			log.WithField("srcPath", srcPath).Errorln("copying data in dir:", err)
			return err
		}
	}

	return nil
}

func (_ Actor) containedInFiles(path string, fileList []Resource) bool {
	for _, resource := range fileList {
		if resource.Filename == path {
			return true
		}
	}

	return false
}
