package v2action

import (
	"archive/zip"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util"
	"code.cloudfoundry.org/cli/util/download"

	pb "gopkg.in/cheggaaa/pb.v1"
)

type Buildpack ccv2.Buildpack

func (buildpack Buildpack) NoStack() bool {
	return len(buildpack.Stack) == 0
}

//go:generate counterfeiter . Downloader

type Downloader interface {
	Download(url string, tmpDirPath string) (string, error)
}

//go:generate counterfeiter . SimpleProgressBar

type SimpleProgressBar interface {
	Initialize(path string) (io.Reader, int64, error)
	Terminate()
}

type ProgressBar struct {
	bar *pb.ProgressBar
}

func NewProgressBar() *ProgressBar {
	return &ProgressBar{}
}

func (p *ProgressBar) Initialize(path string) (io.Reader, int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, 0, err
	}

	p.bar = pb.New(int(fileInfo.Size())).SetUnits(pb.U_BYTES)
	p.bar.ShowTimeLeft = false
	p.bar.Start()
	return p.bar.NewProxyReader(file), fileInfo.Size(), nil

}

func (p *ProgressBar) Terminate() {
	// Adding sleep to ensure UI has finished drawing
	time.Sleep(time.Second)
	p.bar.Finish()
}

func (actor *Actor) CreateBuildpack(name string, position int, enabled bool) (Buildpack, Warnings, error) {
	buildpack := ccv2.Buildpack{
		Name:     name,
		Position: types.NullInt{IsSet: true, Value: position},
		Enabled:  types.NullBool{IsSet: true, Value: enabled},
	}

	ccBuildpack, warnings, err := actor.CloudControllerClient.CreateBuildpack(buildpack)
	if _, ok := err.(ccerror.BuildpackAlreadyExistsWithoutStackError); ok {
		return Buildpack{}, Warnings(warnings), actionerror.BuildpackAlreadyExistsWithoutStackError{BuildpackName: name}
	}

	if _, ok := err.(ccerror.BuildpackNameTakenError); ok {
		return Buildpack{}, Warnings(warnings), actionerror.BuildpackNameTakenError{Name: name}
	}

	return Buildpack{GUID: ccBuildpack.GUID}, Warnings(warnings), err
}

func (actor *Actor) UploadBuildpackFromPath(inputPath, buildpackGuid string, progressBar SimpleProgressBar) (Warnings, error) {
	downloader := download.NewDownloader(time.Second * 30)
	tmpDirPath, err := ioutil.TempDir("", "buildpack-dir-")
	if err != nil {
		return Warnings{}, err
	}
	defer os.RemoveAll(tmpDirPath)

	pathToBuildpackBits, err := actor.PrepareBuildpackBits(inputPath, tmpDirPath, downloader)
	if err != nil {
		return Warnings{}, err
	}

	return actor.UploadBuildpack(buildpackGuid, pathToBuildpackBits, progressBar)
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
		archive := filepath.Join(tmpDirPath, filepath.Base(inputPath)) + ".zip"

		err = Zipit(inputPath, archive, "")
		if err != nil {
			return "", err
		}
		return archive, nil
	}

	return inputPath, nil
}

func (actor *Actor) UploadBuildpack(GUID string, pathToBuildpackBits string, progBar SimpleProgressBar) (Warnings, error) {
	progressBarReader, size, err := progBar.Initialize(pathToBuildpackBits)
	if err != nil {
		return Warnings{}, err
	}

	warnings, err := actor.CloudControllerClient.UploadBuildpack(GUID, pathToBuildpackBits, progressBarReader, size)
	if err != nil {
		if e, ok := err.(ccerror.BuildpackAlreadyExistsForStackError); ok {
			return Warnings(warnings), actionerror.BuildpackAlreadyExistsForStackError{Message: e.Message}
		}
		return Warnings(warnings), err
	}

	progBar.Terminate()
	return Warnings(warnings), nil
}

