package application_bits

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"os"
	"time"

	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/net"
	"github.com/cloudfoundry/gofileutils/fileutils"
)

const (
	DefaultAppUploadBitsTimeout = 15 * time.Minute
)

type ApplicationBitsRepository interface {
	GetApplicationFiles(appFilesRequest []resources.AppFileResource) ([]resources.AppFileResource, error)
	UploadBits(appGuid string, zipFile *os.File, presentFiles []resources.AppFileResource) (apiErr error)
}

type CloudControllerApplicationBitsRepository struct {
	config  configuration.Reader
	gateway net.Gateway
}

func NewCloudControllerApplicationBitsRepository(config configuration.Reader, gateway net.Gateway) (repo CloudControllerApplicationBitsRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerApplicationBitsRepository) UploadBits(appGuid string, zipFile *os.File, presentFiles []resources.AppFileResource) (apiErr error) {
	url := fmt.Sprintf("%s/v2/apps/%s/bits", repo.config.ApiEndpoint(), appGuid)
	fileutils.TempFile("requests", func(requestFile *os.File, err error) {
		if err != nil {
			apiErr = errors.NewWithError(T("Error creating tmp file: {{.Err}}", map[string]interface{}{"Err": err}), err)
			return
		}

		// json.Marshal represents a nil value as "null" instead of an empty slice "[]"
		if presentFiles == nil {
			presentFiles = []resources.AppFileResource{}
		}

		presentFilesJSON, err := json.Marshal(presentFiles)
		if err != nil {
			apiErr = errors.NewWithError(T("Error marshaling JSON"), err)
			return
		}

		boundary, err := repo.writeUploadBody(zipFile, requestFile, presentFilesJSON)
		if err != nil {
			apiErr = errors.NewWithError(T("Error writing to tmp file: {{.Err}}", map[string]interface{}{"Err": err}), err)
			return
		}

		var request *net.Request
		request, apiErr = repo.gateway.NewRequest("PUT", url, repo.config.AccessToken(), requestFile)
		if apiErr != nil {
			return
		}

		contentType := fmt.Sprintf("multipart/form-data; boundary=%s", boundary)
		request.HttpReq.Header.Set("Content-Type", contentType)

		response := &resources.Resource{}
		_, apiErr = repo.gateway.PerformPollingRequestForJSONResponse(request, response, DefaultAppUploadBitsTimeout)
		if apiErr != nil {
			return
		}
	})

	return
}

func (repo CloudControllerApplicationBitsRepository) GetApplicationFiles(appFilesToCheck []resources.AppFileResource) ([]resources.AppFileResource, error) {
	allAppFilesJson, err := json.Marshal(appFilesToCheck)
	if err != nil {
		apiErr := errors.NewWithError(T("Failed to create json for resource_match request"), err)
		return nil, apiErr
	}

	presentFiles := []resources.AppFileResource{}
	apiErr := repo.gateway.UpdateResourceSync(
		repo.config.ApiEndpoint()+"/v2/resource_match",
		bytes.NewReader(allAppFilesJson),
		&presentFiles)

	if apiErr != nil {
		return nil, apiErr
	}

	return presentFiles, nil
}

func (repo CloudControllerApplicationBitsRepository) writeUploadBody(zipFile *os.File, body *os.File, presentResourcesJson []byte) (boundary string, err error) {
	writer := multipart.NewWriter(body)
	defer writer.Close()

	boundary = writer.Boundary()

	part, err := writer.CreateFormField("resources")
	if err != nil {
		return
	}

	_, err = io.Copy(part, bytes.NewBuffer(presentResourcesJson))
	if err != nil {
		return
	}

	if zipFile != nil {
		zipStats, zipErr := zipFile.Stat()
		if zipErr != nil {
			return
		}

		if zipStats.Size() == 0 {
			return
		}

		part, zipErr = createZipPartWriter(zipStats, writer)
		if zipErr != nil {
			return
		}

		_, zipErr = io.Copy(part, zipFile)
		if zipErr != nil {
			return
		}
	}

	return
}

func createZipPartWriter(zipStats os.FileInfo, writer *multipart.Writer) (io.Writer, error) {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="application"; filename="application.zip"`)
	h.Set("Content-Type", "application/zip")
	h.Set("Content-Length", fmt.Sprintf("%d", zipStats.Size()))
	h.Set("Content-Transfer-Encoding", "binary")
	return writer.CreatePart(h)
}
