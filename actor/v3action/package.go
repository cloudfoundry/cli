package v3action

import (
	"archive/zip"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/gofileutils/fileutils"
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

func (actor Actor) CreateAndUploadPackageByApplicationNameAndSpace(appName string, spaceGUID string, bitsPath string) (Package, Warnings, error) {
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

	err = writeZipFile(bitsPath, tmpZipFilepath)
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

func writeZipFile(dir string, targetFile *os.File) error {
	isEmpty, err := fileutils.IsDirEmpty(dir)
	if err != nil {
		return err
	}

	if isEmpty {
		return errors.NewEmptyDirError(dir)
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
