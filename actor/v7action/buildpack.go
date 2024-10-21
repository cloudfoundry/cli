package v7action

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/v7/actor/actionerror"
	"code.cloudfoundry.org/cli/v7/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/v7/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v7/resources"
	"code.cloudfoundry.org/cli/v7/util"
)

type JobURL ccv3.JobURL

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Downloader

type Downloader interface {
	Download(url string, tmpDirPath string) (string, error)
}

func (actor Actor) GetBuildpacks(labelSelector string) ([]resources.Buildpack, Warnings, error) {
	queries := []ccv3.Query{ccv3.Query{Key: ccv3.OrderBy, Values: []string{ccv3.PositionOrder}}}
	if labelSelector != "" {
		queries = append(queries, ccv3.Query{Key: ccv3.LabelSelectorFilter, Values: []string{labelSelector}})
	}

	buildpacks, warnings, err := actor.CloudControllerClient.GetBuildpacks(queries...)

	return buildpacks, Warnings(warnings), err
}

// GetBuildpackByNameAndStack returns a buildpack with the provided name and
// stack. If `buildpackStack` is not specified, and there are multiple
// buildpacks with the same name, it will return the one with no stack, if
// present.
func (actor Actor) GetBuildpackByNameAndStack(buildpackName string, buildpackStack string) (resources.Buildpack, Warnings, error) {
	var (
		buildpacks []resources.Buildpack
		warnings   ccv3.Warnings
		err        error
	)

	if buildpackStack == "" {
		buildpacks, warnings, err = actor.CloudControllerClient.GetBuildpacks(ccv3.Query{
			Key:    ccv3.NameFilter,
			Values: []string{buildpackName},
		})
	} else {
		buildpacks, warnings, err = actor.CloudControllerClient.GetBuildpacks(
			ccv3.Query{
				Key:    ccv3.NameFilter,
				Values: []string{buildpackName},
			},
			ccv3.Query{
				Key:    ccv3.StackFilter,
				Values: []string{buildpackStack},
			},
		)
	}

	if err != nil {
		return resources.Buildpack{}, Warnings(warnings), err
	}

	if len(buildpacks) == 0 {
		return resources.Buildpack{}, Warnings(warnings), actionerror.BuildpackNotFoundError{BuildpackName: buildpackName, StackName: buildpackStack}
	}

	if len(buildpacks) > 1 {
		for _, buildpack := range buildpacks {
			if buildpack.Stack == "" {
				return buildpack, Warnings(warnings), nil
			}
		}
		return resources.Buildpack{}, Warnings(warnings), actionerror.MultipleBuildpacksFoundError{BuildpackName: buildpackName}
	}

	return buildpacks[0], Warnings(warnings), err
}

func (actor Actor) CreateBuildpack(buildpack resources.Buildpack) (resources.Buildpack, Warnings, error) {
	buildpack, warnings, err := actor.CloudControllerClient.CreateBuildpack(buildpack)

	return buildpack, Warnings(warnings), err
}

func (actor Actor) UpdateBuildpackByNameAndStack(buildpackName string, buildpackStack string, buildpack resources.Buildpack) (resources.Buildpack, Warnings, error) {
	var warnings Warnings
	foundBuildpack, getWarnings, err := actor.GetBuildpackByNameAndStack(buildpackName, buildpackStack)
	warnings = append(warnings, getWarnings...)

	if err != nil {
		return resources.Buildpack{}, warnings, err
	}

	buildpack.GUID = foundBuildpack.GUID

	updatedBuildpack, updateWarnings, err := actor.CloudControllerClient.UpdateBuildpack(resources.Buildpack(buildpack))
	warnings = append(warnings, updateWarnings...)
	if err != nil {
		return resources.Buildpack{}, warnings, err
	}

	return updatedBuildpack, warnings, nil
}

func (actor Actor) UploadBuildpack(guid string, pathToBuildpackBits string, progressBar SimpleProgressBar) (ccv3.JobURL, Warnings, error) {
	wrappedReader, size, err := progressBar.Initialize(pathToBuildpackBits)
	if err != nil {
		return "", Warnings{}, err
	}

	defer progressBar.Terminate()

	jobURL, warnings, err := actor.CloudControllerClient.UploadBuildpack(guid, pathToBuildpackBits, wrappedReader, size)
	if err != nil {
		// TODO: Do we actually want to convert this error? Is this the right place?
		if e, ok := err.(ccerror.BuildpackAlreadyExistsForStackError); ok {
			return "", Warnings(warnings), actionerror.BuildpackAlreadyExistsForStackError{Message: e.Message}
		}
		return "", Warnings(warnings), err
	}

	return jobURL, Warnings(warnings), nil
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

func (actor Actor) DeleteBuildpackByNameAndStack(buildpackName string, buildpackStack string) (Warnings, error) {
	var allWarnings Warnings
	buildpack, getBuildpackWarnings, err := actor.GetBuildpackByNameAndStack(buildpackName, buildpackStack)
	allWarnings = append(allWarnings, getBuildpackWarnings...)
	if err != nil {
		return allWarnings, err
	}

	jobURL, deleteBuildpackWarnings, err := actor.CloudControllerClient.DeleteBuildpack(buildpack.GUID)
	allWarnings = append(allWarnings, deleteBuildpackWarnings...)
	if err != nil {
		return allWarnings, err
	}

	pollWarnings, err := actor.CloudControllerClient.PollJob(jobURL)
	allWarnings = append(allWarnings, pollWarnings...)

	return allWarnings, err
}
