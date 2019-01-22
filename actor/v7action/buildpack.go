package v7action

import (
	"archive/zip"
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/util"
	"io"
	"os"
	"path/filepath"
)

type Buildpack ccv3.Buildpack

//go:generate counterfeiter . Downloader

type Downloader interface {
	Download(url string, tmpDirPath string) (string, error)
}

func (actor Actor) GetBuildpacks() ([]Buildpack, Warnings, error) {
	ccv3Buildpacks, warnings, err := actor.CloudControllerClient.GetBuildpacks(ccv3.Query{
		Key:    ccv3.OrderBy,
		Values: []string{ccv3.PositionOrder},
	})

	var buildpacks []Buildpack
	for _, buildpack := range ccv3Buildpacks {
		buildpacks = append(buildpacks, Buildpack(buildpack))
	}

	return buildpacks, Warnings(warnings), err
}

func (actor Actor) CreateBuildpack(buildpack Buildpack) (Buildpack, Warnings, error) {
	ccv3Buildpack, warnings, err := actor.CloudControllerClient.CreateBuildpack(
		ccv3.Buildpack(buildpack),
	)

	return Buildpack(ccv3Buildpack), Warnings(warnings), err
}

func (actor Actor) UploadBuildpack(GUID string, pathToBuildpackBits string, progressBar SimpleProgressBar) (Warnings, error) {
	wrappedReader, size, err := progressBar.Initialize(pathToBuildpackBits)
	if err != nil {
		return Warnings{}, err
	}

	warnings, err := actor.CloudControllerClient.UploadBuildpack(GUID, pathToBuildpackBits, wrappedReader, size)
	if err != nil {
		// TODO: Do we actually want to convert this error? Is this the right place?
		if e, ok := err.(ccerror.BuildpackAlreadyExistsForStackError); ok {
			return Warnings(warnings), actionerror.BuildpackAlreadyExistsForStackError{Message: e.Message}
		}
		return Warnings(warnings), err
	}

	// TODO: Should we defer the terminate instead?
	progressBar.Terminate()

	return Warnings(warnings), nil
}

func (actor *Actor) PrepareBuildpackBits(inputPath string, tmpDirPath string, downloader Downloader) (string, error) {
	if util.IsHTTPScheme(inputPath) {
		pathToDownloadedBits, err := downloader.Download(inputPath, tmpDirPath)
		if err != nil {
			return "", err
		}
		return pathToDownloadedBits, nil
	}

	if filepath.Ext(inputPath) == ".zip" {
		return inputPath, nil
	}

	info, err := os.Stat(inputPath)
	if err != nil {
		return "", err
	}

	if info.IsDir() {
		var empty bool
		empty, err = isEmptyDirectory(inputPath)
		if err != nil {
			return "", err
		}
		if empty {
			return "", actionerror.EmptyBuildpackDirectoryError{Path: inputPath}
		}
		archive := filepath.Join(tmpDirPath, filepath.Base(inputPath)) + ".zip"

		err = Zipit(inputPath, archive, "")
		if err != nil {
			return "", err
		}
		return archive, nil
	}

	return inputPath, nil
}

func isEmptyDirectory(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

// Zipit zips the source into a .zip file in the target dir
func Zipit(source, target, prefix string) error {
	// Thanks to Svett Ralchev
	// http://blog.ralch.com/tutorial/golang-working-with-zip/

	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	if prefix != "" {
		_, err = io.WriteString(zipfile, prefix)
		if err != nil {
			return err
		}
	}

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == source {
			return nil
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name, err = filepath.Rel(source, path)
		if err != nil {
			return err
		}

		header.Name = filepath.ToSlash(header.Name)
		if info.IsDir() {
			header.Name += "/"
			header.SetMode(info.Mode())
		} else {
			header.Method = zip.Deflate
			header.SetMode(fixMode(info.Mode()))
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})

	return err
}
