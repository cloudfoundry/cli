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

// GatherResources returns a list of resources for a directory.
func (_ Actor) GatherResources(sourceDir string) ([]Resource, error) {
	var resources []Resource
	walkErr := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return nil
		}

		if relPath == "." {
			return nil
		}

		resources = append(resources, Resource{
			Filename: filepath.ToSlash(relPath),
		})

		return nil
	})

	return resources, walkErr
}

func (actor Actor) UploadApplicationPackage(appGUID string, existingResources []Resource, newResources io.Reader, newResourcesLength int64) (Warnings, error) {
	var allWarnings Warnings

	job, warnings, err := actor.CloudControllerClient.UploadApplicationPackage(appGUID, actor.actorToCCResources(existingResources), newResources, newResourcesLength)
	allWarnings = Warnings(warnings)
	if err != nil {
		return allWarnings, err
	}

	warnings, err = actor.CloudControllerClient.PollJob(job)
	allWarnings = append(allWarnings, Warnings(warnings)...)

	return allWarnings, err
}

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

func (_ Actor) actorToCCResources(resources []Resource) []ccv2.Resource {
	apiResources := make([]ccv2.Resource, 0, len(resources)) // Explicitly done to prevent nils

	for _, resource := range resources {
		apiResources = append(apiResources, ccv2.Resource(resource))
	}

	return apiResources
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

	header, err := zip.FileInfoHeader(fileInfo)
	if err != nil {
		log.WithField("srcPath", srcPath).Errorln("getting file info in dir:", err)
		return err
	}

	// An extra '/' indicates that this file is a directory
	if fileInfo.IsDir() {
		destPath += "/"
	}

	header.Name = destPath
	header.Method = zip.Deflate

	mode := fixMode(fileInfo.Mode())
	header.SetMode(mode)
	log.WithFields(log.Fields{
		"srcPath":  srcPath,
		"destPath": destPath,
		"mode":     mode,
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
