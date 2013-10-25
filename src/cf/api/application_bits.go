package api

import (
	"archive/zip"
	"bytes"
	"cf"
	"cf/configuration"
	"cf/net"
	"encoding/json"
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

func (repo CloudControllerApplicationBitsRepository) UploadApp(app cf.Application, dir string) (apiResponse net.ApiResponse) {
	dir, resourcesJson, apiResponse := repo.createUploadDir(app, dir)
	if apiResponse.IsNotSuccessful() {
		return
	}

	zipFile, err := repo.zipper.Zip(dir)
	if err != nil {
		apiResponse = net.NewApiResponseWithError("Error zipping application", err)
		return
	}
	defer zipFile.Close()

	apiResponse = repo.uploadBits(app, zipFile, resourcesJson)
	if apiResponse.IsNotSuccessful() {
		return
	}

	return
}

func (repo CloudControllerApplicationBitsRepository) uploadBits(app cf.Application, zipFile *os.File, resourcesJson []byte) (apiResponse net.ApiResponse) {
	url := fmt.Sprintf("%s/v2/apps/%s/bits?async=true", repo.config.Target, app.Guid)

	body, boundary, err := createApplicationUploadBody(zipFile, resourcesJson)
	if err != nil {
		apiResponse = net.NewApiResponseWithError("Error creating upload", err)
		return
	}

	request, apiResponse := repo.gateway.NewRequest("PUT", url, repo.config.AccessToken, body)
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

func (repo CloudControllerApplicationBitsRepository) createUploadDir(app cf.Application, appDir string) (uploadDir string, resourcesJson []byte, apiResponse net.ApiResponse) {
	var err error

	// If appDir is a zip, first extract it to a temporary directory
	if fileIsZip(appDir) {
		appDir, err = extractZip(app, appDir)
		if err != nil {
			apiResponse = net.NewApiResponseWithError("Error extracting archive", err)
			return
		}
	}

	// Find which files need to be uploaded
	allAppFiles, err := cf.AppFilesInDir(appDir)
	if err != nil {
		apiResponse = net.NewApiResponseWithError("Error listing app files", err)
		return
	}

	appFilesToUpload, resourcesJson, apiResponse := repo.getFilesToUpload(allAppFiles)
	if apiResponse.IsNotSuccessful() {
		return
	}

	// Copy files into a temporary directory and return it
	uploadDir, err = cf.TempDirForApp(app)
	if err != nil {
		apiResponse = net.NewApiResponseWithError("Error creating temporary directory", err)
		return
	}

	err = cf.InitializeDir(uploadDir)
	if err != nil {
		apiResponse = net.NewApiResponseWithError("Error creating upload directory", err)
		return
	}

	err = cf.CopyFiles(appFilesToUpload, appDir, uploadDir)
	if err != nil {
		apiResponse = net.NewApiResponseWithError("Error copying files to temp directory", err)
		return
	}

	return
}

func fileIsZip(file string) bool {
	isZip := strings.HasSuffix(file, ".zip")
	isWar := strings.HasSuffix(file, ".war")
	isJar := strings.HasSuffix(file, ".jar")

	return isZip || isWar || isJar
}

func extractZip(app cf.Application, zipFile string) (destDir string, err error) {
	destDir, err = cf.TempDirForApp(app)
	if err != nil {
		return
	}

	destDir = destDir + "-zip"
	err = cf.InitializeDir(destDir)
	if err != nil {
		return
	}

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

		var (
			rc       io.ReadCloser
			destFile *os.File
		)

		rc, err = f.Open()
		if err != nil {
			return
		}
		defer rc.Close()

		destFilePath := filepath.Join(destDir, f.Name)
		err = os.MkdirAll(filepath.Dir(destFilePath), os.ModePerm|os.ModeDir)
		if err != nil {
			return
		}

		destFile, err = os.Create(destFilePath)
		if err != nil {
			return
		}

		_, err = io.Copy(destFile, rc)
		if err != nil {
			return
		}

		rc.Close()
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
		appFilesToUpload = deleteAppFile(appFilesToUpload, appFile)
	}

	return
}

func deleteAppFile(appFiles []cf.AppFile, targetFile cf.AppFile) []cf.AppFile {
	for i, file := range appFiles {
		if file.Path == targetFile.Path {
			appFiles[i] = appFiles[len(appFiles)-1]
			return appFiles[:len(appFiles)-1]
		}
	}
	return appFiles
}

func createApplicationUploadBody(zipFile *os.File, resourcesJson []byte) (body *os.File, boundary string, err error) {
	tempFile, err := cf.TempFileForRequestBody()
	if err != nil {
		return
	}
	body, err = os.Create(tempFile)
	if err != nil {
		return
	}

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
