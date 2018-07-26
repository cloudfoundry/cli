package v2action

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/util"
	pb "gopkg.in/cheggaaa/pb.v1"
)

type Buildpack ccv2.Buildpack

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
		if _, ok := err.(ccerror.BuildpackAlreadyExistsForStackError); ok {
			return Warnings(warnings), actionerror.BuildpackAlreadyExistsForStackError{Message: err.Error()}
		}
		return Warnings(warnings), err
	}

	progBar.Terminate()
	return Warnings(warnings), nil
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
			header.SetMode(info.Mode())
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
