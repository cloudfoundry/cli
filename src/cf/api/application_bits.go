package api

import (
	"bytes"
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
)

type ApplicationBitsRepository interface {
	Upload(app cf.Application, zipBuffer *bytes.Buffer) (apiErr *net.ApiError)
}

type CloudControllerApplicationBitsRepository struct {
	config  *configuration.Configuration
	gateway net.Gateway
}

func NewCloudControllerApplicationBitsRepository(config *configuration.Configuration, gateway net.Gateway) (repo CloudControllerApplicationBitsRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerApplicationBitsRepository) Upload(app cf.Application, zipBuffer *bytes.Buffer) (apiErr *net.ApiError) {
	url := fmt.Sprintf("%s/v2/apps/%s/bits", repo.config.Target, app.Guid)

	body, boundary, err := createApplicationUploadBody(zipBuffer)
	if err != nil {
		apiErr = net.NewApiErrorWithError("Error creating upload", err)
		return
	}

	request, apiErr := repo.gateway.NewRequest("PUT", url, repo.config.AccessToken, body)
	contentType := fmt.Sprintf("multipart/form-data; boundary=%s", boundary)
	request.Header.Set("Content-Type", contentType)
	if apiErr != nil {
		return
	}

	apiErr = repo.gateway.PerformRequest(request)
	return
}

func createApplicationUploadBody(zipBuffer *bytes.Buffer) (body *bytes.Buffer, boundary string, err error) {
	body = new(bytes.Buffer)

	writer := multipart.NewWriter(body)
	boundary = writer.Boundary()

	part, err := writer.CreateFormField("resources")
	if err != nil {
		return
	}

	_, err = io.Copy(part, bytes.NewBufferString("[]"))
	if err != nil {
		return
	}

	part, err = createZipPartWriter(zipBuffer, writer)
	if err != nil {
		return
	}

	_, err = io.Copy(part, zipBuffer)
	if err != nil {
		return
	}

	err = writer.Close()
	return
}

func createZipPartWriter(zipBuffer *bytes.Buffer, writer *multipart.Writer) (io.Writer, error) {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="application"; filename="application.zip"`)
	h.Set("Content-Type", "application/zip")
	h.Set("Content-Length", fmt.Sprintf("%d", zipBuffer.Len()))
	h.Set("Content-Transfer-Encoding", "binary")
	return writer.CreatePart(h)
}
