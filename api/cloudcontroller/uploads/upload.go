package uploads

import (
	"bytes"
	"io"
	"mime/multipart"
	"path/filepath"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
)

func CalculateRequestSize(fileSize int64, path string, fieldName string) (int64, error) {
	body := &bytes.Buffer{}
	form := multipart.NewWriter(body)

	bpFileName := filepath.Base(path)

	_, err := form.CreateFormFile(fieldName, bpFileName)
	if err != nil {
		return 0, err
	}

	err = form.Close()
	if err != nil {
		return 0, err
	}

	return int64(body.Len()) + fileSize, nil
}

func CreateMultipartBodyAndHeader(file io.Reader, path string, fieldName string) (string, io.ReadSeeker, <-chan error) {
	writerOutput, writerInput := cloudcontroller.NewPipeBomb()

	form := multipart.NewWriter(writerInput)

	writeErrors := make(chan error)

	go func() {
		defer close(writeErrors)
		defer writerInput.Close()

		bpFileName := filepath.Base(path)
		writer, err := form.CreateFormFile(fieldName, bpFileName)
		if err != nil {
			writeErrors <- err
			return
		}

		_, err = io.Copy(writer, file)
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
