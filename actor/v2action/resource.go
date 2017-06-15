package v2action

import (
	"archive/zip"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/ykk"
	log "github.com/sirupsen/logrus"
)

type FileChangedError struct {
	Filename string
}

func (e FileChangedError) Error() string {
	return fmt.Sprint("SHA1 mismatch for:", e.Filename)
}

type Resource ccv2.Resource

// GatherArchiveResources returns a list of resources for a directory.
func (_ Actor) GatherArchiveResources(archivePath string) ([]Resource, error) {
	var resources []Resource

	archive, err := os.Open(archivePath)
	if err != nil {
		return nil, err
	}
	defer archive.Close()

	info, err := archive.Stat()
	if err != nil {
		return nil, err
	}

	reader, err := ykk.NewReader(archive, info.Size())
	if err != nil {
		return nil, err
	}

	for _, archivedFile := range reader.File {
		resource := Resource{Filename: filepath.ToSlash(archivedFile.Name)}
		if !archivedFile.FileInfo().IsDir() {
			fileReader, err := archivedFile.Open()
			if err != nil {
				return nil, err
			}
			defer fileReader.Close()

			hash := sha1.New()

			_, err = io.Copy(hash, fileReader)
			if err != nil {
				return nil, err
			}
			info := archivedFile.FileInfo()

			resource.Size = archivedFile.FileInfo().Size()
			resource.SHA1 = fmt.Sprintf("%x", hash.Sum(nil))
			resource.Mode = info.Mode()
		}
		resources = append(resources, resource)
	}
	return resources, nil
}

// GatherDirectoryResources returns a list of resources for a directory.
func (_ Actor) GatherDirectoryResources(sourceDir string) ([]Resource, error) {
	var resources []Resource
	walkErr := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		resource := Resource{
			Filename: filepath.ToSlash(relPath),
		}

		if !info.IsDir() {
			resource.Size = info.Size()
			resource.Mode = fixMode(info.Mode())
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			sum := sha1.New()
			_, err = io.Copy(sum, file)
			if err != nil {
				return err
			}
			resource.SHA1 = fmt.Sprintf("%x", sum.Sum(nil))
		}
		resources = append(resources, resource)
		return nil
	})

	return resources, walkErr
}

// ZipDirectoryResources zips a directory and a sorted (based on full
// path/filename) list of resources and returns the location. On Windows, the
// filemode for user is forced to be readable and executable.
func (actor Actor) ZipDirectoryResources(sourceDir string, filesToInclude []Resource) (string, error) {
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
		err := actor.addFileToZip(fullPath, resource.Filename, resource.SHA1, writer)
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

func (_ Actor) addFileToZip(srcPath string, destPath string, sha1Sum string, zipFile *zip.Writer) error {
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
		sum := sha1.New()

		multi := io.MultiWriter(sum, destFileWriter)
		if _, err := io.Copy(multi, srcFile); err != nil {
			log.WithField("srcPath", srcPath).Errorln("copying data in dir:", err)
			return err
		}

		if sha1Sum != fmt.Sprintf("%x", sum.Sum(nil)) {
			return FileChangedError{Filename: srcPath}
		}
	}

	return nil
}
