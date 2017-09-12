package v3action

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/gofileutils/fileutils"
	"code.cloudfoundry.org/ykk"
)

const (
	DefaultFolderPermissions      = 0755
	DefaultArchiveFilePermissions = 0744
)

type PackageProcessingFailedError struct{}

func (PackageProcessingFailedError) Error() string {
	return "Package failed to process correctly after upload"
}

type PackageProcessingExpiredError struct{}

func (PackageProcessingExpiredError) Error() string {
	return "Package expired after upload"
}

type Package ccv3.Package

type EmptyDirectoryError struct {
	Path string
}

func (e EmptyDirectoryError) Error() string {
	return fmt.Sprint(e.Path, "is empty")
}

type DockerImageCredentials struct {
	Path     string
	Username string
	Password string
}

func (actor Actor) CreatePackageByApplicationNameAndSpace(appName string, spaceGUID string, bitsPath string, dockerImageCredentials DockerImageCredentials) (Package, Warnings, error) {
	if dockerImageCredentials.Path == "" {
		if bitsPath == "" {
			var err error
			bitsPath, err = os.Getwd()
			if err != nil {
				return Package{}, nil, err
			}
		}
		return actor.createAndUploadBitsPackageByApplicationNameAndSpace(appName, spaceGUID, bitsPath)
	}
	return actor.createDockerPackageByApplicationNameAndSpace(appName, spaceGUID, dockerImageCredentials)
}

func (actor Actor) createDockerPackageByApplicationNameAndSpace(appName string, spaceGUID string, dockerImageCredentials DockerImageCredentials) (Package, Warnings, error) {
	app, allWarnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return Package{}, allWarnings, err
	}
	inputPackage := ccv3.Package{
		Type: ccv3.PackageTypeDocker,
		Relationships: ccv3.Relationships{
			ccv3.ApplicationRelationship: ccv3.Relationship{GUID: app.GUID},
		},
		DockerImage:    dockerImageCredentials.Path,
		DockerUsername: dockerImageCredentials.Username,
		DockerPassword: dockerImageCredentials.Password,
	}
	pkg, warnings, err := actor.CloudControllerClient.CreatePackage(inputPackage)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return Package{}, allWarnings, err
	}
	return Package(pkg), allWarnings, err
}

func (actor Actor) createAndUploadBitsPackageByApplicationNameAndSpace(appName string, spaceGUID string, bitsPath string) (Package, Warnings, error) {
	app, allWarnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return Package{}, allWarnings, err
	}

	inputPackage := ccv3.Package{
		Type: ccv3.PackageTypeBits,
		Relationships: ccv3.Relationships{
			ccv3.ApplicationRelationship: ccv3.Relationship{GUID: app.GUID},
		},
	}

	tmpZipFilepath, err := ioutil.TempFile("", "cli-package-upload")
	if err != nil {
		return Package{}, allWarnings, err
	}
	defer os.Remove(tmpZipFilepath.Name())

	fileInfo, err := os.Stat(bitsPath)
	if err != nil {
		return Package{}, allWarnings, err
	}

	if fileInfo.IsDir() {
		err = zipDirToFile(bitsPath, tmpZipFilepath)
	} else {
		err = copyZipArchive(bitsPath, tmpZipFilepath)
	}

	if err != nil {
		return Package{}, allWarnings, err
	}

	pkg, warnings, err := actor.CloudControllerClient.CreatePackage(inputPackage)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return Package{}, allWarnings, err
	}

	_, warnings, err = actor.CloudControllerClient.UploadPackage(pkg, tmpZipFilepath.Name())
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return Package{}, allWarnings, err
	}

	for pkg.State != ccv3.PackageStateReady &&
		pkg.State != ccv3.PackageStateFailed &&
		pkg.State != ccv3.PackageStateExpired {
		time.Sleep(actor.Config.PollingInterval())
		pkg, warnings, err = actor.CloudControllerClient.GetPackage(pkg.GUID)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return Package{}, allWarnings, err
		}
	}

	if pkg.State == ccv3.PackageStateFailed {
		return Package{}, allWarnings, PackageProcessingFailedError{}
	} else if pkg.State == ccv3.PackageStateExpired {
		return Package{}, allWarnings, PackageProcessingExpiredError{}
	}

	return Package(pkg), allWarnings, err
}

