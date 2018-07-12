package v2action

import (
	"fmt"
	"io"
	"os"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	pb "gopkg.in/cheggaaa/pb.v1"
)

type Buildpack struct {
	GUID string
	// Name     string
	// Position int
	// Enabled  bool
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
	fmt.Printf("file %v", file)

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, 0, err
	}
	fmt.Printf("fileinfo %v", fileInfo)
	fmt.Printf("fileinfosize %v", fileInfo.Size())

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
	if _, ok := err.(ccerror.BuildpackAlreadyExistsError); ok {
		return Buildpack{}, Warnings(warnings), actionerror.BuildpackAlreadyExistsError(name)
	}

	return Buildpack{GUID: ccBuildpack.GUID}, Warnings(warnings), err
}

func (actor *Actor) UploadBuildpack(GUID string, path string, progBar SimpleProgressBar) (Warnings, error) {
	progressBarReader, size, _ := progBar.Initialize(path)
	warnings, err := actor.CloudControllerClient.UploadBuildpack(GUID, progressBarReader, size)

	if err != nil {
		return Warnings(warnings), err
	}

	progBar.Terminate()
	return Warnings(warnings), nil
}