// GetBuildpackByName returns a given buildpack with the provided name. It
// assumes the stack name is empty.
func (actor *Actor) GetBuildpackByName(name string) (Buildpack, Warnings, error) {
	bpName := ccv2.Filter{
		Type:     constant.NameFilter,
		Operator: constant.EqualOperator,
		Values:   []string{name},
	}

	buildpacks, warnings, err := actor.CloudControllerClient.GetBuildpacks(bpName)
	if err != nil {
		return Buildpack{}, Warnings(warnings), err
	}

	switch len(buildpacks) {
	case 0:
		return Buildpack{}, Warnings(warnings), actionerror.BuildpackNotFoundError{BuildpackName: name}
	case 1:
		return Buildpack(buildpacks[0]), Warnings(warnings), nil
	default:
		for _, bp := range buildpacks {
			if buildpack := Buildpack(bp); buildpack.NoStack() {
				return buildpack, Warnings(warnings), nil
			}
		}
		return Buildpack{}, Warnings(warnings), actionerror.MultipleBuildpacksFoundError{BuildpackName: name}
	}
}

func (actor *Actor) GetBuildpackByNameAndStack(buildpackName string, stackName string) (Buildpack, Warnings, error) {
	bpFilter := ccv2.Filter{
		Type:     constant.NameFilter,
		Operator: constant.EqualOperator,
		Values:   []string{buildpackName},
	}

	stackFilter := ccv2.Filter{
		Type:     constant.StackFilter,
		Operator: constant.EqualOperator,
		Values:   []string{stackName},
	}

	buildpacks, warnings, err := actor.CloudControllerClient.GetBuildpacks(bpFilter, stackFilter)
	if err != nil {
		return Buildpack{}, Warnings(warnings), err
	}

	switch len(buildpacks) {
	case 0:
		return Buildpack{}, Warnings(warnings), actionerror.BuildpackNotFoundError{BuildpackName: buildpackName, StackName: stackName}
	case 1:
		return Buildpack(buildpacks[0]), Warnings(warnings), nil
	default:
		return Buildpack{}, Warnings(warnings), actionerror.MultipleBuildpacksFoundError{BuildpackName: buildpackName}
	}
}

func (actor *Actor) UpdateBuildpackByName(name string, position types.NullInt) (string, Warnings, error) {
	warnings := Warnings{}
	buildpack, execWarnings, err := actor.GetBuildpackByName(name)
	warnings = append(warnings, execWarnings...)
	if err != nil {
		return "", warnings, err
	}

	if position != buildpack.Position {
		buildpack.Position = position
		_, execWarnings, err = actor.UpdateBuildpack(buildpack)
		warnings = append(warnings, execWarnings...)
	}

	if err != nil {
		return "", warnings, err
	}

	return buildpack.GUID, warnings, err
}

func (actor *Actor) UpdateBuildpack(buildpack Buildpack) (Buildpack, Warnings, error) {
	updatedBuildpack, warnings, err := actor.CloudControllerClient.UpdateBuildpack(ccv2.Buildpack(buildpack))
	if err != nil {
		switch err.(type) {
		case ccerror.ResourceNotFoundError:
			return Buildpack{}, Warnings(warnings), actionerror.BuildpackNotFoundError{BuildpackName: buildpack.Name}
		case ccerror.BuildpackAlreadyExistsWithoutStackError:
			return Buildpack{}, Warnings(warnings), actionerror.BuildpackAlreadyExistsWithoutStackError{BuildpackName: buildpack.Name}
		case ccerror.BuildpackAlreadyExistsForStackError:
			return Buildpack{}, Warnings(warnings), actionerror.BuildpackAlreadyExistsForStackError{Message: err.Error()}
		default:
			return Buildpack{}, Warnings(warnings), err
		}
	}

	return Buildpack(updatedBuildpack), Warnings(warnings), nil
}

func (actor *Actor) RenameBuildpack(oldName string, newName string, stackName string) (Warnings, error) {
	var (
		getWarnings Warnings
		allWarnings Warnings

		oldBp Buildpack
		err   error
	)

	if len(stackName) == 0 {
		oldBp, getWarnings, err = actor.GetBuildpackByName(oldName)
	} else {
		oldBp, getWarnings, err = actor.GetBuildpackByNameAndStack(oldName, stackName)
	}

	allWarnings = append(allWarnings, getWarnings...)

	if err != nil {
		return Warnings(allWarnings), err
	}

	oldBp.Name = newName

	_, updateWarnings, err := actor.UpdateBuildpack(oldBp)
	allWarnings = append(allWarnings, updateWarnings...)
	if err != nil {
		return Warnings(allWarnings), err
	}

	return Warnings(allWarnings), nil
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