// GetApplicationPackages returns a list of package of an app.
func (actor *Actor) GetApplicationPackages(appName string, spaceGUID string) ([]Package, Warnings, error) {
	app, allWarnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return nil, allWarnings, err
	}

	ccv3Packages, warnings, err := actor.CloudControllerClient.GetPackages(url.Values{
		ccv3.AppGUIDFilter: []string{app.GUID},
	})
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	var packages []Package
	for _, ccv3Package := range ccv3Packages {
		packages = append(packages, Package(ccv3Package))
	}

	return packages, allWarnings, nil
}

func copyZipArchive(sourceArchivePath string, destZipFile *os.File) error {
	writer := zip.NewWriter(destZipFile)
	defer writer.Close()

	source, err := os.Open(sourceArchivePath)
	if err != nil {
		return err
	}
	defer source.Close()

	reader, err := newArchiveReader(source)
	if err != nil {
		return err
	}

	for _, archiveFile := range reader.File {
		reader, openErr := archiveFile.Open()
		if openErr != nil {
			return openErr
		}

		err = addFileToZipFromFileSystem(reader, archiveFile.FileInfo(), filepath.ToSlash(archiveFile.Name), writer)
		if err != nil {
			return err
		}
	}

	return nil
}

func addFileToZipFromFileSystem(srcFile io.ReadCloser, fileInfo os.FileInfo, destPath string, zipFile *zip.Writer) error {
	defer srcFile.Close()

	header, err := zip.FileInfoHeader(fileInfo)
	if err != nil {
		return err
	}

	// An extra '/' indicates that this file is a directory
	if fileInfo.IsDir() && !strings.HasSuffix(destPath, "/") {
		destPath += "/"
	}

	header.Name = destPath
	header.Method = zip.Deflate

	if fileInfo.IsDir() {
		header.SetMode(DefaultFolderPermissions)
	} else {
		header.SetMode(DefaultArchiveFilePermissions)
	}

	destFileWriter, err := zipFile.CreateHeader(header)
	if err != nil {
		return err
	}

	if !fileInfo.IsDir() {
		multi := io.Writer(destFileWriter)
		if _, err := io.Copy(multi, srcFile); err != nil {
			return err
		}
	}

	return nil
}

func zipDirToFile(dir string, targetFile *os.File) error {
	isEmpty, err := fileutils.IsDirEmpty(dir)
	if err != nil {
		return err
	}

	if isEmpty {
		return EmptyDirectoryError{Path: dir}
	}

	writer := zip.NewWriter(targetFile)
	defer writer.Close()

	return filepath.Walk(dir, func(filePath string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filePath == dir {
			return nil
		}

		fileRelativePath, _ := filepath.Rel(dir, filePath)

		header, err := zip.FileInfoHeader(fileInfo)
		if err != nil {
			return err
		}

		if runtime.GOOS == "windows" {
			header.SetMode(header.Mode() | 0700)
		}
		header.Name = filepath.ToSlash(fileRelativePath)
		header.Method = zip.Deflate

		if fileInfo.IsDir() {
			header.Name += "/"
		}

		zipFilePart, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}

		if fileInfo.IsDir() {
			return nil
		}

		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(zipFilePart, file)
		if err != nil {
			return err
		}

		return nil
	})
}

func newArchiveReader(archive *os.File) (*zip.Reader, error) {
	info, err := archive.Stat()
	if err != nil {
		return nil, err
	}

	return ykk.NewReader(archive, info.Size())
}
