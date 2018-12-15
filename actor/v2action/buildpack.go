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

	"gopkg.in/cheggaaa/pb.v1"
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

func (actor *Actor) getBuildpacks(name string, stack string) ([]Buildpack, Warnings, error) {
	var filters []ccv2.Filter

	bpName := ccv2.Filter{
		Type:     constant.NameFilter,
		Operator: constant.EqualOperator,
		Values:   []string{name},
	}
	filters = append(filters, bpName)

	if len(stack) > 0 {
		stackFilter := ccv2.Filter{
			Type:     constant.StackFilter,
			Operator: constant.EqualOperator,
			Values:   []string{stack},
		}
		filters = append(filters, stackFilter)
	}

	ccv2Buildpacks, warnings, err := actor.CloudControllerClient.GetBuildpacks(filters...)
	if err != nil {
		return nil, Warnings(warnings), err
	}

	var buildpacks []Buildpack
	for _, buildpack := range ccv2Buildpacks {
		buildpacks = append(buildpacks, Buildpack(buildpack))
	}

	if len(buildpacks) == 0 {
		return nil, Warnings(warnings), actionerror.BuildpackNotFoundError{BuildpackName: name, StackName: stack}
	}

	return buildpacks, Warnings(warnings), nil
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

func (actor *Actor) RenameBuildpack(oldName string, newName string, stackName string) (Warnings, error) {
	var (
		getWarnings Warnings
		allWarnings Warnings

		foundBuildpacks []Buildpack
		oldBp           Buildpack
		err             error
		found           bool
	)

	foundBuildpacks, getWarnings, err = actor.getBuildpacks(oldName, stackName)
	allWarnings = append(allWarnings, getWarnings...)
	if err != nil {
		return allWarnings, err
	}

	if len(foundBuildpacks) == 1 {
		oldBp = foundBuildpacks[0]
	} else {
		if stackName == "" {
			for _, bp := range foundBuildpacks {
				if bp.NoStack() {
					oldBp = bp
					found = true
					break
				}
			}
		}

		if !found {
			return allWarnings, actionerror.MultipleBuildpacksFoundError{BuildpackName: oldName}
		}
	}

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

func (actor *Actor) UpdateBuildpackByNameAndStack(name, currentStack string, position types.NullInt, locked types.NullBool, enabled types.NullBool, newStack string) (string, Warnings, error) {
	warnings := Warnings{}
	var (
		buildpack    Buildpack
		execWarnings Warnings
		err          error
	)

	execWarnings, err = actor.checkIfNewStackExists(newStack)
	warnings = append(warnings, execWarnings...)

	if err != nil {
		return "", warnings, err
	}

	var buildpacks []Buildpack

	buildpacks, execWarnings, err = actor.getBuildpacks(name, currentStack)

	warnings = append(warnings, execWarnings...)
	if err != nil {
		return "", warnings, err
	}

	allBuildpacksHaveStacks := true
	for _, buildpack := range buildpacks {
		if buildpack.NoStack() {
			allBuildpacksHaveStacks = false
		}
	}
	if allBuildpacksHaveStacks && len(newStack) > 0 {
		return "", warnings, actionerror.BuildpackStackChangeError{
			BuildpackName: buildpacks[0].Name,
			BinaryName:    actor.Config.BinaryName(),
		}
	} else if allBuildpacksHaveStacks && len(buildpacks) > 1 {
		return "", Warnings(warnings), actionerror.MultipleBuildpacksFoundError{BuildpackName: name}
	}

	buildpack = buildpacks[0]
	if position != buildpack.Position || locked != buildpack.Enabled || enabled != buildpack.Enabled || newStack != buildpack.Stack {
		buildpack.Position = position
		buildpack.Locked = locked
		buildpack.Enabled = enabled
		buildpack.Stack = newStack

		_, execWarnings, err = actor.UpdateBuildpack(buildpack)
		warnings = append(warnings, execWarnings...)

		if err != nil {
			return "", warnings, err
		}
	}

	return buildpack.GUID, warnings, err
}

func (actor *Actor) checkIfNewStackExists(newStack string) (Warnings, error) {
	if len(newStack) > 0 {
		_, execWarnings, err := actor.GetStackByName(newStack)
		if err != nil {
			return execWarnings, err
		}
	}
	return nil, nil
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
