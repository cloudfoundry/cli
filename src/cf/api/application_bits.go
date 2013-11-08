package api

import (
	"archive/zip"
	"bytes"
	"cf"
	"cf/configuration"
	"cf/net"
	"encoding/json"
	"errors"
	"fileutils"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type AppFileResource struct {
	Path string `json:"fn"`
	Sha1 string `json:"sha1"`
	Size int64  `json:"size"`
}

type ApplicationBitsRepository interface {
	UploadApp(app cf.Application, dir string) (apiResponse net.ApiResponse)
}

type CloudControllerApplicationBitsRepository struct {
	config  *configuration.Configuration
	gateway net.Gateway
	zipper  cf.Zipper
}

func NewCloudControllerApplicationBitsRepository(config *configuration.Configuration, gateway net.Gateway, zipper cf.Zipper) (repo CloudControllerApplicationBitsRepository) {
	repo.config = config
	repo.gateway = gateway
	repo.zipper = zipper
	return
}

func (repo CloudControllerApplicationBitsRepository) UploadApp(app cf.Application, appDir string) (apiResponse net.ApiResponse) {
	fileutils.TempDir("apps", func(uploadDir string, err error) {
		if err != nil {
			apiResponse = net.NewApiResponseWithMessage(err.Error())
			return
		}

		var resourcesJson []byte
		repo.sourceDir(appDir, func(sourceDir string, sourceErr error) {
			if sourceErr != nil {
				err = sourceErr
				return
			}
			resourcesJson, err = repo.copyUploadableFiles(sourceDir, uploadDir)
		})

		if err != nil {
			apiResponse = net.NewApiResponseWithMessage(err.Error())
			return
		}

		fileutils.TempFile("uploads", func(zipFile *os.File, err error) {
			if err != nil {
				apiResponse = net.NewApiResponseWithMessage(err.Error())
				return
			}

			err = repo.zipper.Zip(uploadDir, zipFile)
			if err != nil {
				apiResponse = net.NewApiResponseWithError("Error zipping application", err)
				return
			}

			apiResponse = repo.uploadBits(app, zipFile, resourcesJson)
			if apiResponse.IsNotSuccessful() {
				return
			}
		})
	})
	return
}

func (repo CloudControllerApplicationBitsRepository) uploadBits(app cf.Application, zipFile *os.File, resourcesJson []byte) (apiResponse net.ApiResponse) {
	url := fmt.Sprintf("%s/v2/apps/%s/bits?async=true", repo.config.Target, app.Guid)

	fileutils.TempFile("requests", func(requestFile *os.File, err error) {
		if err != nil {
			apiResponse = net.NewApiResponseWithError("Error creating tmp file: %s", err)
			return
		}

		boundary, err := repo.writeUploadBody(zipFile, requestFile, resourcesJson)
		if err != nil {
			apiResponse = net.NewApiResponseWithError("Error writing to tmp file: %s", err)
			return
		}

		var request *net.Request
		request, apiResponse = repo.gateway.NewRequest("PUT", url, repo.config.AccessToken, requestFile)
		if apiResponse.IsNotSuccessful() {
			return
		}

		contentType := fmt.Sprintf("multipart/form-data; boundary=%s", boundary)
		request.HttpReq.Header.Set("Content-Type", contentType)

		response := &Resource{}
		_, apiResponse = repo.gateway.PerformRequestForJSONResponse(request, response)
		if apiResponse.IsNotSuccessful() {
			return
		}

		jobGuid := response.Metadata.Guid
		apiResponse = repo.pollUploadProgress(jobGuid)
	})

	return
}

const (
	uploadStatusFinished = "finished"
	uploadStatusFailed   = "failed"
)

type UploadProgressEntity struct {
	Status string
}

type UploadProgressResponse struct {
	Metadata Metadata
	Entity   UploadProgressEntity
}

func (repo CloudControllerApplicationBitsRepository) pollUploadProgress(jobGuid string) (apiResponse net.ApiResponse) {
	finished := false
	for !finished {
		finished, apiResponse = repo.uploadProgress(jobGuid)
		if apiResponse.IsNotSuccessful() {
			return
		}
		time.Sleep(time.Second)
	}
	return
}

func (repo CloudControllerApplicationBitsRepository) uploadProgress(jobGuid string) (finished bool, apiResponse net.ApiResponse) {
	url := fmt.Sprintf("%s/v2/jobs/%s", repo.config.Target, jobGuid)
	request, apiResponse := repo.gateway.NewRequest("GET", url, repo.config.AccessToken, nil)
	response := &UploadProgressResponse{}
	_, apiResponse = repo.gateway.PerformRequestForJSONResponse(request, response)

	switch response.Entity.Status {
	case uploadStatusFinished:
		finished = true
	case uploadStatusFailed:
		apiResponse = net.NewApiResponseWithMessage("Failed to complete upload.")
	}

	return
}

func (repo CloudControllerApplicationBitsRepository) sourceDir(appDir string, cb func(sourceDir string, err error)) {
	// If appDir is a zip, first extract it to a temporary directory
	if !repo.fileIsZip(appDir) {
		cb(appDir, nil)
		return
	}

	fileutils.TempDir("unzipped-app", func(tmpDir string, err error) {
		if err != nil {
			cb("", err)
			return
		}

		err = repo.extractZip(appDir, tmpDir)
		cb(tmpDir, err)
	})
}

func (repo CloudControllerApplicationBitsRepository) copyUploadableFiles(appDir string, uploadDir string) (resourcesJson []byte, err error) {
	// Find which files need to be uploaded
	allAppFiles, err := cf.AppFilesInDir(appDir)
	if err != nil {
		return
	}

	appFilesToUpload, resourcesJson, apiResponse := repo.getFilesToUpload(allAppFiles)
	if apiResponse.IsNotSuccessful() {
		err = errors.New(apiResponse.Message)
		return
	}

	// Copy files into a temporary directory and return it
	err = cf.CopyFiles(appFilesToUpload, appDir, uploadDir)
	if err != nil {
		return
	}

	return
}

func (repo CloudControllerApplicationBitsRepository) fileIsZip(file string) bool {
	isZip := strings.HasSuffix(file, ".zip")
	isWar := strings.HasSuffix(file, ".war")
	isJar := strings.HasSuffix(file, ".jar")

	return isZip || isWar || isJar
}

func (repo CloudControllerApplicationBitsRepository) extractZip(zipFile string, destDir string) (err error) {
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return
	}
	defer r.Close()

	for _, f := range r.File {
		// Don't try to extract directories
		if f.FileInfo().IsDir() {
			continue
		}

		if err != nil {
			return
		}

		var rc io.ReadCloser
		rc, err = f.Open()
		if err != nil {
			return
		}

		destFilePath := filepath.Join(destDir, f.Name)

		err = fileutils.CopyReaderToPath(rc, destFilePath)

		rc.Close()

		if err != nil {
			return
		}
	}

	return
}

