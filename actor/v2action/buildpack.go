package v2action

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/util/download"
	pb "gopkg.in/cheggaaa/pb.v1"
)

type Buildpack ccv2.Buildpack

//go:generate counterfeiter . SimpleProgressBar

type SimpleProgressBar interface {
	Initialize(path string) (io.Reader, int64, error)
	Terminate()
}

//go:generate counterfeiter . Downloader

type Downloader interface {
	Download(string) (string, error)
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
		Position: position,
		Enabled:  enabled,
	}

	ccBuildpack, warnings, err := actor.CloudControllerClient.CreateBuildpack(buildpack)
	if _, ok := err.(ccerror.BuildpackAlreadyExistsWithoutStackError); ok {
		return Buildpack{}, Warnings(warnings), actionerror.BuildpackAlreadyExistsWithoutStackError(name)
	}

	if _, ok := err.(ccerror.BuildpackNameTakenError); ok {
		return Buildpack{}, Warnings(warnings), actionerror.BuildpackNameTakenError(name)
	}

	return Buildpack{GUID: ccBuildpack.GUID}, Warnings(warnings), err
}

func (actor *Actor) UploadBuildpack(GUID string, path string, progBar SimpleProgressBar) (Warnings, error) {
	downloader := download.NewDownloader(time.Second * 30)

	pathToBuildpackBits, err := actor.PrepareBuildpackBits(path, downloader)
	if err != nil {
		return Warnings{}, err
	}

	progressBarReader, size, err := progBar.Initialize(pathToBuildpackBits)
	if err != nil {
		return Warnings{}, err
	}

	warnings, err := actor.CloudControllerClient.UploadBuildpack(GUID, pathToBuildpackBits, progressBarReader, size)
	if err != nil {
		if _, ok := err.(ccerror.BuildpackAlreadyExistsForStackError); ok {
			return Warnings(warnings), actionerror.BuildpackAlreadyExistsForStackError{Message: err.Error()}
		}
		return Warnings(warnings), err
	}

	progBar.Terminate()

	if actor.isURL(path) {
		parentDir, _ := filepath.Split(pathToBuildpackBits)
		os.RemoveAll(parentDir)
	}
	return Warnings(warnings), nil
}

func (actor *Actor) PrepareBuildpackBits(path string, downloader Downloader) (string, error) {

	if actor.isURL(path) {
		tempPath, err := downloader.Download(path)
		if err != nil {
			parentDir, _ := filepath.Split(path)
			os.RemoveAll(parentDir)

			return "", err
		}
		return tempPath, nil
	}
	return path, nil
}

func (actor *Actor) isURL(path string) bool {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return true
	}
	return false
}
