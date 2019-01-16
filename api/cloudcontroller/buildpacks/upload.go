package buildpacks

import (
	"bytes"
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"io"
	"mime/multipart"
	"path/filepath"
)

func CalculateRequestSize(buildpackSize int64, bpPath string) (int64, error) {
	body := &bytes.Buffer{}
	form := multipart.NewWriter(body)

	bpFileName := filepath.Base(bpPath)

	_, err := form.CreateFormFile("buildpack", bpFileName)
	if err != nil {
		return 0, err
	}

	err = form.Close()
	if err != nil {
		return 0, err
	}

	return int64(body.Len()) + buildpackSize, nil
}

func CreateMultipartBodyAndHeader(buildpack io.Reader, bpPath string) (string, io.ReadSeeker, <-chan error) {
	writerOutput, writerInput := cloudcontroller.NewPipeBomb()

	form := multipart.NewWriter(writerInput)

	writeErrors := make(chan error)

	go func() {
		defer close(writeErrors)
		defer writerInput.Close()

		bpFileName := filepath.Base(bpPath)
		writer, err := form.CreateFormFile("buildpack", bpFileName)
		if err != nil {
			writeErrors <- err
			return
		}

		_, err = io.Copy(writer, buildpack)
		if err != nil {
			writeErrors <- err
			return
		}

		err = form.Close()
		if err != nil {
			writeErrors <- err
		}
	}()

	return form.FormDataContentType(), writerOutput, writeErrors
}