func (repo CloudControllerApplicationBitsRepository) getFilesToUpload(allAppFiles []cf.AppFile) (appFilesToUpload []cf.AppFile, resourcesJson []byte, apiResponse net.ApiResponse) {
	appFilesRequest := []AppFileResource{}
	for _, file := range allAppFiles {
		appFilesRequest = append(appFilesRequest, AppFileResource{
			Path: file.Path,
			Sha1: file.Sha1,
			Size: file.Size,
		})
	}

	resourcesJson, err := json.Marshal(appFilesRequest)
	if err != nil {
		apiResponse = net.NewApiResponseWithError("Failed to create json for resource_match request", err)
		return
	}

	path := fmt.Sprintf("%s/v2/resource_match", repo.config.Target)
	req, apiResponse := repo.gateway.NewRequest("PUT", path, repo.config.AccessToken, bytes.NewReader(resourcesJson))
	if apiResponse.IsNotSuccessful() {
		return
	}

	res := []AppFileResource{}
	_, apiResponse = repo.gateway.PerformRequestForJSONResponse(req, &res)

	appFilesToUpload = make([]cf.AppFile, len(allAppFiles))
	copy(appFilesToUpload, allAppFiles)
	for _, file := range res {
		appFile := cf.AppFile{
			Path: file.Path,
			Sha1: file.Sha1,
			Size: file.Size,
		}
		appFilesToUpload = repo.deleteAppFile(appFilesToUpload, appFile)
	}

	return
}

func (repo CloudControllerApplicationBitsRepository) deleteAppFile(appFiles []cf.AppFile, targetFile cf.AppFile) []cf.AppFile {
	for i, file := range appFiles {
		if file.Path == targetFile.Path {
			appFiles[i] = appFiles[len(appFiles)-1]
			return appFiles[:len(appFiles)-1]
		}
	}
	return appFiles
}

func (repo CloudControllerApplicationBitsRepository) writeUploadBody(zipFile *os.File, body *os.File, resourcesJson []byte) (boundary string, err error) {
	writer := multipart.NewWriter(body)
	defer writer.Close()

	boundary = writer.Boundary()

	part, err := writer.CreateFormField("resources")
	if err != nil {
		return
	}

	_, err = io.Copy(part, bytes.NewBuffer(resourcesJson))
	if err != nil {
		return
	}

	zipStats, err := zipFile.Stat()
	if err != nil {
		return
	}

	if zipStats.Size() == 0 {
		return
	}

	part, err = createZipPartWriter(zipStats, writer)
	if err != nil {
		return
	}

	_, err = io.Copy(part, zipFile)
	if err != nil {
		return
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
